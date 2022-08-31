package sip

import (
	"context"
	"errors"
	"fmt"

	_ "github.com/go-av/gosip/pkg/log"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/transport"
)

func NewSipStack(name string) *SipStack {
	stack := new(SipStack)
	stack.name = name
	stack.transportChannel = make(chan message.Message, 100)
	return stack
}

type SipStack struct {
	ctx  context.Context
	name string

	ListeningPoints  []transport.ListeningPoint
	transportChannel chan message.Message
	listener         transport.Listener

	funcMap map[string]func(message.Message)
}

func (stack *SipStack) CreateListenPoint(protocol string, host string, port int) transport.ListeningPoint {
	listenpoint := transport.NewTransportListenPoint(protocol, host, port)
	listenpoint.SetTransportChannel(stack.transportChannel)
	stack.ListeningPoints = append(stack.ListeningPoints, listenpoint)
	return listenpoint
}

func (stack *SipStack) SetListener(listener transport.Listener) {
	stack.listener = listener
}

func (stack *SipStack) SetFuncHandler(funcName string, handler func(message.Message)) {
	stack.funcMap[funcName] = handler
}

func (stack *SipStack) Start(ctx context.Context) {
	defer fmt.Println("SipStack id close")
	for _, listeningPoint := range stack.ListeningPoints {
		go listeningPoint.Start()
	}
	stack.ctx = ctx
	for {
		select {
		case <-stack.ctx.Done():
			return
		case msg := <-stack.transportChannel:
			if stack.listener != nil {
				if resp, ok := msg.(message.Response); ok {
					go stack.listener.HandleResponses(resp)
					continue
				}
				if req, ok := msg.(message.Request); ok {
					go stack.listener.HandleRequests(req)
					continue
				}
			}
		}
	}
}

func (stack *SipStack) Send(address *message.Address, msg message.Message) error {
	if len(stack.ListeningPoints) > 0 {
		return stack.ListeningPoints[0].Send(address.Host, address.Port.String(), msg)
	}
	return errors.New("not found Listening Point")
}
