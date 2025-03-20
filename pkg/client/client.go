package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/sip"
	"github.com/go-av/gosip/pkg/utils"
	"github.com/sirupsen/logrus"
)

type Client struct {
	ctx        context.Context
	cancelFunc context.CancelFunc

	displayName string
	user        string
	password    string

	auth         bool
	authCallback func(msg message.Response)

	stack *sip.SipStack

	localAddr  *utils.HostAndPort
	serverAddr *utils.HostAndPort

	protocol string // 传输协议  UDP or TCP

	handler   Handler
	dialogs   sync.Map
	responses sync.Map

	receive     chan dialog.Dialog // 接收到的会话
	once        sync.Once
	loginTicker *time.Ticker

	loginExpire          int // 注册有效期单位秒
	requestUser          string
	updateRegisterHeader func(expire int, req message.Message, resp message.Message)
}

func NewClient(ctx context.Context, displayName string, user string, password string, address string, handler Handler) (*Client, error) {
	addr, err := utils.ParseHostAndPort(address)
	if err != nil {
		return nil, err
	}
	ctx, cancelFunc := context.WithCancel(ctx)
	client := &Client{
		ctx:         ctx,
		cancelFunc:  cancelFunc,
		displayName: displayName,
		user:        user,
		password:    password,
		stack:       sip.NewSipStack(user),
		localAddr:   addr,
		receive:     make(chan dialog.Dialog, 10),
		handler:     handler,
		loginExpire: 3600,
		requestUser: user,
	}

	return client, nil
}
func (client *Client) IsAuth() bool {
	return client.auth
}

func (client *Client) Logout() error {
	return client.Login(0, nil)
}

func (client *Client) Registrar(address string, protocol string) error {
	client.protocol = protocol
	logrus.Infof("client %s registrar %s(%s)", client.user, address, protocol)

	addr, err := utils.ParseHostAndPort(address)
	if err != nil {
		return err
	}

	client.serverAddr = addr
	_, err = client.stack.CreateListenPoint(client.protocol, fmt.Sprintf("0.0.0.0:%d", client.localAddr.Port))
	if err != nil {
		return err
	}

	client.stack.SetListener(client)
	go client.stack.Start(client.ctx)
	time.Sleep(1 * time.Second)
	client.once.Do(func() {
		client.loginTicker = time.NewTicker(1 * time.Minute)
		go func() {
			for {
				<-client.loginTicker.C
				client.Login(-1, nil)
			}
		}()
	})
	return client.Login(-1, nil)
}

func (client *Client) WithAuthCllback(callback func(resp message.Response)) {
	client.authCallback = callback
}

func (client *Client) WithLoginExpire(expire int) {
	client.loginExpire = expire
}

func (client *Client) WithRequestUser(requestUser string) {
	client.requestUser = requestUser
}

func (client *Client) WithUpdateRegisterHeader(updateRegisterHeader func(expire int, req message.Message, resp message.Message)) {
	client.updateRegisterHeader = updateRegisterHeader
}

