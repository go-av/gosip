package dialog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/utils"
	"github.com/sirupsen/logrus"
)

type Sender interface {
	Send(protocol string, address string, msg message.Message) error
}

type Dialog interface {
	Run(del func(callID string))
	SDP() []byte
	DialogID() string
	Context() context.Context
	Headers() []message.Header

	HandleResponse(msg message.Response)
	HandleRequest(req message.Request)

	State() chan State

	Answer(sdp string) error // 接收
	Reject() error           // 拒收
	Bye()                    // 挂断

	From() From
	To() To
}

type State interface {
	State() DialogState
	Reason() string
}

type stateWithReason struct {
	state  DialogState
	reason string
}

func (c *stateWithReason) State() DialogState {
	return c.state
}

func (c *stateWithReason) Reason() string {
	return c.reason
}

type OriginType string

const (
	CallIN  OriginType = "IN"
	CallOUT OriginType = "OUT"
)

type dialog struct {
	origin OriginType // 来源方向
	sender Sender
	from   From
	to     To

	dialogID string
	branchID string

	ctx    context.Context
	cancel context.CancelFunc // 用于取消和关闭

	timer *time.Timer

	sdp    []byte // 对方的 sdp
	invite message.Message

	currentstate *stateWithReason // 当前状态
	statechan    chan State       // 状态变更
}

func newDialog(ctx context.Context, origin OriginType, sender Sender, from From, to To) *dialog {
	logrus.Infof("%s call %s", from.User(), to.User())
	ctx, cancel := context.WithCancel(ctx)
	dl := &dialog{
		origin: origin,
		sender: sender,
		from:   from,
		to:     to,
		ctx:    ctx,
		cancel: cancel,
		timer:  time.NewTimer(20 * time.Second),
		currentstate: &stateWithReason{
			state: Proceeding,
		},
		statechan: make(chan State, 1),
	}
	return dl
}

func (dl *dialog) DialogID() string {
	return dl.dialogID
}

func (dl *dialog) Context() context.Context {
	return dl.ctx
}

func (dl *dialog) From() From {
	return dl.from
}

func (dl *dialog) To() To {
	return dl.to
}

func (dl *dialog) Headers() []message.Header {
	return dl.invite.Headers()
}

// 状态变化通知
func (dl *dialog) State() chan State {
	return dl.statechan
}

func (dl *dialog) CurrentState() State {
	return dl.currentstate
}

// 对方的 SDP
func (dl *dialog) SDP() []byte {
	return dl.sdp
}

// 呼叫
func Invite(ctx context.Context, sender Sender, from From, to To, sdp []byte, updateMsg func(message.Message)) (Dialog, error) {
	dl := newDialog(ctx, CallOUT, sender, from, to)
	dl.dialogID = utils.RandString(30)
	dl.branchID = utils.GenerateBranchID()

	fromAddress := message.NewAddress("", dl.from.HostAndPort().Host, dl.from.HostAndPort().Port)
	toAddress := message.NewAddress("", dl.to.HostAndPort().Host, dl.to.HostAndPort().Port)
	msg := message.NewRequestMessage(dl.from.Protocol(), method.INVITE, toAddress.Clone().SetUser(dl.to.User()))
	msg.AppendHeader(
		message.NewViaHeader(dl.from.Protocol(), toAddress.Host, toAddress.Port, message.NewParams().Set("branch", dl.branchID).Set("rport", "")),
		message.NewFromHeader("", toAddress.Clone().SetUser(dl.from.User()), message.NewParams().Set("tag", utils.RandString(20))),
		message.NewToHeader("", toAddress.Clone().SetUser(dl.to.User()), nil),
		message.NewContactHeader("", fromAddress.Clone().SetUser(dl.from.User()), dl.from.Protocol(), message.NewParams().Set("expires", "4800")),
		message.NewAllowHeader(),
		message.NewCSeqHeader(10, method.INVITE),
		message.NewMaxForwardsHeader(70),
		message.NewCallIDHeader(dl.dialogID),
	)

	msg.SetBody(string(message.ContentType__SDP), sdp)
	if updateMsg != nil {
		updateMsg(msg)
	}
	dl.invite = msg
	err := dl.sender.Send(dl.from.Protocol(), dl.to.HostAndPort().String(), msg)
	if err != nil {
		return nil, err
	}
	return dl, nil
}

// 接收
func Receive(sender Sender, from From, to To, callID string, msg message.Request) (Dialog, error) {
	dl := newDialog(context.Background(), CallIN, sender, from, to)
	dl.dialogID = callID
	dl.sdp = msg.Body()
	dl.invite = msg
	// todo bandid
	resp := message.NewResponse(dl.invite, 180, "ok")
	resp.SetHeader(message.NewRecordRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.from.HostAndPort().Host)))
	resp.SetHeader(message.NewContactHeader("", message.NewAddress(dl.to.User(), dl.to.HostAndPort().Host, dl.to.HostAndPort().Port), dl.from.Protocol(), nil))
	err := dl.sender.Send(dl.from.Protocol(), dl.from.HostAndPort().String(), resp)
	if err != nil {
		return nil, err
	}

	dl.timer.Reset(30 * time.Second)
	return dl, nil
}

