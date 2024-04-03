package actor

import (
	"fmt"
	"monkey/utils"
)

type ActorId struct {
	ActorType string `json:"actorType" description:"Actor类型"`
	Id        uint64 `json:"actorId" description:"ActorID"`
}

func (ai ActorId) String() string {
	return fmt.Sprintf("actorType: %s, actorId: %d", ai.ActorType, ai.Id)
}

type ActorInfo struct {
	ActorId         ActorId         `json:"actorId" description:"ActorID"`
	ServerId        uint64          `json:"serverId" description:"服务器ID"`
	SessionId       utils.SessionId `json:"sessionId" description:"sessionID"`
	SessionServerId uint64          `json:"sessionServerId" description:"session服务器ID"`
	AccountToken    string          `json:"accountToken" description:"账号Token"`
	TtlToken        string          `json:"ttlToken" description:"TTL Token"`
}

func (ai ActorInfo) String() string {
	return fmt.Sprintf("actorId: %s, serverId: %d, sessionId: %d, sessionServerId: %d, accountToken: %s, ttlToken: %s", ai.ActorId, ai.ServerId, ai.SessionId, ai.SessionServerId, ai.AccountToken, ai.TtlToken)
}
