package storage

import (
	"fmt"
	"monkey/placement"
)

func GetStorageKey(actorType string, id uint64) string {
	return fmt.Sprintf("%s_%d", actorType, id)
}

type ActorStorage interface {
	Name() string
	ClearActors()
	GetActorInfo(actorType string, id uint64) (*placement.PlacementActorPosition, bool, error)
	PutActorInfo(info *placement.PlacementActorPosition) error
	DeleteActor(actorType string, id uint64) error
}

type SequenceStorage interface {
	Name() string
	NewSequence(sequenceType string, step uint64) (*placement.SequenceResponse, error)
}

type ServerStorage interface {
	Name() string
	RegisterServer(info *placement.PlacementHostInfo) error
	DeleteServer(serverId uint64) error
	KeepAliveServer(serverId uint64, leaseId uint64, load uint64) (*placement.PlacementHostInfo, error)
	GetServerInfo(serverId uint64) (*placement.PlacementHostInfo, error)
	GetAllServerInfo() ([]*placement.PlacementHostInfo, error)
	IsServerValid(serverId uint64) bool
}