func (dl *dialog) HandleResponse(resp message.Response) {
	cseq, ok := resp.CSeq()
	if !ok {
		return
	}
	switch dl.origin {
	case CallIN:

	case CallOUT:
		switch cseq.Method {
		case method.INVITE:
			switch resp.StatusCode() {
			case 100:
				to, _ := resp.To()
				dl.invite.SetHeader(to)
				dl.updateState(Trying, Trying.String())
			case 180:
				to, _ := resp.To()
				dl.invite.SetHeader(to)
				dl.updateState(Ringing, Ringing.String())
			case 200:
				to, _ := resp.To()
				dl.invite.SetHeader(to)
				dl.timer.Stop()
				dl.sdp = resp.Body()

				contact, _ := resp.Contact()
				dl.invite.SetHeader(contact)
				toAddress := message.NewAddress("", contact.Address.Host, contact.Address.Port)

				req := message.NewRequestMessage(dl.from.Protocol(), method.ACK, toAddress.Clone().SetUser(dl.to.User()))
				message.CopyHeaders(dl.invite, req, "Max-Forwards", "Call-ID", "From", "To")
				req.AppendHeader(
					message.NewViaHeader(dl.from.Protocol(), contact.Address.Host, contact.Address.Port, message.NewParams().Set("branch", dl.branchID).Set("rport", "")),
					message.NewCSeqHeader(cseq.SeqNo, method.ACK),
					message.NewRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.to.HostAndPort().String())),
				)

				err := dl.sender.Send(dl.from.Protocol(), dl.to.HostAndPort().String(), req)
				if err != nil {
					logrus.Error(err)
				}

				dl.updateState(Accepted, Accepted.String())
			default:
				// status code < 400
				if resp.StatusCode() < 400 {
					return
				}

				// 收到错误信息时，返回 ACK.
				fromAddress := message.NewAddress("", dl.from.HostAndPort().Host, dl.from.HostAndPort().Port)
				req := message.NewRequestMessage(dl.from.Protocol(), method.ACK, fromAddress.Clone().SetUser(dl.to.User()))
				message.CopyHeaders(resp, req, "Via", "Call-ID", "From", "To")
				req.AppendHeader(
					message.NewMaxForwardsHeader(70),
					message.NewCSeqHeader(10, method.ACK),
				)
				err := dl.sender.Send(dl.from.Protocol(), dl.to.HostAndPort().String(), req)
				if err != nil {
					logrus.Error(err)
				}
				msg := resp.Reason()
				vals := resp.GetHeaders("Warning")
				if vals != nil {
					warningHeader, ok := vals[0].(*message.WarningHeader)
					if ok {
						msg += "(" + warningHeader.Value() + ")"
					}
				}
				dl.updateState(Error, fmt.Sprintf("code:%d msg:%s", resp.StatusCode(), msg))
				dl.cancel()
			}

		case method.ACK:

		case method.BYE:
			switch resp.StatusCode() {
			case 200:
				dl.cancel()
				return
			}
		}
	}
}

