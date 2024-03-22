package network

import (
	"monkey/utils"
	"net"
)

const SessionMaxCacheSize = 10

type NetWorkType int

const (
	NetWorkTypeNone NetWorkType = iota
	NetWorkTypeWebSocket
)

type SocketFrameType int

const (
	SocketFrameTypeText SocketFrameType = iota
	SocketFrameTypeBinary
)

type ConnSession interface {
	GetSessionId() utils.SessionId
	GetMessageId() uint64
	ConnectType() NetWorkType
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetManager(manager SessionManager)
	ReceiveMessage(timeout int64)
	SendMessage(message interface{}) error
	SendMessageBatch(messages []interface{}) error
	Close()
}
