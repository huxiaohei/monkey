package network

import (
	"monkey/codec"
	"monkey/logger"
	"monkey/utils"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

var (
	_         ConnSession = &WebSocketConnSession{}
	wsclog, _             = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type WebSocketConnSession struct {
	createTime     time.Time
	lastActiveTime time.Time
	token          string
	sessionId      utils.SessionId
	conn           *websocket.Conn
	messageType    int
	codec          codec.Codec
	manager        SessionManager
	messageHandler MessageHandler
	messageSeqId   uint64
}

func NewWebSocketSession(conn *websocket.Conn, codec codec.Codec, messageHandler MessageHandler, sessionMgr SessionManager) *WebSocketConnSession {
	s := &WebSocketConnSession{
		createTime:     time.Now(),
		lastActiveTime: time.Now(),
		token:          "",
		sessionId:      sessionMgr.GenerateSessionId(),
		conn:           conn,
		messageType:    websocket.TextMessage,
		codec:          codec,
		manager:        sessionMgr,
		messageHandler: messageHandler,
		messageSeqId:   0,
	}
	sessionMgr.AddSession(s)
	return s
}

func (s *WebSocketConnSession) GetSessionId() utils.SessionId {
	return s.sessionId
}

func (s *WebSocketConnSession) GetMessageId() uint64 {
	return s.messageSeqId
}

func (s *WebSocketConnSession) ConnectType() NetWorkType {
	return NetWorkTypeWebSocket
}

func (s *WebSocketConnSession) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *WebSocketConnSession) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *WebSocketConnSession) ReceiveMessage(timeout int64) {
	defer s.manager.RemoveSession(s.sessionId)
	s.conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	s.conn.SetCloseHandler(func(code int, text string) error {
		wsclog.Info("receive close message from ", s.RemoteAddr(), " code: ", code, " text: ", text)
		s.messageHandler.ProcessClose(s, code)
		return nil
	})
	for {
		messageType, message, err := s.conn.ReadMessage()
		if err != nil {
			if netError, ok := err.(net.Error); ok && netError.Timeout() {
				wsclog.Info("receive message from ", s.RemoteAddr(), " timeout: ", err)
				if s.messageHandler.ProcessTimeout(s) {
					break
				}
				s.conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
				continue
			}
			break
		}
		s.lastActiveTime = time.Now()
		s.messageType = messageType
		s.conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		_, msg, err := s.codec.Decode(message)
		if err != nil {
			wsclog.Error("decode message from ", s.RemoteAddr(), " error: ", err)
			if s.messageSeqId == 0 {
				break
			}
			continue
		}
		s.messageSeqId++
		s.messageHandler.ProcessMessage(s, msg)
	}
}

func (s *WebSocketConnSession) SetManager(manager SessionManager) {
	s.manager = manager
}

func (s *WebSocketConnSession) SendMessage(message interface{}) error {
	s.lastActiveTime = time.Now()
	msg, err := s.codec.Encode(message)
	if err != nil {
		return err
	}
	err = s.conn.WriteMessage(s.messageType, msg)
	if err != nil {
		wsclog.Error("send message to ", s.RemoteAddr(), " error: ", err)
		return err
	}
	return err
}

func (s *WebSocketConnSession) SendMessageBatch(messages []interface{}) error {
	for _, message := range messages {
		err := s.SendMessage(message)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *WebSocketConnSession) Close() {
	s.conn.Close()
}
