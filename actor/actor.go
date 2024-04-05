package actor

import (
	"fmt"
	"monkey/utils"
)

type ActorId struct {
	ActorType string `json:"actorType" description:"Actor类型"`
	Id        uint64 `json:"actorId" description:"Id"`
}

func (ai ActorId) String() string {
	return fmt.Sprintf("actorType: %s, Id: %d", ai.ActorType, ai.Id)
}

type Actor interface {
	OnActivate()
	Init(ttl int64, weight uint64, ttlToken string, sessionId utils.SessionId, sessionServerId uint64)
	OnDeactivate()

	ActorId() ActorId
	TTL() int64
	ActerWeight() uint64
	IsActivated() bool

	RegisterTimer(interval int64, leftCount uint64, callback func(...interface{}), args ...interface{}) utils.TimerId
	UnregisterTimer(id utils.TimerId)
	UnregisterAllTimer()

	SendMessageToClient(msg []byte)
	RegisterMessageProcess(name string, process func(msg []byte) []byte)

	DispatchMessage()
}
