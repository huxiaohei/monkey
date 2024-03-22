package logger

import "sync"

var (
	instance *LoggerManager
	once     sync.Once
)

type LoggerTag string

const (
	MainTag LoggerTag = "main"
)

type LoggerManager struct {
	loggrs map[LoggerTag]*Logger
	lock   sync.RWMutex
}

func GetLoggerManager() *LoggerManager {
	once.Do(func() {
		instance = &LoggerManager{
			loggrs: make(map[LoggerTag]*Logger),
		}
	})
	return instance
}

func (lm *LoggerManager) GetLogger(tag LoggerTag) (*Logger, bool) {
	lm.lock.RLock()
	defer lm.lock.RUnlock()

	if logger, ok := lm.loggrs[tag]; ok {
		return logger, ok
	}
	logger, e := NewLogger("./", string(tag))
	if e != nil {
		return nil, false
	}
	lm.loggrs[tag] = logger
	return logger, true
}

func (lm *LoggerManager) AddLogger(tag LoggerTag, logger *Logger) bool {
	lm.lock.Lock()
	defer lm.lock.Unlock()

	if _, ok := lm.loggrs[tag]; ok {
		return false
	}
	lm.loggrs[tag] = logger
	return true
}
