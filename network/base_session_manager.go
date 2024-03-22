package network

import (
	"fmt"
	"monkey/logger"
	"monkey/utils"
	"sync"
)

var (
	_        SessionManager = &BaseSessionManager{}
	cmlog, _                = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type BaseSessionManager struct {
	conns          map[utils.SessionId]ConnSession
	uniqueSequence *utils.UniqueSequence
	lock           sync.RWMutex
}

func NewSessionManager() *BaseSessionManager {
	sessionManager := &BaseSessionManager{
		conns:          make(map[utils.SessionId]ConnSession),
		uniqueSequence: utils.NewUniqueSequence(1000),
		lock:           sync.RWMutex{},
	}
	return sessionManager
}

func (cm *BaseSessionManager) GenerateSessionId() utils.SessionId {
	return cm.uniqueSequence.GetNewSessionId()
}

func (cm *BaseSessionManager) AddSession(conn ConnSession) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if _, ok := cm.conns[conn.GetSessionId()]; ok {
		cmlog.Error("connection already exists, sessionId", conn.GetSessionId(), conn.RemoteAddr().String(), conn.LocalAddr().String())
		return fmt.Errorf("connection already exists, sessionId: %d", conn.GetSessionId())
	}
	cm.conns[conn.GetSessionId()] = conn
	conn.SetManager(cm)
	return nil
}

func (cm *BaseSessionManager) RemoveSession(sessionId utils.SessionId) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if conn, ok := cm.conns[sessionId]; ok {
		conn.Close()
		delete(cm.conns, sessionId)
	}
}

func (cm *BaseSessionManager) GetSession(sessionId utils.SessionId) (ConnSession, bool) {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	conn, ok := cm.conns[sessionId]
	return conn, ok
}
