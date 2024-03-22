package storage

import (
	"fmt"
	"monkey/actor"
	"monkey/placement"
)

func ActorIdToStorageKey(actorId *actor.ActorId) string {
	return fmt.Sprintf("%s_%d", actorId.ActorType, actorId.Id)
}

type ActorStorage interface {
	Name() string
	ClearActors()
	GetActorInfo(actorId *actor.ActorId) (*placement.PlacementActorPosition, bool, error)
	PutActorInfo(info *placement.PlacementActorPosition) error
	DeleteActor(actorId *actor.ActorId) error
}

type SequenceStorage interface {
	Name() string
	NewSequence(sequenceType string, step uint64) (*placement.SequenceResponse, error)
}

type ServerStorage interface {
	Name() string
	RegisterServer(info *placement.PlacementActorHostInfo) error
	DeleteServer(serverId uint64) error
	KeepAliveServer(serverId uint64, leaseId uint64, load uint64) (*placement.PlacementActorHostInfo, error)
	GetServerInfo(serverId uint64) (*placement.PlacementActorHostInfo, error)
	GetAllServerInfo() ([]*placement.PlacementActorHostInfo, error)
	IsServerValid(serverId uint64) bool
}
