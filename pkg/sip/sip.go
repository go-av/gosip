package sip

import (
	"context"
	"fmt"
	"strings"
	"sync"

	_ "github.com/go-av/gosip/pkg/log"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/transport"
)

func NewSipStack(name string) *SipStack {
	stack := new(SipStack)
	stack.name = name
	stack.listeningPoints = &sync.Map{}
	stack.transportChannel = make(chan message.Message, 100)
	return stack
}

type SipStack struct {
	ctx  context.Context
	name string

	listeningPoints  *sync.Map
	transportChannel chan message.Message
	listener         transport.Listener

	funcMap map[method.Method]func(message.Message)
}

func (stack *SipStack) CreateListenPoint(protocol string, addr string) (transport.ListeningPoint, error) {
	protocol = strings.ToLower(protocol)
	if _, ok := stack.listeningPoints.Load(protocol); ok {
		return nil, fmt.Errorf("%s listen point is exist", protocol)
	}
	listenpoint, err := transport.NewTransportListenPoint(protocol, addr)
	if err != nil {
		return nil, err
	}
	listenpoint.SetTransportChannel(stack.transportChannel)
	stack.listeningPoints.Store(protocol, listenpoint)
	return listenpoint, nil
}

func (stack *SipStack) SetListener(listener transport.Listener) {
	stack.listener = listener
}

func (stack *SipStack) SetFuncHandler(method method.Method, handler func(message.Message)) {
	stack.funcMap[method] = handler
}

func (stack *SipStack) Start(ctx context.Context) {
	defer func() {
		stack.listeningPoints.Range(func(key, value any) bool {
			return true
		})

		fmt.Println("sip stack  stop")
	}()
	stack.listeningPoints.Range(func(key, value any) bool {
		if lp, ok := value.(transport.ListeningPoint); ok {
			go lp.Start()
		}
		return true
	})

	stack.ctx = ctx
	for {
		select {
		case <-stack.ctx.Done():
			return
		case msg := <-stack.transportChannel:
			if stack.listener != nil {
				if resp, ok := msg.(message.Response); ok {
					go stack.listener.HandleResponse(resp)
					continue
				}
				if req, ok := msg.(message.Request); ok {
					go stack.listener.HandleRequest(req)
					continue
				}
			}
			if stack.funcMap != nil {
				if m, ok := msg.CSeq(); ok {
					if f, ok := stack.funcMap[m.Method]; ok {
						go f(msg)
						continue
					}
				}
			}
		}
	}
}

func (stack *SipStack) Send(protocol string, address string, msg message.Message) error {
	protocol = strings.ToLower(protocol)
	if lp, ok := stack.listeningPoints.Load(protocol); ok {
		return lp.(transport.ListeningPoint).Send(address, msg)
	}
	return fmt.Errorf("not found %s listening point", protocol)
}