// todo 接听时 10次，200 指导收到 ack
func (dl *dialog) HandleRequest(req message.Request) {
	switch req.Method() {
	case method.BYE:
		var addr *utils.HostAndPort
		if dl.origin == CallOUT {
			addr = dl.to.HostAndPort()
		} else {
			addr = dl.from.HostAndPort()
		}
		resp := message.NewResponse(req, 200, "success.")
		resp.SetHeader(
			message.NewViaHeader(dl.from.Protocol(), addr.Host, addr.Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
		)

		resp.SetHeader(message.NewContactHeader("", message.NewAddress(dl.to.User(), dl.to.HostAndPort().Host, dl.to.HostAndPort().Port), dl.from.Protocol(), nil))
		err := dl.sender.Send(dl.from.Protocol(), addr.String(), resp)
		if err != nil {
			logrus.Error(err)
		}

		dl.cancel()
	case method.CANCEL:
		resp := message.NewResponse(req, 200, "success.")
		err := dl.sender.Send(dl.from.Protocol(), dl.to.HostAndPort().String(), resp)
		if err != nil {
			logrus.Error(err)
		}

		dl.cancel()
	case method.ACK:
		dl.updateState(Accepted, Accepted.String())
		dl.timer.Stop()
		resp := message.NewResponse(req, 200, "success.")
		addr := dl.to.HostAndPort().String()
		if dl.origin == CallIN {
			addr = dl.from.HostAndPort().String()
		}
		err := dl.sender.Send(dl.from.Protocol(), addr, resp)
		if err != nil {
			logrus.Error(err)
		}
	default:
		logrus.Debugf("收到的%s消息未处理", req.Method())
	}
}

func (dl *dialog) updateState(state DialogState, reason string) {
	if state == Accepted {
		dl.timer.Stop()
	} else {
		dl.timer.Reset(10 * time.Second)
	}
	if state != dl.currentstate.state {
		sr := &stateWithReason{
			state:  state,
			reason: reason,
		}
		dl.currentstate = sr
		dl.statechan <- sr
	}
}

func (dl *dialog) Run(del func(callID string)) {
	defer func() {
		logrus.Infof("%s call %s 结束", dl.from.User(), dl.to.User())
		del(dl.dialogID)
	}()

	for {
		select {
		case <-dl.timer.C:
			if dl.origin == CallOUT && dl.currentstate.state < Accepted {
				// 取消请求
				fromAddress := message.NewAddress("", dl.from.HostAndPort().Host, dl.from.HostAndPort().Port)
				req := message.NewRequestMessage(dl.from.Protocol(), method.CANCEL, fromAddress.Clone().SetUser(dl.to.User()))
				message.CopyHeaders(dl.invite, req, "Via", "Call-ID", "From", "To")
				req.AppendHeader(
					message.NewMaxForwardsHeader(70),
					message.NewCSeqHeader(10, method.ACK),
				)
				err := dl.sender.Send(dl.from.Protocol(), dl.to.HostAndPort().String(), req)
				if err != nil {
					logrus.Error(err)
				}

				if err != nil {
					logrus.Error(err)
				}
			}
			if dl.origin == CallIN && dl.currentstate.state < Accepted {
				resp := message.NewResponse(dl.invite, 486, "Busy Here")
				resp.SetHeader(message.NewContactHeader("", message.NewAddress(dl.to.User(), dl.to.HostAndPort().Host, dl.to.HostAndPort().Port), dl.from.Protocol(), nil))
				err := dl.sender.Send(dl.from.Protocol(), dl.from.HostAndPort().String(), resp)
				if err != nil {
					logrus.Error(err)
				}
			}
			dl.cancel()
		case <-dl.ctx.Done():
			return
		}
	}
}

// 接收
// todo  200或错误返回，直到 ack 截止（有效时间内）
func (dl *dialog) Answer(sdp string) error {
	if dl.origin != CallIN {
		return errors.New("非法操作")
	}
	dl.timer.Reset(10 * time.Second)

	resp := message.NewResponse(dl.invite, 200, "success.")
	resp.SetBody(string(message.ContentType__SDP), []byte(sdp))
	resp.SetHeader(message.NewContactHeader("", message.NewAddress(dl.to.User(), dl.to.HostAndPort().Host, dl.to.HostAndPort().Port), dl.from.Protocol(), nil))
	resp.SetHeader(message.NewRecordRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.from.HostAndPort().Host)))

	err := dl.sender.Send(dl.from.Protocol(), dl.from.HostAndPort().String(), resp)
	if err != nil {
		logrus.Error(err)
		return err
	}

	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

// 拒绝接收
func (dl *dialog) Reject() error {
	if dl.origin != CallIN {
		return errors.New("非法操作")
	}
	dl.timer.Reset(10 * time.Second)
	resp := message.NewResponse(dl.invite, 603, "Decline")
	err := dl.sender.Send(dl.from.Protocol(), dl.from.HostAndPort().String(), resp)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

func (dl *dialog) Bye() {
	dl.timer.Reset(10 * time.Second)
	if dl.origin == CallOUT {
		contact, _ := dl.invite.Contact()
		toAddress := message.NewAddress(dl.to.User(), contact.Address.Host, contact.Address.Port)
		req := message.NewRequestMessage(dl.from.Protocol(), method.BYE, toAddress)
		message.CopyHeaders(dl.invite, req, "Max-Forwards", "Call-ID", "From", "To")
		req.AppendHeader(
			message.NewViaHeader(dl.from.Protocol(), contact.Address.Host, contact.Address.Port, message.NewParams().Set("branch", dl.branchID).Set("rport", "")),
			message.NewCSeqHeader(12, method.BYE),
			message.NewRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.to.HostAndPort().String())),
		)
		err := dl.sender.Send(dl.from.Protocol(), dl.to.HostAndPort().String(), req)
		if err != nil {
			logrus.Error(err)
		}
		return
	}

	if dl.origin == CallIN {
		contact, _ := dl.invite.Contact()

		toAddress := message.NewAddress(dl.from.User(), contact.Address.Host, contact.Address.Port)
		req := message.NewRequestMessage(dl.from.Protocol(), method.BYE, toAddress)
		message.CopyHeaders(dl.invite, req, "Max-Forwards", "Call-ID", "From", "To")
		req.AppendHeader(
			message.NewViaHeader(dl.from.Protocol(), dl.From().HostAndPort().Host, dl.From().HostAndPort().Port, message.NewParams().Set("branch", utils.GenerateBranchID()).Set("rport", "")),
			message.NewCSeqHeader(21, method.BYE),
			message.NewRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.from.HostAndPort().String())),
		)

		to, _ := dl.invite.To()
		from, _ := dl.invite.From()

		req.SetHeader(message.NewFromHeader(to.DisplayName, to.Address.Clone(), to.Params))
		req.SetHeader(message.NewToHeader(from.DisplayName, from.Address.Clone(), from.Params))

		err := dl.sender.Send(dl.from.Protocol(), dl.from.HostAndPort().String(), req)
		if err != nil {
			logrus.Error(err)
		}
	}
}