func (client *Client) Login(expire int, resp message.Response) error {
	msg := message.NewRequestMessage(client.protocol, method.REGISTER, message.NewAddress(client.requestUser, client.serverAddr.Host, client.serverAddr.Port))
	if expire < 0 {
		expire = client.loginExpire
	}
	contactParam := message.NewParams()
	contactParam.Set("expires", fmt.Sprintf("%d", expire))

	if expire != 0 && client.loginTicker != nil {
		client.loginTicker.Reset(1 * time.Minute)
	}

	localAddr := message.NewAddress(client.user, client.localAddr.Host, client.localAddr.Port)
	msg.AppendHeader(
		message.NewViaHeader(client.protocol, client.localAddr.Host, client.localAddr.Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		message.NewAllowHeader(),
		message.NewCSeqHeader(uint32(time.Now().Unix()), method.REGISTER),
		message.NewFromHeader(client.displayName, localAddr, message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader(client.displayName, localAddr, nil),
		message.NewCallIDHeader(utils.RandString(20)),
		message.NewMaxForwardsHeader(70),
		message.NewContactHeader(client.displayName, localAddr, "", contactParam),
		message.NewExpiresHeader(expire),
	)

	if resp != nil {
		authHeader, ok := resp.WWWAuthenticate()
		if ok {
			msg.SetHeader(authHeader.Auth(client.user, client.password, "sip:"+client.serverAddr.Host))
		}

		if cseq, ok := resp.CSeq(); ok {
			cseq.SeqNo += 1
			msg.SetHeader(cseq)
		}
		if callID, ok := resp.CallID(); ok {
			msg.SetHeader(callID)
		}
	}

	if client.updateRegisterHeader != nil {
		client.updateRegisterHeader(expire, msg, resp)
	}

	err := client.Send(client.protocol, client.serverAddr.String(), msg)
	if err != nil {
		logrus.Errorf("%s registrar failed: %s", client.user, err)
		return err
	}
	return nil
}

func (client *Client) HandleRequest(req message.Request) {
	switch req.Method() {
	case method.INVITE:
		to, _ := req.To()
		to.Params.Set("tag", utils.RandString(10))
		req.SetHeader(to)

		resp := message.NewResponse(req, 100, "Trying")
		err := client.Send(client.protocol, client.serverAddr.String(), resp)
		if err != nil {
			return
		}

		callID, ok := req.CallID()
		if !ok {
			return
		}

		if _, ok := client.dialogs.Load(callID.Value()); ok {
			resp = message.NewResponse(req, 400, "Bad Request:"+"会话已经存在！")
			err := client.Send(client.protocol, client.serverAddr.String(), resp)
			if err != nil {
				logrus.Errorf("%s send error: %s", client.user, err)
			}
			return
		}

		from, _ := req.From()

		dl, err := dialog.Receive(client, dialog.NewFrom(from.DisplayName, from.Address.User, client.protocol, client.serverAddr.String()), dialog.NewTo(to.Address.User, client.localAddr.String()), callID.Value(), req)
		if err != nil {
			resp = message.NewResponse(req, 500, err.Error())
			_ = client.Send(client.protocol, client.serverAddr.String(), resp)
			return
		}

		client.dialogs.Store(callID.Value(), dl)
		go dl.Run(func(callID string) {
			client.dialogs.Delete(callID)
		})
		client.receive <- dl
	case method.ACK, method.BYE, method.CANCEL:
		callID, ok := req.CallID()
		if !ok {
			return
		}

		if v, ok := client.dialogs.Load(callID.Value()); ok {
			dl := v.(dialog.Dialog)
			dl.HandleRequest(req)
			return
		}

		if req.Method() == method.BYE || req.Method() == method.CANCEL {
			resp := message.NewResponse(req, 200, "OK")
			_ = client.Send(client.protocol, client.serverAddr.String(), resp)
			return
		}

		resp := message.NewResponse(req, 404, "Dialog not found")
		_ = client.Send(client.protocol, client.serverAddr.String(), resp)

	default:
		if client.handler == nil {
			return
		}

		contentTypeHeader, ok := req.ContentType()
		if !ok {
			return
		}

		response, err := client.handler.ReceiveMessage(client.ctx, req.Method(), req, message.NewBody(contentTypeHeader.Value(), req.Body()))
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
		if response.header != nil {
			resp.AppendHeader(response.header.Headers()...)
		}

		client.stack.Send(client.protocol, client.serverAddr.String(), resp)
	}
}

func (client *Client) HandleResponse(resp message.Response) {
	cseq, ok := resp.CSeq()
	if !ok {
		return
	}

	switch cseq.Method {
	case method.REGISTER:
		switch resp.StatusCode() {
		case 200:
			client.auth = true
			if client.authCallback != nil {
				client.authCallback(resp)
			}

			ex := false

			if expires, ok := resp.Expires(); ok {
				expire := int(*expires)
				if (expire - 10) > 0 {
					ex = true
					client.loginTicker.Reset(time.Duration((expire - 10)) * time.Second)
				}
			}

			if con, ok := resp.Contact(); !ex && ok {
				if param, ok := con.Params.Get("expires"); ok {
					expire, _ := strconv.ParseInt(param, 10, 64)
					if (expire - 10) > 0 {
						ex = true
						client.loginTicker.Reset(time.Duration((expire - 10)) * time.Second)
					}
				}
			}

			if !ex {
				client.loginTicker.Reset(time.Duration(client.loginExpire-10) * time.Second)
			}

		case 401:
			client.auth = false
			client.Login(-1, resp)
		default:
			client.auth = false
			if client.authCallback != nil {
				client.authCallback(resp)
			}
		}

	case method.INVITE, method.ACK, method.BYE, method.CANCEL:
		callID, ok := resp.CallID()
		if !ok {
			return
		}

		if v, ok := client.dialogs.Load(callID.Value()); ok {
			dl := v.(dialog.Dialog)
			dl.HandleResponse(resp)
		}
	default:
		if resp.StatusCode() == 401 {
			client.auth = false
			client.Login(-1, nil)
		}
		callID, ok := resp.CallID()
		if ok {
			if callback, ok := client.responses.Load(callID.Value()); ok {
				r := response{}
				if !resp.IsSuccess() {
					r.err = errors.New(resp.Reason())
				}
				r.resp = resp
				callback.(chan response) <- r
			}
		} else {
			logrus.Debugf("Client 消息 %s 未处理", resp.String())
		}
	}
}

func (client *Client) Call(ctx context.Context, user string, sdp string) (dialog.Dialog, error) {
	return client.CallWithUpdateMessage(ctx, user, sdp, nil)
}

func (client *Client) CallWithUpdateMessage(ctx context.Context, user string, sdp string, updateMsg func(message.Message)) (dialog.Dialog, error) {
	if !client.auth {
		return nil, errors.New("Unauthorized")
	}
	dl, err := dialog.Invite(ctx, client,
		dialog.NewFrom(client.displayName, client.user, client.protocol, client.localAddr.String()),
		dialog.NewTo(user, client.serverAddr.String()), []byte(sdp), updateMsg)

	if err != nil {
		return nil, err
	}
	go dl.Run(func(callID string) {
		client.dialogs.Delete(callID)
	})
	client.dialogs.Store(dl.DialogID(), dl)
	return dl, nil
}

func (client *Client) Send(protocol string, address string, msg message.Message) error {
	return client.stack.Send(protocol, address, msg)
}

func (client *Client) Receive() chan dialog.Dialog {
	return client.receive
}

type response struct {
	err  error
	resp message.Response
}

func (client *Client) SendMessage(request message.Request) (message.Response, error) {
	callID := utils.RandString(30)
	request.SetHeader(message.NewCallIDHeader(callID))

	respChan := make(chan response, 1)
	client.responses.Store(callID, respChan)
	defer client.responses.Delete(callID)

	err := client.stack.Send(client.protocol, client.serverAddr.String(), request)
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

		if resp.resp == nil {
			return nil, errors.New("resp is nil")
		}

		return resp.resp, nil
	}
}
