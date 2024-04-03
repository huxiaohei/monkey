package handler

import (
	"encoding/json"
	"monkey/actor"
	"monkey/gateway/protos"
	"monkey/logger"
	"monkey/network"
	"monkey/placement"
	"monkey/rpc"
	"monkey/rpc/pb"
)

var (
	_         network.MessageHandler = &GatewayMessageHandler{}
	gmhlog, _                        = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type GatewayMessageHandler struct {
	PD        placement.Placement
	ActorInfo actor.ActorInfo
}

func (gmh *GatewayMessageHandler) processFirstMessage(session network.ConnSession, msg interface{}) {
	// data, ok := msg.([]byte)
	data, ok := msg.(string)
	if !ok {
		gmhlog.Errorf("first message is not byte array %v", msg)
		resp := protos.NewSessionCloseResponse(0, "first message is invalid")
		session.SendMessage(resp)
		return
	}
	var firstPacket protos.FirstPacket
	err := json.Unmarshal([]byte(data), &firstPacket)
	if err != nil {
		gmhlog.Errorf("unmarshal first packet error %v", err)
		resp := protos.NewSessionCloseResponse(0, "first message is invalid")
		session.SendMessage(resp)
		return
	}

	position := gmh.PD.FindActorPositon(&placement.PlacementFindActorPositionArgs{
		ActorId: actor.ActorId{
			ActorType: firstPacket.ServerType,
			Id:        firstPacket.UserId,
		},
		TTL: 1800,
	})
	if position == nil {
		gmhlog.Errorf("can not find actor position %v, %v", firstPacket.ServerType, firstPacket.UserId)
		session.SendMessage(protos.NewSessionCloseResponse(0, "can not find actor position"))
		return
	}

	gmh.ActorInfo.ActorId.ActorType = firstPacket.ServerType
	gmh.ActorInfo.ActorId.Id = firstPacket.UserId
	gmh.ActorInfo.ServerId = position.ServerId
	gmh.ActorInfo.SessionId = session.GetSessionId()
	gmh.ActorInfo.SessionServerId = gmh.PD.GetCurServerId()
	gmh.ActorInfo.AccountToken = firstPacket.Token
	gmh.ActorInfo.TtlToken = position.Token

	player, err := rpc.GetRPCClientManager().GetPlayerClient(gmh.ActorInfo.ActorId.Id)
	if err != nil {
		gmhlog.Errorf("get player rpc client error %v", err)
		return
	}
	bindMsg := &pb.BindMsg{
		ServerId:     gmh.PD.GetCurServerId(),
		UserId:       gmh.ActorInfo.ActorId.Id,
		SessionId:    uint64(session.GetSessionId()),
		MsgSeq:       firstPacket.MsgSeq,
		AccountToken: gmh.ActorInfo.AccountToken,
		TtlToken:     gmh.ActorInfo.TtlToken,
	}
	player.Bind(bindMsg)
}

func (gmh *GatewayMessageHandler) processCommonMessage(session network.ConnSession, msg interface{}) {
	pos := gmh.PD.FindActorPositon(&placement.PlacementFindActorPositionArgs{ActorId: gmh.ActorInfo.ActorId, TTL: 1800})
	if pos == nil {
		gmhlog.Errorf("can not find actor position %v", gmh.ActorInfo.ActorId)
		return
	}
	if pos.ServerId != gmh.ActorInfo.ServerId {
		gmh.ActorInfo.ServerId = pos.ServerId
	}

	player, err := rpc.GetRPCClientManager().GetPlayerClient(gmh.ActorInfo.ActorId.Id)
	if err != nil {
		gmhlog.Errorf("get player rpc client error %v", err)
		return
	}
	data, ok := msg.(string)
	if !ok {
		gmhlog.Errorf("message is not byte array %v", msg)
		return
	}
	player.ReceiveMessage(&pb.CommonMsg{
		ServerId:  gmh.PD.GetCurServerId(),
		SessionId: uint64(session.GetSessionId()),
		Payload:   []byte(data),
	})
}

func (gmh *GatewayMessageHandler) ProcessMessage(session network.ConnSession, msg interface{}) {
	if gmh.ActorInfo.ServerId == 0 {
		gmh.processFirstMessage(session, msg)
	} else {
		gmh.processCommonMessage(session, msg)
	}
}

func (gmh *GatewayMessageHandler) ProcessTimeout(session network.ConnSession) bool {
	resp := protos.NewSessionCloseResponse(gmh.ActorInfo.ActorId.Id, "timeout")
	session.SendMessage(resp)
	return true
}

func (gmh *GatewayMessageHandler) ProcessClose(session network.ConnSession, code int) {
	gmhlog.Info("session ", session.GetMessageId(), " closed, code: ", code)
}
