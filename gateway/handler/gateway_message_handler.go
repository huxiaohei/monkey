package handler

import (
	"encoding/json"
	"monkey/gateway/protos"
	"monkey/logger"
	"monkey/network"
	"monkey/placement"
	"monkey/rpc"
	"monkey/rpc/pb"
)

var (
	_       network.MessageHandler = &GatewayMessageHandler{}
	mlog, _                        = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type GatewayMessageHandler struct {
	PD            placement.Placement
	ActorType     string
	Id            uint64
	ActorServerId uint64
	AccountToken  string
	TtlToken      string
}

func (gmh *GatewayMessageHandler) processFirstMessage(session network.ConnSession, msg interface{}) {
	// data, ok := msg.([]byte)
	data, ok := msg.(string)
	if !ok {
		mlog.Errorf("first message is not byte array %v", msg)
		resp := protos.NewSessionCloseResponse(0, "first message is invalid")
		session.SendMessage(resp)
		return
	}
	var firstPacket protos.FirstPacket
	err := json.Unmarshal([]byte(data), &firstPacket)
	if err != nil {
		mlog.Errorf("unmarshal first packet error %v", err)
		resp := protos.NewSessionCloseResponse(0, "first message is invalid")
		session.SendMessage(resp)
		return
	}

	position := gmh.PD.FindActorPositon(&placement.PlacementFindActorPositionArgs{
		ActorType: firstPacket.ServerType,
		Id:        firstPacket.UserId,
		TTL:       1800,
	})
	if position == nil {
		mlog.Errorf("can not find actor position %v, %v", firstPacket.ServerType, firstPacket.UserId)
		session.SendMessage(protos.NewSessionCloseResponse(0, "can not find actor position"))
		return
	}

	gmh.ActorType = firstPacket.ServerType
	gmh.Id = firstPacket.UserId
	gmh.ActorServerId = position.ServerId
	gmh.AccountToken = firstPacket.Token
	gmh.TtlToken = position.Token

	player, err := rpc.GetRPCClientManager().GetPlayerClient(gmh.Id)
	if err != nil {
		mlog.Errorf("get player rpc client error %v", err)
		return
	}
	bindMsg := &pb.BindMsg{
		Id:              gmh.Id,
		Ttl:             1800,
		TtlToken:        gmh.TtlToken,
		Weight:          1,
		SessionId:       uint64(session.GetSessionId()),
		SessionServerId: gmh.PD.GetCurServerId(),
		AccountToken:    gmh.AccountToken,
	}
	player.Bind(bindMsg)
	session.SendMessage(bindMsg)
}

func (gmh *GatewayMessageHandler) processCommonMessage(session network.ConnSession, msg interface{}) {
	pos := gmh.PD.FindActorPositon(&placement.PlacementFindActorPositionArgs{
		ActorType: gmh.ActorType,
		Id:        gmh.Id,
		TTL:       1800,
	})
	if pos == nil {
		mlog.Errorf("can not find actor position %s_%d", gmh.ActorType, gmh.Id)
		return
	}
	if pos.ServerId != gmh.ActorServerId {
		gmh.ActorServerId = pos.ServerId
	}

	player, err := rpc.GetRPCClientManager().GetPlayerClient(gmh.Id)
	if err != nil {
		mlog.Errorf("get player rpc client error %v", err)
		return
	}
	data, ok := msg.(string)
	if !ok {
		mlog.Errorf("message is not byte array %v", msg)
		return
	}
	player.ReceiveMessage(&pb.CommonMsg{
		ServerId:  gmh.PD.GetCurServerId(),
		SessionId: uint64(session.GetSessionId()),
		Payload:   []byte(data),
	})
}

func (gmh *GatewayMessageHandler) ProcessMessage(session network.ConnSession, msg interface{}) {
	if gmh.ActorServerId == 0 {
		gmh.processFirstMessage(session, msg)
	} else {
		gmh.processCommonMessage(session, msg)
	}
}

func (gmh *GatewayMessageHandler) ProcessTimeout(session network.ConnSession) bool {
	resp := protos.NewSessionCloseResponse(gmh.Id, "timeout")
	session.SendMessage(resp)
	return true
}

func (gmh *GatewayMessageHandler) ProcessClose(session network.ConnSession, code int) {
	mlog.Info("session ", session.GetMessageId(), " closed, code: ", code)
}
