package dialog

import (
	"fmt"
	"time"

	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/sdp"
	"github.com/sirupsen/logrus"
)

// 呼出
type callOutDialog struct {
	client Client
	state  chan DialogState
	invite message.Message
	callID string
	timer  *time.Timer
	msgs   chan message.Message
	sdp    *sdp.SDP

	hangup chan bool // 挂断

	reason     string
	statusCode message.StatusCode
}

func (dl *callOutDialog) run(mgr manager) {
	defer func() {
		mgr.remove(dl.callID)
	}()
	for {
		select {
		case <-dl.hangup:
			to, _ := dl.invite.Contact()
			ss := to.Address.Clone()
			byeReq := message.NewRequestMessage("UDP", method.BYE, ss)
			message.CopyHeaders(dl.invite, byeReq, "Call-ID", "Via", "From", "To", "Max-Forwards")
			byeReq.SetHeader(message.NewCSeqHeader(12, method.BYE))
			byeReq.SetHeader(message.NewRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.client.Address().Host)))
			err := dl.client.Send(dl.client.Address(), byeReq)
			if err != nil {
				logrus.Error(err)
			}
			dl.state <- Hangup
		case <-dl.timer.C:
			dl.state <- Missed
			dl.hangup <- true
		case msg := <-dl.msgs:
			if resp, ok := msg.(message.Response); ok {
				dl.reason = resp.Reason()
				dl.statusCode = resp.StatusCode()
				switch resp.StatusCode() {
				case 100:
					dl.state <- Proceeding
				case 180:
					dl.state <- Ringing
				case 200:
					cseq, _ := resp.CSeq()
					switch cseq.Method {
					case method.INVITE:
						if sd, ok := resp.Body().(*sdp.SDP); ok {
							dl.sdp = sd
						}
						dl.timer.Stop()
						con, _ := resp.Contact()
						newResp := message.NewRequestMessage("UDP", method.ACK, con.Address)
						message.CopyHeaders(msg, newResp, "Call-ID", "Via", "From", "To", "CSeq", "Max-Forwards")
						newResp.SetHeader(message.NewCSeqHeader(1, method.ACK))
						newResp.SetHeader(message.NewRouteHeader(fmt.Sprintf("<sip:%s;lr>", dl.client.Address().Host)))
						err := dl.client.Send(dl.client.Address(), newResp)

						if err != nil {
							logrus.Error("err", err)
							continue
						}

						form, _ := msg.From()
						to, _ := msg.To()

						dl.invite.SetHeader(form.Clone())
						dl.invite.SetHeader(to.Clone())
						dl.invite.SetHeader(con)

						dl.state <- Answered
					case method.BYE:
						dl.state <- Hangup
						return
					}
				default:
					dl.state <- Error
					return
				}
			}
			if req, ok := msg.(message.Request); ok {
				cseq, _ := req.CSeq()
				switch cseq.Method {
				case method.BYE:
					resp := message.NewResponse(req, 200, "Ok")
					err := dl.client.Send(dl.client.Address(), resp)
					if err != nil {
						logrus.Error(err)
					}
					logrus.Debug("收到退出指令")
					dl.state <- Hangup
					return
				}
			}
		}
	}
}

func (dl *callOutDialog) User() (displayName string, user string) {
	return "", ""
}

func (dl *callOutDialog) State() chan DialogState {
	return dl.state
}

func (dl *callOutDialog) SDP() *sdp.SDP {
	return dl.sdp
}

func (dl *callOutDialog) CallID() string {
	return dl.callID
}

func (dl *callOutDialog) Hangup() {
	dl.hangup <- true
}

func (dl *callOutDialog) WriteMsg(msg message.Message) {
	dl.msgs <- msg
}

func (dl *callOutDialog) SetState(DialogState) error {
	return nil
}

func (res *callOutDialog) Reason() string {
	return res.reason
}

func (res *callOutDialog) StatusCode() message.StatusCode {
	return res.statusCode
}
