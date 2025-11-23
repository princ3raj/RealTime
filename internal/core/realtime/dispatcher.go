package realtime

import (
	"RealTime/internal/logger"

	"go.uber.org/zap"
)

type Dispatcher struct {
	handlers map[string]MessageHandler
}

func NewDispatcher() *Dispatcher {
	d := &Dispatcher{
		handlers: make(map[string]MessageHandler),
	}

	d.handlers["chat"] = ChatHandler{}
	d.handlers["ping"] = PingHandler{}
	d.handlers["private"] = PrivateHandler{}
	d.handlers["join"] = ChatHandler{}
	d.handlers["leave"] = ChatHandler{}
	d.handlers["market_news"] = NewsHandler{}

	return d

}

func (d *Dispatcher) Dispatch(hub *Hub, msg *Message) {
	if handler, ok := d.handlers[msg.Type]; ok {
		handler.Handle(hub, msg)
	} else {
		logger.Logger.Warn("Unknown message type received.", zap.String("type", msg.Type))
	}
}
