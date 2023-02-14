package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-av/gosip/pkg/client/dialog"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/sdp"
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
	// address       *message.Address // 客户端的地址及端口
	// serverAddrees *message.Address // 服务器地址
	protocol  string // 传输协议  UDP or TCP
	dialogMgr *dialog.DialogManger
	dialogs   chan dialog.Dialog

	sdp func(*sdp.SDP) *sdp.SDP
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
		dialogs:     make(chan dialog.Dialog, 10),
	}
	client.dialogMgr = dialog.NewDialogManger(client)
	return client, nil
}

func (client *Client) Start(ctx context.Context, address string) error {
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
	fmt.Println("xxx")

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

func (client *Client) Stop() {
	client.cancelFunc()
}

func (client *Client) Protocol() string {
	return client.protocol
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
			msg.AppendHeader(authHeader.Auth(client.user, client.password, client.serverAddr.String()))
		}

		if cseq, ok := resp.CSeq(); ok {
			cseq.SeqNo += 1
			msg.SetHeader(cseq)
		}
	}

	err := client.Send(client.serverAddr.String(), msg)
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
		err := client.Send(client.serverAddr.String(), resp)
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

	msg := message.NewRequestMessage(client.protocol, method.INVITE, message.NewAddress(user, client.serverAddr.Host, 0))

	msg.AppendHeader(
		message.NewViaHeader(client.protocol, client.localAddr.Host, client.localAddr.Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		message.NewAllowHeader(),
		message.NewCSeqHeader(1, method.INVITE),
		message.NewFromHeader(client.displayName, message.NewAddress(client.user, client.serverAddr.Host, 0), message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader("", message.NewAddress(user, client.serverAddr.Host, 0), nil),
		message.NewCallIDHeader(callID),
		message.NewMaxForwardsHeader(70),
		message.NewContactHeader(client.displayName, message.NewAddress(user, client.localAddr.Host, client.localAddr.Port), client.protocol, message.NewParams().Set("expires", "3600")),
		message.NewAllowEventHeader("talk"),
	)

	msg.SetBody(client.sdp(nil))
	err := client.stack.Send(client.protocol, client.serverAddr.String(), msg)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	dialog := client.dialogMgr.HandleMessage(msg)
	return dialog, nil
}

func (client *Client) Send(address string, msg message.Message) error {
	return client.stack.Send(client.protocol, address, msg)
}

func (client *Client) Address() *utils.HostAndPort {
	return client.localAddr
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
