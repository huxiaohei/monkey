package network

type MessageHandler interface {
	ProcessMessage(session ConnSession, msg interface{})
	ProcessTimeout(session ConnSession) bool
	ProcessClose(session ConnSession, code int)
}
