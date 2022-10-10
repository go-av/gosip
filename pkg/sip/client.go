package sip

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/sdp"
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

	stack *SipStack

	address       *message.Address // 客户端的地址及端口
	serverAddrees *message.Address // 服务器地址
	transport     string           // 传输协议  UDP or TCP
	dialogMgr     *dialog.DialogManger
	dialogs       chan dialog.Dialog

	sdp func(*sdp.SDP) *sdp.SDP
}

func NewClient(displayName string, user string, password string, host string, port types.Port) *Client {
	client := &Client{
		displayName: displayName,
		user:        user,
		password:    password,
		stack:       NewSipStack(user),
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

func (client *Client) Start(ctx context.Context, transport string, host string, port int) error {
	client.serverAddrees = &message.Address{
		Host: host,
		Port: types.Port(port),
		User: client.user,
	}

	client.transport = transport

	client.stack.CreateListenPoint(transport, client.address.Host, int(client.address.Port))
	// client.stack.CreateListenPoint(transport, "0.0.0.0", int(client.address.Port))

	client.stack.SetListener(client)

	ctx, cancelFunc := context.WithCancel(ctx)

	client.ctx = ctx
	client.cancelFunc = cancelFunc

	go client.stack.Start(ctx)
	go func() {
		<-client.ctx.Done()
		// 注销
		client.registrar(0)
	}()

	if err := client.registrar(-1); err != nil {
		cancelFunc()
		return err
	}
	return nil
}

// 暂时未做认证
func (client *Client) registrar(expire int) error {
	msg := message.NewRequestMessage("UDP", method.REGISTER, client.serverAddrees.Clone())
	contactParam := message.NewParams()
	if expire >= 0 {
		contactParam.Set("expires", fmt.Sprintf("%d", expire))
		msg.AppendHeader(message.NewExpiresHeader(expire))
	}
	// contactParam.Set("message-expires", "604800")

	msg.AppendHeader(
		message.NewViaHeader("UDP", client.address.Host, client.address.Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		message.NewAllowHeader(),
		message.NewCSeqHeader(5, method.REGISTER),
		message.NewFromHeader(client.displayName, client.address, message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader(client.displayName, client.address, nil),
		message.NewCallIDHeader(utils.RandString(20)),
		message.NewMaxForwardsHeader(70),
		message.NewContactHeader(client.displayName, client.address, contactParam),
		message.NewSupportedHeader([]string{"replaces", "outbound", "gruu"}),
		message.NewAcceptHeader("application/sdp"),
		message.NewAcceptHeader("text/plain"),
		message.NewAcceptHeader("application/vnd.gsma.rcs-ft-http+xml"),
	)

	err := client.Send(client.serverAddrees, msg)
	if err != nil {
		logrus.Errorf("%s registrar failed", client.user)
	}
	return nil
}

func (client *Client) HandleRequests(msg message.Request) {
	logrus.Infof("client:%s req: %s", client.user, msg.Method())

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

	logrus.Infof("client:%s method: %s response: %d  %s", client.user, cseq.Value(), msg.StatusCode(), msg.Reason())

	switch cseq.Method {
	case method.REGISTER:
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
		client.registrar(-1)
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
	callID := utils.RandString(30)
	da := client.serverAddrees.Clone()
	da.User = user
	da.Port = 0
	msg := message.NewRequestMessage("UDP", method.INVITE, da)
	msg.AppendHeader(
		message.NewViaHeader("UDP", client.address.Host, client.address.Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		message.NewAllowHeader(),
		message.NewCSeqHeader(10, method.INVITE),
		message.NewFromHeader(client.displayName, message.NewAddress(client.user, client.serverAddrees.Host, 0), message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader("", client.serverAddrees, nil),
		message.NewCallIDHeader(callID),
		message.NewMaxForwardsHeader(70),
		message.NewContactHeader(client.displayName, client.address, nil),
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
