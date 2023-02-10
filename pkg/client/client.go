package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-av/gosip/pkg/client/dialog"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/sdp"
	"github.com/go-av/gosip/pkg/sip"
	"github.com/go-av/gosip/pkg/types"
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

	address       *message.Address // 客户端的地址及端口
	serverAddrees *message.Address // 服务器地址
	transport     string           // 传输协议  UDP or TCP
	dialogMgr     *dialog.DialogManger
	dialogs       chan dialog.Dialog

	sdp func(*sdp.SDP) *sdp.SDP
}

func NewClient(displayName string, user string, password string, transport string, host string, port types.Port) *Client {
	client := &Client{
		displayName: displayName,
		user:        user,
		password:    password,
		stack:       sip.NewSipStack(user),
		transport:   transport,
		address: &message.Address{
			User: user,
			Port: port,
			Host: host,
		},
		dialogs: make(chan dialog.Dialog, 10),
	}
	client.dialogMgr = dialog.NewDialogManger(client)
	return client
}

func (client *Client) Start(ctx context.Context, host string, port int) error {
	client.serverAddrees = &message.Address{
		Host: host,
		Port: types.Port(port),
		User: client.user,
	}

	client.stack.CreateListenPoint(client.transport, client.address.Host, int(client.address.Port))
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
func (client *Client) Transport() string {
	return client.transport
}

// 暂时未做认证
func (client *Client) registrar(expire int, resp message.Response) error {
	msg := message.NewRequestMessage(strings.ToUpper(client.transport), method.REGISTER, client.serverAddrees.Clone())
	contactParam := message.NewParams()
	if expire >= 0 {
		contactParam.Set("expires", fmt.Sprintf("%d", expire))
		msg.AppendHeader(message.NewExpiresHeader(expire))
	}

	msg.AppendHeader(
		message.NewViaHeader(strings.ToUpper(client.transport), client.address.Host, client.address.Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		message.NewAllowHeader(),
		message.NewCSeqHeader(1, method.REGISTER),
		message.NewFromHeader(client.displayName, client.address, message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader(client.displayName, client.address, nil),
		message.NewCallIDHeader(utils.RandString(20)),
		message.NewMaxForwardsHeader(70),
		message.NewContactHeader(client.displayName, client.address, client.transport, contactParam),
		message.NewSupportedHeader([]string{"replaces", "outbound", "gruu"}),
	)

	if resp != nil {
		authHeader, ok := resp.WWWAuthenticate()
		if ok {
			msg.AppendHeader(authHeader.Auth(client.user, client.password, client.serverAddrees.String()))
		}

		if cseq, ok := resp.CSeq(); ok {
			cseq.SeqNo += 1
			msg.SetHeader(cseq)
		}
	}

	err := client.Send(client.serverAddrees, msg)
	if err != nil {
		logrus.Errorf("%s registrar failed: %s", client.user, err)
		return err
	}
	return nil
}

func (client *Client) HandleRequests(msg message.Request) {
	switch msg.Method() {
	case method.INVITE, method.ACK, method.BYE:
		dialog := client.dialogMgr.HandleMessage(msg)
		if dialog != nil {
			client.dialogs <- dialog
		}
	default:
		resp := message.NewResponse(msg, 200, "Ok")
		err := client.Send(client.serverAddrees, resp)
		if err != nil {
			logrus.Error(err)
		}
	}
}

func (client *Client) HandleResponses(msg message.Response) {
	cseq, ok := msg.CSeq()
	if !ok {
		return
	}

	switch cseq.Method {
	case method.REGISTER:
		switch msg.StatusCode() {
		case 200:
			client.auth = true
			if client.authCallback != nil {
				client.authCallback(nil)
			}
			var d = time.Duration(1 * time.Second)
			if con, ok := msg.Contact(); ok {
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
			client.registrar(4800, msg)
		case 403, 404:
			client.auth = false
			if client.authCallback != nil {
				client.authCallback(fmt.Errorf(msg.Reason()))
			}
		}

	case method.INVITE:
		client.dialogMgr.HandleMessage(msg)
	case method.BYE:
		client.dialogMgr.HandleMessage(msg)
	default:
		fmt.Println("\n ====== 未处理 ======")
		fmt.Println(msg.String())
		fmt.Println(" ====== 未处理 ======")
	}
}

func (client *Client) Call(user string) (dialog.Dialog, error) {
	if !client.auth {
		return nil, errors.New("Unauthorized")
	}
	callID := utils.RandString(30)
	da := client.serverAddrees.Clone()
	da.User = user
	da.Port = 0
	msg := message.NewRequestMessage("UDP", method.INVITE, da)

	msg.AppendHeader(
		message.NewViaHeader("UDP", client.address.Host, client.address.Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		message.NewAllowHeader(),
		message.NewCSeqHeader(1, method.INVITE),
		message.NewFromHeader(client.displayName, message.NewAddress(client.user, client.serverAddrees.Host, 0), message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader("", message.NewAddress(user, client.serverAddrees.Host, 0), nil),
		message.NewCallIDHeader(callID),
		message.NewMaxForwardsHeader(70),
		message.NewContactHeader(client.displayName, client.address, client.transport, message.NewParams().Set("expires", "3600")),
		message.NewAllowEventHeader("talk"),
	)

	msg.SetBody(client.sdp(nil))
	err := client.stack.Send(client.serverAddrees, msg)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	dialog := client.dialogMgr.HandleMessage(msg)
	return dialog, nil
}

func (client *Client) Send(address *message.Address, msg message.Message) error {
	return client.stack.Send(address, msg)
}

func (client *Client) Address() *message.Address {
	return client.serverAddrees
}

func (client *Client) Dialog() chan dialog.Dialog {
	return client.dialogs
}

func (client *Client) User() string {
	return client.user
}

func (client *Client) SDP(sd *sdp.SDP) *sdp.SDP {
	return client.sdp(sd)
}

func (client *Client) SetSDP(sd func(*sdp.SDP) *sdp.SDP) {
	client.sdp = sd
}
