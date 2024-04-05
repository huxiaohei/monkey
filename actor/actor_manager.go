package actor

import (
	"fmt"
	"monkey/placement"
	"sync"
)

var (
	instance     *ActorManager
	once         sync.Once
	timerManager = NewActorTimerManager()
)

type ActorManager struct {
	pd     placement.Placement
	actors map[ActorId]Actor
	lock   *sync.RWMutex
	impls  map[string]func(actorType string, id uint64, timerManager *ActorTimerManager, pd placement.Placement, localServerId uint64) Actor
}

func GetActorManager() *ActorManager {
	once.Do(func() {
		instance = &ActorManager{
			pd:     nil,
			actors: make(map[ActorId]Actor),
			lock:   &sync.RWMutex{},
			impls:  make(map[string]func(actorType string, id uint64, timerManager *ActorTimerManager, pd placement.Placement, localServerId uint64) Actor),
		}
	})
	return instance
}

func (am *ActorManager) SetPlacement(pd placement.Placement) {
	am.pd = pd
}

func (am *ActorManager) RegisterActorImpl(actorType string, impl func(actorType string, id uint64, timerManager *ActorTimerManager, pd placement.Placement, localServerId uint64) Actor) {
	am.impls[actorType] = impl
}

func (am *ActorManager) AddOrUpdateActor(actor Actor) {
	am.lock.Lock()
	defer am.lock.Unlock()

	if _, ok := am.actors[actor.ActorId()]; ok {
		am.actors[actor.ActorId()].OnDeactivate()
		delete(am.actors, actor.ActorId())
	}
	am.actors[actor.ActorId()] = actor
}

func (am *ActorManager) GetOrNew(actorId ActorId) (Actor, error) {
	am.lock.Lock()
	defer am.lock.Unlock()

	if actor, ok := am.actors[actorId]; ok {
		if actor.IsActivated() {
			return actor, nil
		}
		am.actors[actorId].OnActivate()
		delete(am.actors, actorId)
	}

	if impl, ok := am.impls[actorId.ActorType]; ok {
		actor := impl(actorId.ActorType, actorId.Id, timerManager, am.pd, 0)
		am.actors[actorId] = actor
		am.actors[actorId].OnActivate()
		return actor, nil
	}

	return nil, fmt.Errorf("actor type %s not found", actorId.ActorType)
}
