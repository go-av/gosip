package server

import (
	"context"
	"fmt"

	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/sip"
)

type Server struct {
	protocols []string

	ctx        context.Context
	cancelFunc context.CancelFunc

	stack *sip.SipStack
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) ListenUDPServer(ctx context.Context, addr string, protocols []string) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	s.ctx = ctx
	s.cancelFunc = cancelFunc
	s.protocols = protocols

	s.stack = sip.NewSipStack(addr)
	for _, protocol := range s.protocols {
		_, err := s.stack.CreateListenPoint(protocol, addr)
		if err != nil {
			panic(err)
		}
	}

	s.stack.SetListener(s)

	go s.stack.Start(ctx)

	return nil
}

func (s *Server) HandleRequests(req message.Request) {
	fmt.Println("req", req.Method())
}

func (s *Server) HandleResponses(message.Response) {

}
