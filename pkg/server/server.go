package server

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-av/gosip/pkg/authentication"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/sip"
	"github.com/go-av/gosip/pkg/utils"
	"github.com/sirupsen/logrus"
)

type response struct {
	err     error
	content Content
}

type server struct {
	protocols []string

	ctx        context.Context
	cancelFunc context.CancelFunc

	stack     *sip.SipStack
	needauth  bool
	handler   ServerHandler
	address   *message.Address
	responses *sync.Map
}

func NewServer(handler ServerHandler) *server {
	s := &server{
		needauth:  true,
		handler:   handler,
		responses: &sync.Map{},
	}
	handler.SetServer(s)
	return s
}

func (s *server) ListenUDPServer(ctx context.Context, addr string, protocols []string) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	s.ctx = ctx
	s.cancelFunc = cancelFunc
	s.protocols = protocols

	hostAndport, err := utils.ParseHostAndPort(addr)
	if err != nil {
		return err
	}

	s.address = message.NewAddress("", hostAndport.Host, hostAndport.Port)

	s.stack = sip.NewSipStack(addr)
	for _, protocol := range s.protocols {
		_, err := s.stack.CreateListenPoint(protocol, addr)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}

	s.stack.SetListener(s)
	s.stack.Start(ctx)
	return nil
}

func (s *server) HandleRequests(req message.Request) {
	user := ""
	if from, ok := req.From(); ok {
		user = from.Address.User
	}
	protocol, adddress := req.RequestFrom()
	client, err := s.handler.GetClient(user)
	if err != nil {
		resp := message.NewResponse(req, 500, err.Error())
		s.stack.Send(protocol, adddress, resp)
		logrus.Error(err)
		return
	}
	client.SetTransport(protocol, adddress)

	if !client.IsAuth() && s.needauth {
		authheader, ok := req.Authorization()
		if !ok {
			resp := message.NewResponse(req, 401, "Unauthorized")
			resp.SetHeader(message.NewWWWAuthenticateHeader(s.handler.Realm(), utils.RandString(50)))
			s.stack.Send(protocol, adddress, resp)
			return
		}

		auth := authentication.Parse(authheader.Value())
		if auth.Response() != auth.Clone().Auth(auth.Username(), client.Password(), string(req.Method()), auth.Uri()).Response() {
			resp := message.NewResponse(req, 403, "Password Error")
			s.stack.Send(protocol, adddress, resp)
			return
		}
		client.SetAuth(true)
	}

	switch req.Method() {
	case method.REGISTER:
		var expires int64 = 0
		if ex, ok := req.Expires(); ok {
			expires = int64(*ex)
		}

		if contact, ok := req.Contact(); ok {
			if str, ok := contact.Params.Get("expires"); ok {
				n, _ := strconv.ParseInt(str, 10, 64)
				expires = n
			}
		}

		// 注销
		if expires == 0 {
			client.Logout() // 设备注销
		}
		resp := message.NewResponse(req, 200, "OK")
		s.stack.Send(protocol, adddress, resp)
		return
	case method.MESSAGE:
		contentTypeHeader, ok := req.ContentType()
		if !ok {
			return
		}
		statusCode, reason := s.handler.ReceiveMessage(NewContent(contentTypeHeader.Value(), req.Body()))
		resp := message.NewResponse(req, message.StatusCode(statusCode), reason)
		s.stack.Send(protocol, adddress, resp)
	default:
		fmt.Println("---------==========-----------")
		fmt.Println("---------=====", req.Method(), "=====-----------")
		fmt.Println("---------==========-----------")
		resp := message.NewResponse(req, 200, "OK")
		s.stack.Send(protocol, adddress, resp)
	}
}

func (s *server) HandleResponses(resp message.Response) {
	cseq, ok := resp.CSeq()
	if !ok {
		return
	}
	switch cseq.Method {
	case method.MESSAGE:
		callID, ok := resp.CallID()
		if ok {
			if callback, ok := s.responses.Load(callID.Value()); ok {
				r := response{}
				if !resp.IsSuccess() {
					r.err = errors.New(resp.Reason())
				}
				if contentType, ok := resp.ContentType(); ok {
					r.content = NewContent(contentType.Value(), resp.Body())
				}
				callback.(chan response) <- r
			}
		}
	default:
		fmt.Println(cseq.Method, resp.StartLine(), "暂未处理")
	}
}

func (s *server) ServerAddress() *message.Address {
	return s.address
}

func (s *server) Send(protocol string, address string, msg message.Message) error {
	return s.stack.Send(protocol, address, msg)
}

func (s *server) SendMessage(client Client, content Content) (Content, error) {
	callID := utils.RandString(30)
	respChan := make(chan response, 1)
	s.responses.Store(callID, respChan)
	protocol, address := client.Transport()
	hostAndPort, _ := utils.ParseHostAndPort(address)
	clientAddress := message.NewAddress(client.User(), hostAndPort.Host, hostAndPort.Port)
	msg := message.NewRequestMessage(protocol, method.MESSAGE, clientAddress)
	msg.AppendHeader(
		message.NewViaHeader(protocol, s.address.Host, s.address.Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		message.NewAllowHeader(),
		message.NewCSeqHeader(1, method.MESSAGE),
		message.NewFromHeader("", s.address.Clone().SetUser(client.User()), message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader("", clientAddress, nil),
		message.NewCallIDHeader(callID),
		message.NewMaxForwardsHeader(70),
	)
	msg.SetBody(content.ContentType(), content.Data())
	err := s.stack.Send(protocol, address, msg)
	if err != nil {
		return nil, err
	}

	t := time.NewTimer(10 * time.Second)
	select {
	case <-t.C:
		return nil, errors.New("请求超时")
	case resp := <-respChan:
		t.Stop()
		if resp.err != nil {
			return nil, err
		}
		return resp.content, nil
	}
}
