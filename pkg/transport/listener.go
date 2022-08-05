package transport

import "github.com/go-av/gosip/pkg/message"

type ListeningPoint interface {
	Build(host string, port int) error
	Start()
	SetTransportChannel(channel chan message.Message)
	Send(host string, port string, msg message.Message) error
}

type Listener interface {
	HandleRequests(message.Request)
	HandleResponses(message.Response)
}
