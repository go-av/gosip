package server

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-av/gosip/pkg/authentication"
	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/sip"
	"github.com/go-av/gosip/pkg/utils"
	"github.com/sirupsen/logrus"
)

type response struct {
	err  error
	resp message.Response
}

type server struct {
	protocols []string

	ctx        context.Context
	cancelFunc context.CancelFunc

	stack     *sip.SipStack
	needauth  bool
	handler   Handler
	address   *message.Address
	responses sync.Map
	dialog    sync.Map           // 呼叫会话管理
	receive   chan dialog.Dialog // 接收到的会话
}

func NewServer(needauth bool, handler Handler) *server {
	s := &server{
		needauth: needauth,
		handler:  handler,
		receive:  make(chan dialog.Dialog, 5),
	}
	return s
}

// monitorIP 监控ID，可指定监听IP，或设置为 0.0.0.0 为监听所有
// ip 对外暴露的IP
// port 监听端口
func (s *server) SIPListen(ctx context.Context, monitorIP string, ip string, port uint16, protocols ...string) error {
	if monitorIP == "" {
		monitorIP = "0.0.0.0"
	}
	logrus.Infof("SIPListen: %s:%d", monitorIP, port)
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

func (s *server) HandleRequest(req message.Request) {
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
	case method.INVITE:
		to, _ := req.To()
		to.Params.Set("tag", utils.RandString(10))
		req.SetHeader(to)

		resp := message.NewResponse(req, 100, "Trying")
		err := s.stack.Send(protocol, adddress, resp)
		if err != nil {
			return
		}

		callID, ok := req.CallID()
		if !ok {
			return
		}

		if _, ok := s.dialog.Load(callID.Value()); ok {
			resp = message.NewResponse(req, 400, "Bad Request:"+"会话已经存在！")
			_ = s.Send(protocol, adddress, resp)
			return
		}

		from, _ := req.From()

		dl, err := dialog.Receive(s, dialog.NewFrom(from.DisplayName, from.Address.User, protocol, adddress), dialog.NewTo(to.Address.User, (&utils.HostAndPort{
			Host: s.address.Host,
			Port: s.address.Port,
		}).String()), callID.Value(), req)
		if err != nil {
			resp = message.NewResponse(req, 500, err.Error())
			_ = s.Send(protocol, adddress, resp)
			return
		}

		s.dialog.Store(callID.Value(), dl)
		go dl.Run(func(callID string) {
			s.dialog.Delete(callID)
		})
		s.receive <- dl
	case method.ACK, method.BYE, method.CANCEL:
		callID, ok := req.CallID()
		if !ok {
			return
		}

		if v, ok := s.dialog.Load(callID.Value()); ok {
			dl := v.(dialog.Dialog)
			dl.HandleRequest(req)
		} else {
			resp := message.NewResponse(req, 200, "success")
			_ = s.Send(protocol, adddress, resp)
		}

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
	case method.INFO:

	case method.MESSAGE:
		contentTypeHeader, ok := req.ContentType()
		if !ok {
			return
		}
		response, err := s.handler.ReceiveMessage(s.ctx, client, message.NewBody(contentTypeHeader.Value(), req.Body()))
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
		logrus.Debugf("Server Req %s 未处理", req.Method())
		resp := message.NewResponse(req, 200, "OK")
		s.stack.Send(protocol, adddress, resp)
	}
}

func (s *server) HandleResponse(resp message.Response) {
	cseq, ok := resp.CSeq()
	if !ok {
		return
	}
	switch cseq.Method {
	case method.INVITE, method.ACK, method.BYE, method.CANCEL:
		callID, ok := resp.CallID()
		if !ok {
			return
		}

		if v, ok := s.dialog.Load(callID.Value()); ok {
			dl := v.(dialog.Dialog)
			dl.HandleResponse(resp)
		}

	default:
		callID, ok := resp.CallID()
		if ok {
			if callback, ok := s.responses.Load(callID.Value()); ok {
				r := response{}
				if !resp.IsSuccess() {
					r.err = errors.New(resp.Reason())
				}
				r.resp = resp
				callback.(chan response) <- r
			}
		} else {
			logrus.Debugf("Server %s %s 未处理", cseq.Method, resp.StartLine())
		}

	}
}

func (s *server) ServerAddress() *message.Address {
	return s.address
}

func (s *server) Send(protocol string, address string, msg message.Message) error {
	return s.stack.Send(protocol, address, msg)
}

func (s *server) SendMessage(client Client, request message.Request) (message.Response, error) {
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

		return resp.resp, nil
	}
}

func (s *server) Invite(ctx context.Context, from dialog.From, to dialog.To, sdp string, updateMsg func(msg message.Message)) (dialog.Dialog, error) {
	dl, err := dialog.Invite(ctx, s, from, to, []byte(sdp), updateMsg)
	if err != nil {
		return nil, err
	}
	go dl.Run(func(callID string) {
		s.dialog.Delete(callID)
	})
	s.dialog.Store(dl.DialogID(), dl)
	return dl, nil
}

func (s *server) Receive() chan dialog.Dialog {
	return s.receive
}
