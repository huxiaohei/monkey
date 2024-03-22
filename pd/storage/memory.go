package storage

import (
	"fmt"
	"monkey/actor"
	"monkey/logger"
	"monkey/placement"
	"monkey/utils"
	"sync"
)

var (
	_       ActorStorage    = &MemoryStorage{}
	_       SequenceStorage = &MemoryStorage{}
	_       ServerStorage   = &MemoryStorage{}
	mlog, _                 = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

const (
	minSequenceId = uint64(100)
)

type MemoryStorage struct {
	sequences   map[string]*utils.UniqueSequence
	mulex       sync.RWMutex
	actorCache  map[string]*placement.PlacementActorPosition
	serverCache map[uint64]*placement.PlacementActorHostInfo
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		sequences:   make(map[string]*utils.UniqueSequence),
		mulex:       sync.RWMutex{},
		actorCache:  make(map[string]*placement.PlacementActorPosition),
		serverCache: make(map[uint64]*placement.PlacementActorHostInfo),
	}
}

func (s *MemoryStorage) Name() string {
	return "memory"
}

func (s *MemoryStorage) ClearActors() {
	defer s.mulex.Unlock()
	s.mulex.Lock()

	mlog.Info("MemoryStorage closed")
	s.actorCache = make(map[string]*placement.PlacementActorPosition)
}

func (s *MemoryStorage) GetActorInfo(actorId *actor.ActorId) (*placement.PlacementActorPosition, bool, error) {
	defer s.mulex.RUnlock()
	s.mulex.RLock()

	key := ActorIdToStorageKey(actorId)
	if record, ok := s.actorCache[key]; ok {
		return record, true, nil
	}
	return nil, false, nil
}

func (s *MemoryStorage) PutActorInfo(info *placement.PlacementActorPosition) error {
	defer s.mulex.Unlock()
	s.mulex.Lock()

	key := ActorIdToStorageKey(&info.ActorId)
	s.actorCache[key] = info
	return nil
}

func (s *MemoryStorage) DeleteActor(actorId *actor.ActorId) error {
	defer s.mulex.Unlock()
	s.mulex.Lock()

	key := ActorIdToStorageKey(actorId)
	delete(s.actorCache, key)
	return nil
}

func (s *MemoryStorage) NewSequence(sequenceType string, step uint64) (*placement.SequenceResponse, error) {
	defer s.mulex.Unlock()
	s.mulex.Lock()

	if _, ok := s.sequences[sequenceType]; !ok {
		s.sequences[sequenceType] = utils.NewUniqueSequence(minSequenceId)
	}
	seq := s.sequences[sequenceType]
	resp := &placement.SequenceResponse{
		Id: seq.GetNewSequence(step),
	}
	return resp, nil
}

func (s *MemoryStorage) RegisterServer(info *placement.PlacementActorHostInfo) error {
	defer s.mulex.Unlock()
	s.mulex.Lock()

	if _, ok := s.serverCache[info.ServerId]; ok {
		return fmt.Errorf("server already exists %v", info.ServerId)
	}
	if info.LeaseId == 0 {
		info.LeaseId = info.ServerId
	}
	info.StartTime = utils.GetNowSec()
	info.DeadTime = utils.GetNowSec() + info.TTL
	s.serverCache[info.ServerId] = info
	return nil
}

func (s *MemoryStorage) DeleteServer(serverId uint64) error {
	defer s.mulex.Unlock()
	s.mulex.Lock()

	if _, ok := s.serverCache[serverId]; !ok {
		return fmt.Errorf("server not found %v", serverId)
	}
	delete(s.serverCache, serverId)
	return nil
}

func (s *MemoryStorage) KeepAliveServer(serverId uint64, leaseId uint64, load uint64) (*placement.PlacementActorHostInfo, error) {
	defer s.mulex.Unlock()
	s.mulex.Lock()

	if _, ok := s.serverCache[serverId]; !ok {
		return nil, fmt.Errorf("server not found %v", serverId)
	}
	info := s.serverCache[serverId]
	if info.LeaseId != leaseId {
		return nil, fmt.Errorf("lease id not match %v", leaseId)
	}
	info.Load = load
	info.DeadTime = utils.GetNowSec() + info.TTL
	return info, nil
}

func (s *MemoryStorage) GetServerInfo(serverId uint64) (*placement.PlacementActorHostInfo, error) {
	defer s.mulex.RUnlock()
	s.mulex.RLock()

	if info, ok := s.serverCache[serverId]; ok {
		return info, nil
	}
	return nil, fmt.Errorf("server not found %v", serverId)
}

func (s *MemoryStorage) GetAllServerInfo() ([]*placement.PlacementActorHostInfo, error) {
	defer s.mulex.RUnlock()
	s.mulex.RLock()

	var infos []*placement.PlacementActorHostInfo
	for _, info := range s.serverCache {
		if info.DeadTime <= utils.GetNowSec() {
			continue
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (s *MemoryStorage) IsServerValid(serverId uint64) bool {
	defer s.mulex.RUnlock()
	s.mulex.RLock()

	if info, ok := s.serverCache[serverId]; ok {
		if info.DeadTime > utils.GetNowSec() {
			return true
		}
	}
	return false
}
