package transport

import (
	"github.com/go-av/gosip/pkg/message"
)

type ListeningPoint interface {
	Build(addr string) error
	Start()
	SetTransportChannel(channel chan message.Message)
	Send(address string, msg message.Message) error
}

type Listener interface {
	HandleRequest(message.Request)
	HandleResponse(message.Response)
}
