package dialog

import (
	"fmt"

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
			fmt.Println("dl.state", state)
			switch state {
			case Ringing:
				resp := message.NewResponse(dl.invite, 180, "Ringing")
				err := dl.client.Send(dl.client.Address(), resp)
				if err != nil {
					logrus.Error(err)
					fmt.Println(err)
				}
			case Answered:
				// to, _ := dl.invite.To()
				// if to.Params != nil {
				// 	to.Params = message.NewParams()
				// }
				// if _, ok := to.Params.Get("tag"); !ok {
				// 	to.Params.Set("tag", "")
				// }
				// dl.invite.SetHeader(to)
				// from, _ := dl.invite.From()
				resp := message.NewResponse(dl.invite, 200, "Ok")
				// resp.SetHeader(message.NewFromHeader(to.DisplayName, to.Address, to.Params))
				// resp.SetHeader(message.NewToHeader(from.DisplayName, from.Address, from.Params))

				resp.SetHeader(message.NewRecordRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.client.Address().Host)))
				resp.SetBody(dl.client.SDP())
				err := dl.client.Send(dl.client.Address(), resp)
				if err != nil {
					logrus.Error(err)
					fmt.Println(err)
				}

			case Missed:
				resp := message.NewResponse(dl.invite, 480, "Missed")
				err := dl.client.Send(dl.client.Address(), resp)
				if err != nil {
					logrus.Error(err)
					fmt.Println(err)
				}
			case Hangup:
				contact, _ := dl.invite.Contact()
				ss := contact.Address.Clone()
				byeReq := message.NewRequestMessage("UDP", method.BYE, ss)
				message.CopyHeaders(dl.invite, byeReq, "Call-ID", "Via", "Max-Forwards")
				byeReq.SetHeader(message.NewCSeqHeader(12, method.BYE))
				from, _ := dl.invite.From()
				to, _ := dl.invite.To()
				byeReq.SetHeader(message.NewFromHeader(to.DisplayName, to.Address.Clone(), to.Params.Clone()))
				byeReq.SetHeader(message.NewToHeader(from.DisplayName, from.Address.Clone(), from.Params.Clone()))
				byeReq.SetHeader(message.NewRecordRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.client.Address().Host)))
				err := dl.client.Send(dl.client.Address(), byeReq)
				if err != nil {
					fmt.Println(err)
				}
			}

		case msg := <-dl.msgs:
			if req, ok := msg.(message.Request); ok {
				fmt.Println("req.Method()", req.Method())
				switch req.Method() {
				case method.INVITE:
					resp := message.NewResponse(msg, 100, "Trying")
					err := dl.client.Send(dl.client.Address(), resp)
					if err != nil {
						logrus.Error(err)
						fmt.Println(err)
					}
				case method.ACK:
					resp := message.NewResponse(msg, 200, "ok")

					err := dl.client.Send(dl.client.Address(), resp)
					if err != nil {
						logrus.Error(err)
						fmt.Println(err)
					}
					dl.fromstate <- Answered
				case method.BYE:
					resp := message.NewResponse(msg, 200, "ok")
					err := dl.client.Send(dl.client.Address(), resp)
					if err != nil {
						logrus.Error(err)
						fmt.Println(err)
					}
					dl.fromstate <- Hangup
					return
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
