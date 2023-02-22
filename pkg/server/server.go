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
	err  error
	body message.Body
}

type server struct {
	protocols []string

	ctx        context.Context
	cancelFunc context.CancelFunc

	stack     *sip.SipStack
	needauth  bool
	handler   Handler
	address   *message.Address
	responses *sync.Map
}

func NewServer(handler Handler) *server {
	s := &server{
		needauth:  true,
		handler:   handler,
		responses: &sync.Map{},
	}
	handler.SetServer(s)
	return s
}

// monitorIP 监控ID，可指定监听IP，或设置为 0.0.0.0 为监听所有
// ip 对外暴露的IP
// port 监听端口
func (s *server) ListenUDPServer(ctx context.Context, monitorIP string, ip string, port uint16, protocols []string) error {
	if monitorIP == "" {
		monitorIP = "0.0.0.0"
	}
	logrus.Infof("ListenUDPServer: %s(%s):%d", ip, monitorIP, port)
	ctx, cancelFunc := context.WithCancel(ctx)
	s.ctx = ctx
	s.cancelFunc = cancelFunc
	s.protocols = protocols

	s.address = message.NewAddress("", ip, port)

	s.stack = sip.NewSipStack(ip)
	for _, protocol := range s.protocols {
		_, err := s.stack.CreateListenPoint(protocol, fmt.Sprintf("%s:%d", monitorIP, port))
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
		response, err := s.handler.ReceiveMessage(message.NewBody(contentTypeHeader.Value(), req.Body()))
		if err != nil {
			logrus.Error(err)
			return
		}

		if response == nil || response.code == 0 {
			return
		}

		resp := message.NewResponse(req, message.StatusCode(response.code), response.reason)
		if response.body != nil {
			resp.SetBody(string(response.contentType), response.body)
		}
		s.stack.Send(protocol, adddress, resp)
	default:
		fmt.Println("-----------==========-----------")
		fmt.Println("---------", req.Method(), "-----------")
		fmt.Println("-----------==========-----------")
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
					r.body = message.NewBody(contentType.Value(), resp.Body())
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

func (s *server) SendMessage(client Client, request message.Request) (message.Body, error) {
	callID := utils.RandString(30)
	request.SetHeader(message.NewCallIDHeader(callID))

	respChan := make(chan response, 1)
	s.responses.Store(callID, respChan)
	defer s.responses.Delete(callID)

	protocol, address := client.Transport()

	err := s.stack.Send(protocol, address, request)
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
		return resp.body, nil
	}
}
