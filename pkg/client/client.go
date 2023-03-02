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
	authCallback func(err error)

	stack *sip.SipStack

	localAddr  *utils.HostAndPort
	serverAddr *utils.HostAndPort

	protocol string // 传输协议  UDP or TCP
	dialogs  *sync.Map
	receive  chan dialog.Dialog // 接收到的会话
}

func NewClient(displayName string, user string, password string, protocol string, address string) (*Client, error) {
	addr, err := utils.ParseHostAndPort(address)
	if err != nil {
		return nil, err
	}

	client := &Client{
		displayName: displayName,
		user:        user,
		password:    password,
		stack:       sip.NewSipStack(user),
		protocol:    protocol,
		localAddr:   addr,
		dialogs:     &sync.Map{},
		receive:     make(chan dialog.Dialog, 10),
	}
	return client, nil
}

func (client *Client) Start(ctx context.Context, address string) error {
	logrus.Infof("client %s registrar %s", client.user, address)

	addr, err := utils.ParseHostAndPort(address)
	if err != nil {
		return err
	}
	client.serverAddr = addr
	_, err = client.stack.CreateListenPoint(client.protocol, client.localAddr.String())
	if err != nil {
		return err
	}
	client.stack.SetListener(client)
	ctx, cancelFunc := context.WithCancel(ctx)
	client.ctx = ctx
	client.cancelFunc = cancelFunc
	go client.stack.Start(ctx)
	time.Sleep(1 * time.Second)
	go func() {
		<-client.ctx.Done()
		// 注销
		client.registrar(0, nil)
	}()

	if err := client.registrar(-1, nil); err != nil {
		cancelFunc()
		return err
	}
	t := time.NewTicker(1 * time.Minute)
	autherr := make(chan error, 1)
	client.authCallback = func(err error) {
		autherr <- err
	}
	select {
	case <-t.C:
		cancelFunc()
		return fmt.Errorf("认证超时")
	case err := <-autherr:
		return err
	}
}

// 暂时未做认证
func (client *Client) registrar(expire int, resp message.Response) error {
	msg := message.NewRequestMessage(client.protocol, method.REGISTER, message.NewAddress(client.user, client.serverAddr.Host, client.serverAddr.Port))

	contactParam := message.NewParams()
	if expire >= 0 {
		contactParam.Set("expires", fmt.Sprintf("%d", expire))
		msg.AppendHeader(message.NewExpiresHeader(expire))
	}

	localAddr := message.NewAddress(client.user, client.localAddr.Host, client.localAddr.Port)
	msg.AppendHeader(
		message.NewViaHeader(client.protocol, client.localAddr.Host, client.localAddr.Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		message.NewAllowHeader(),
		message.NewCSeqHeader(1, method.REGISTER),
		message.NewFromHeader(client.displayName, localAddr, message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader(client.displayName, localAddr, nil),
		message.NewCallIDHeader(utils.RandString(20)),
		message.NewMaxForwardsHeader(70),
		message.NewContactHeader(client.displayName, localAddr, client.protocol, contactParam),
		message.NewSupportedHeader([]string{"replaces", "outbound", "gruu"}),
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
		}

	default:
		// resp := message.NewResponse(msg, 200, "Ok")
		// err := client.Send(client.serverAddr.String(), resp)
		// if err != nil {
		// 	logrus.Error(err)
		// }
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
				client.authCallback(nil)
			}
			var d = time.Duration(1 * time.Second)
			if con, ok := resp.Contact(); ok {
				if param, ok := con.Params.Get("expires"); ok {
					expire, _ := strconv.ParseInt(param, 10, 64)
					if (expire - 10) > 0 {
						d = time.Duration((expire - 10)) * time.Second
					}
				}
			}
			time.Sleep(d)
			client.registrar(-1, nil)
		case 401:
			client.auth = false
			client.registrar(4800, resp)
		case 403, 404:
			client.auth = false
			if client.authCallback != nil {
				client.authCallback(fmt.Errorf(resp.Reason()))
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
		logrus.Debugf("Client 消息 %s 未处理", resp.String())
	}
}

func (client *Client) Call(ctx context.Context, user string, sdp string) (dialog.Dialog, error) {
	if !client.auth {
		return nil, errors.New("Unauthorized")
	}
	dl, err := dialog.Invite(ctx, client,
		dialog.NewFrom(client.displayName, client.user, client.protocol, client.localAddr.String()),
		dialog.NewTo(user, client.serverAddr.String()), []byte(sdp), nil)

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
