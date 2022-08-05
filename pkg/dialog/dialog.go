package dialog

import (
	"sync"
	"time"

	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/sdp"
)

type Dialog interface {
	User() (displayName string, user string)
	State() chan DialogState
	SetState(DialogState) error
	SDP() *sdp.SDP
	Hangup()
	WriteMsg(message.Message)
	// 当产生错误时，可通过 Reason() StatusCode() 获取原因
	Reason() string
	StatusCode() message.StatusCode
}

type Client interface {
	User() string
	Send(address *message.Address, msg message.Message) error
	Address() *message.Address
	SDP() *sdp.SDP
}

func NewDialogManger(client Client) *DialogManger {
	return &DialogManger{
		client: client,
	}
}

type DialogManger struct {
	dialogs sync.Map
	client  Client
}

func (mgr *DialogManger) remove(callID string) {
	mgr.dialogs.Delete(callID)
}

func (mgr *DialogManger) HandleMessage(msg message.Message) Dialog {
	callID, ok := msg.CallID()
	if !ok {
		return nil
	}

	if dl, ok := mgr.dialogs.Load(callID.Value()); ok {
		dl.(Dialog).WriteMsg(msg)
		return nil
	}

	cseq, ok := msg.CSeq()
	if !ok {
		return nil
	}

	if cseq.Method != method.INVITE {
		return nil
	}
	from, _ := msg.From()
	if from.Address.User == mgr.client.User() {
		dl := &callOutDialog{
			client: mgr.client,
			callID: callID.Value(),
			timer:  time.NewTimer(20 * time.Second),
			msgs:   make(chan message.Message, 10),
			state:  make(chan DialogState, 4),
			invite: msg,
			hangup: make(chan bool, 4),
		}
		mgr.dialogs.Store(callID.Value(), dl)
		go dl.run(mgr)
		dl.state <- Ringing
		dl.msgs <- msg
		return dl
	}

	dl := &callInDialog{
		client: mgr.client,
		callID: callID.Value(),

		msgs:      make(chan message.Message, 10),
		tostate:   make(chan DialogState, 3),
		fromstate: make(chan DialogState, 3),

		invite: msg,

		sdp: msg.Body().(*sdp.SDP),

		displayName: from.DisplayName,
		user:        from.Address.User,
	}

	mgr.dialogs.Store(callID.Value(), dl)
	go dl.run(mgr)
	dl.WriteMsg(msg)
	return dl
}

type manager interface {
	remove(callID string)
}
