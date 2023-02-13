package dialog

import (
	"fmt"
	"strings"

	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/sdp"
	"github.com/sirupsen/logrus"
)

// 呼入
type callInDialog struct {
	client    Client
	tostate   chan DialogState
	fromstate chan DialogState

	invite message.Message
	callID string
	msgs   chan message.Message
	sdp    *sdp.SDP

	reason     string
	statusCode message.StatusCode

	user        string
	displayName string
}

func (dl *callInDialog) run(mgr manager) {
	defer func() {
		mgr.remove(dl.callID)
	}()

	for {
		select {
		case state := <-dl.tostate:
			switch state {
			case Ringing:
				resp := message.NewResponse(dl.invite, 180, "Ringing")
				resp.SetHeader(message.NewRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.client.Address().String())))
				resp.SetHeader(message.NewContactHeader("", dl.invite.(message.Request).Recipient(), dl.client.Protocol(), nil))
				err := dl.client.Send(dl.client.Address().String(), resp)
				if err != nil {
					logrus.Error(err)
				}
			case Answered:
				resp := message.NewResponse(dl.invite, 200, "Ok")
				resp.SetHeader(message.NewRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.client.Address().String())))
				resp.SetHeader(message.NewContactHeader("", dl.invite.(message.Request).Recipient(), dl.client.Protocol(), nil))
				resp.SetBody(dl.client.SDP(dl.sdp))
				err := dl.client.Send(dl.client.Address().String(), resp)
				if err != nil {
					logrus.Error(err)
				}

			case Missed:
				resp := message.NewResponse(dl.invite, 480, "Missed")
				err := dl.client.Send(dl.client.Address().String(), resp)
				if err != nil {
					logrus.Error(err)
				}
			case Hangup:
				con, _ := dl.invite.Contact()
				byeReq := message.NewRequestMessage(strings.ToUpper(dl.client.Protocol()), method.BYE, con.Address)
				message.CopyHeaders(dl.invite, byeReq, "Call-ID", "Via", "From", "To", "Max-Forwards")
				byeReq.SetHeader(message.NewCSeqHeader(1, method.BYE))
				err := dl.client.Send(dl.client.Address().String(), byeReq)
				if err != nil {
					logrus.Error(err)
				}
				dl.fromstate <- Hangup
			}

		case msg := <-dl.msgs:
			if req, ok := msg.(message.Request); ok {
				switch req.Method() {
				case method.INVITE:
					resp := message.NewResponse(msg, 100, "Trying")
					resp.SetHeader(message.NewRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.client.Address().String())))
					resp.DelHeader("Contact")
					err := dl.client.Send(dl.client.Address().String(), resp)
					if err != nil {
						logrus.Error(err)
					}
				case method.ACK:
					resp := message.NewResponse(msg, 200, "ok")
					err := dl.client.Send(dl.client.Address().String(), resp)
					if err != nil {
						logrus.Error(err)
					}
					dl.fromstate <- Answered
				case method.BYE:
					resp := message.NewResponse(msg, 200, "ok")
					err := dl.client.Send(dl.client.Address().String(), resp)
					if err != nil {
						logrus.Error(err)
					}
					dl.fromstate <- Hangup
					return
				case method.CANCEL:
					dl.fromstate <- Hangup
				}
			}

			if resp, ok := msg.(message.Response); ok {
				dl.statusCode = resp.StatusCode()
				dl.reason = resp.Reason()
				switch resp.StatusCode() {
				case 200:
					cseq, _ := resp.CSeq()
					switch cseq.Method {
					case method.BYE:
						dl.fromstate <- Hangup
					}
				}

			}
		}
	}
}

func (dl *callInDialog) User() (displayName string, user string) {
	return dl.displayName, dl.user
}

func (dl *callInDialog) State() chan DialogState {
	return dl.fromstate
}

func (dl *callInDialog) SDP() *sdp.SDP {
	return dl.sdp
}

func (dl *callInDialog) Hangup() {
	dl.tostate <- Hangup
}

func (dl *callInDialog) CallID() string {
	return dl.callID
}

func (dl *callInDialog) WriteMsg(msg message.Message) {
	dl.msgs <- msg
}

func (dl *callInDialog) SetState(state DialogState) error {
	dl.tostate <- state
	return nil
}

func (res *callInDialog) Reason() string {
	return res.reason
}

func (res *callInDialog) StatusCode() message.StatusCode {
	return res.statusCode
}
