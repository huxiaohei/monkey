package actor

import (
	"encoding/binary"
	"math"
	"monkey/placement"
	"monkey/rpc"
	"monkey/rpc/pb"
	"monkey/utils"
)

var (
	_ Actor = &BaseActor{}
)

type BaseActor struct {
	actorId          ActorId
	ttl              int64
	ttlToken         string
	weight           uint64
	sessionId        utils.SessionId
	sessionServerId  uint64
	localServerId    uint64
	keepAliveTimerId utils.TimerId
	timerIds         map[utils.TimerId]bool
	context          *ActorContext
	timerManager     *ActorTimerManager
	pd               placement.Placement
}

func NewBaseActor(actorType string, id uint64, timerManager *ActorTimerManager, pd placement.Placement, localServerId uint64) *BaseActor {
	return &BaseActor{
		actorId: ActorId{
			ActorType: actorType,
			Id:        id,
		},
		ttl:              -1,
		ttlToken:         "",
		weight:           1,
		sessionId:        0,
		sessionServerId:  0,
		localServerId:    localServerId,
		keepAliveTimerId: 0,
		timerIds:         make(map[utils.TimerId]bool),
		context:          NewActorContext(),
		timerManager:     timerManager,
		pd:               pd,
	}
}

func (ba *BaseActor) OnActivate() {
	go ba.DispatchMessage()
}

func (ba *BaseActor) Init(ttl int64, weight uint64, ttlToken string, sessionId utils.SessionId, sessionServerId uint64) {
	if ba.keepAliveTimerId > 0 {
		ba.timerManager.UnregisterTimer(ba.keepAliveTimerId)
	}
	ba.ttl = ttl
	ba.ttlToken = ttlToken
	ba.weight = weight
	ba.sessionId = sessionId
	ba.sessionServerId = sessionServerId
	ba.keepAliveTimerId = ba.RegisterTimer(ba.ttl/3, math.MaxUint64, ba.KeepAlive)
	ba.timerIds[ba.keepAliveTimerId] = true
}

func (ba *BaseActor) OnDeactivate() {
	ba.UnregisterAllTimer()
	ba.UnregisterTimer(ba.keepAliveTimerId)
}

func (ba *BaseActor) ActorId() ActorId {
	return ba.actorId
}

func (ba *BaseActor) TTL() int64 {
	return ba.ttl

}
func (ba *BaseActor) ActerWeight() uint64 {
	return ba.weight
}

func (ba *BaseActor) IsActivated() bool {
	return ba.context.lastMsgTime+ba.ttl > utils.GetNowMs()
}

func (ba *BaseActor) RegisterTimer(interval int64, leftCount uint64, callback func(...interface{}), args ...interface{}) utils.TimerId {
	timerId := ba.timerManager.RegisterTimer(interval, leftCount, callback, args...)
	ba.timerIds[timerId] = false
	return timerId
}

func (ba *BaseActor) UnregisterTimer(id utils.TimerId) {
	ba.timerManager.UnregisterTimer(id)
	delete(ba.timerIds, id)
}

func (ba *BaseActor) UnregisterAllTimer() {
	for id := range ba.timerIds {
		if id == ba.keepAliveTimerId {
			continue
		}
		ba.UnregisterTimer(id)
	}
}

func (ba *BaseActor) SendMessageToClient(msg []byte) {
	if ba.sessionId == 0 {
		mlog.Infof("ActiveId: %s SessionId:0, can't send message", ba.actorId)
		return
	}
	client, err := rpc.GetRPCClientManager().GetGatewayClient(ba.sessionServerId)
	if err != nil {
		mlog.Errorf("get gateway client error: %v", err)
		return
	}
	client.SendMessage(&pb.ClientMsg{
		SessionId: uint64(ba.sessionId),
		Payload:   msg,
	})
}

func (ba *BaseActor) RegisterMessageProcess(name string, process func(msg []byte) []byte) {
	ba.context.RegisterMessageProcess(name, process)
}

func (ba *BaseActor) DispatchMessage() {
	for msg := ba.context.PopMessage(); msg != nil; msg = ba.context.PopMessage() {
		nameLen := int16(binary.LittleEndian.Uint32(msg[:2]))
		name := string(msg[2 : 2+nameLen])
		process, ok := ba.context.processHandlers[name]
		if !ok {
			mlog.Errorf("process handler not found: %s", name)
			continue
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					mlog.Errorf("panic when process message: %v", r)
				}
			}()
			resp := process(msg[2+nameLen:])
			if resp != nil {
				ba.SendMessageToClient(resp)
			}
		}()
	}
}

func (ba *BaseActor) KeepAlive(args ...interface{}) {
	if ba.pd == nil {
		mlog.Errorf("keep alive actor failed, pd is nil, actorId: %s", ba.actorId)
		return
	}
	resp := ba.pd.ActorKeepAliveActor(ba.actorId.ActorType, ba.actorId.Id, ba.ttlToken)
	if resp == nil {
		mlog.Errorf("keep alive actor failed, actorId: %s", ba.actorId)
		return
	}
	mlog.Infof("keep alive actor success, actorId: %s resp:%s", ba.actorId, resp)
}
