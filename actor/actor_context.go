package actor

import (
	"monkey/utils"
)

type ActorContext struct {
	mailbox         chan []byte
	processHandlers map[string]func(msg []byte) []byte
	lastMsgTime     int64
}

func NewActorContext() *ActorContext {
	return &ActorContext{
		mailbox:     make(chan []byte, 128),
		lastMsgTime: utils.GetNowMs(),
	}
}

func (ac *ActorContext) RegisterMessageProcess(name string, process func(msg []byte) []byte) {
	ac.processHandlers[name] = process
}

func (ac *ActorContext) PopMessage() []byte {
	msg := <-ac.mailbox
	return msg
}

func (ac *ActorContext) PushMessage(msg []byte) {
	ac.mailbox <- msg
	ac.lastMsgTime = utils.GetNowMs()
}
