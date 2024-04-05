package actor

import (
	"monkey/logger"
	"monkey/utils"
	"sync"
)

var (
	timerSeq = utils.NewUniqueSequence(0)
	mlog, _  = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type ActorTimer struct {
	timerId   utils.TimerId
	begin     int64
	interval  int64
	passCount uint64
	leftCount uint64
	isCancel  bool
	manager   *ActorTimerManager
	callback  func(...interface{})
	args      []interface{}
}

func newActorTimer(interval int64, leftCount uint64, callback func(...interface{}), manager *ActorTimerManager, args ...interface{}) *ActorTimer {
	if interval < 0 {
		interval = 0
	}
	return &ActorTimer{
		timerId:   timerSeq.GetNewTimerId(),
		begin:     utils.GetNowMs(),
		interval:  interval,
		passCount: 0,
		leftCount: leftCount,
		isCancel:  false,
		manager:   manager,
		callback:  callback,
		args:      args,
	}
}

func (at *ActorTimer) GetTimerId() utils.TimerId {
	return at.timerId
}

func (at *ActorTimer) GetBeginTs() int64 {
	return at.begin
}

func (at *ActorTimer) GetIntervalMs() int64 {
	return at.interval
}

func (at *ActorTimer) GetPassCount() uint64 {
	return at.passCount
}

func (at *ActorTimer) GetLeftCount() uint64 {
	return at.leftCount
}

func (at *ActorTimer) IsCancel() bool {
	return at.isCancel
}

func (at *ActorTimer) Cancel() {
	at.isCancel = true
}

func (at *ActorTimer) Tick() int64 {
	defer func() {
		if err := recover(); err != nil {
			mlog.Errorf("ActorTimer Tick panic:%s, timerId: %d passCount:%d leftCount:%d", err, at.timerId, at.passCount, at.leftCount)
			at.isCancel = true
		}
	}()
	if at.isCancel {
		return -1
	}
	at.callback(at.args...)
	at.passCount++
	at.leftCount--
	if at.leftCount == 0 {
		at.isCancel = true
	}
	if !at.isCancel {
		return utils.GetNowSec() + at.interval
	}
	return -1
}

func (at *ActorTimer) run() {
	if at.passCount != 0 {
		return
	}
	for !at.IsCancel() {
		nextTs := at.Tick()
		if nextTs <= 0 {
			at.isCancel = true
			break
		}
		if nextTs > 0 {
			utils.SleepSec(nextTs - utils.GetNowSec())
		}
	}
	at.manager.UnregisterTimer(at.timerId)
}

type ActorTimerManager struct {
	timers map[utils.TimerId]*ActorTimer
	mutex  *sync.Mutex
}

func NewActorTimerManager() *ActorTimerManager {
	return &ActorTimerManager{
		timers: make(map[utils.TimerId]*ActorTimer),
		mutex:  &sync.Mutex{},
	}
}

func (atm *ActorTimerManager) RegisterTimer(interval int64, leftCount uint64, callback func(...interface{}), args ...interface{}) utils.TimerId {
	defer atm.mutex.Unlock()
	atm.mutex.Lock()

	timer := newActorTimer(interval, leftCount, callback, atm, args...)
	atm.timers[timer.GetTimerId()] = timer
	go timer.run()
	return timer.GetTimerId()
}

func (atm *ActorTimerManager) UnregisterTimer(timerId utils.TimerId) {
	defer atm.mutex.Unlock()
	atm.mutex.Lock()

	if timer, ok := atm.timers[timerId]; ok {
		timer.Cancel()
		delete(atm.timers, timerId)
	}
}

func (atm *ActorTimerManager) UnregisterAll() {
	defer atm.mutex.Unlock()
	atm.mutex.Lock()

	for _, timer := range atm.timers {
		timer.Cancel()
	}
	atm.timers = make(map[utils.TimerId]*ActorTimer)
}

func (atm *ActorTimerManager) GetTimer(timerId utils.TimerId) *ActorTimer {
	defer atm.mutex.Unlock()
	atm.mutex.Lock()

	if timer, ok := atm.timers[timerId]; ok {
		return timer
	}
	return nil
}
