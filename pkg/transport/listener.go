package transport

import (
	"context"

	"github.com/go-av/gosip/pkg/message"
)

type ListeningPoint interface {
	Listen(addr string, funcs ...ListenOptionFunc) error
	Start(ctx context.Context)
	SetTransportChannel(channel chan message.Message)
	Send(address string, msg message.Message) error
}

type Listener interface {
	HandleRequest(message.Request)
	HandleResponse(message.Response)
}

type ListenOptionFunc func(listen ListeningPoint)
