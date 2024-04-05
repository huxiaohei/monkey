package player

import (
	"monkey/actor"
	"monkey/logger"
	"monkey/placement"
	"monkey/rpc/pb"
	"monkey/utils"
)

var (
	mlog, _ = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type Player struct {
	actor.BaseActor
}

func NewPlayer(actorType string, id uint64, timerManager *actor.ActorTimerManager, pd placement.Placement, localServerId uint64) actor.Actor {
	p := &Player{
		BaseActor: *actor.NewBaseActor(actorType, id, timerManager, pd, localServerId),
	}
	return p
}

func (ba *Player) Bind(msg *pb.BindMsg) {
	ba.BaseActor.Init(msg.Ttl, msg.Weight, msg.TtlToken, utils.SessionId(msg.SessionId), msg.SessionServerId)
}
