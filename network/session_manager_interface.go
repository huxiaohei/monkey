package network

import "monkey/utils"

type SessionManager interface {
	GenerateSessionId() utils.SessionId
	AddSession(conn ConnSession) error
	RemoveSession(sessionId utils.SessionId)
	GetSession(sessionId utils.SessionId) (ConnSession, bool)
}
