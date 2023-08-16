package server

import (
	"context"

	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/message"
)

func NewResponse(code int, reason string) *Response {
	return &Response{
		code:   message.StatusCode(code),
		reason: reason,
	}
}

func (r *Response) WithBody(contentType message.ContentType, body []byte) *Response {
	r.contentType = contentType
	r.body = body
	return r
}

type Response struct {
	code        message.StatusCode
	reason      string
	body        []byte
	contentType message.ContentType
}

type Handler interface {
	GetClient(deviceID string) (Client, error)
	Realm() string
	ReceiveMessage(context.Context, Client, message.Body) (*Response, error)
}

type Server interface {
	Send(protocol string, address string, msg message.Message) error
	SendMessage(client Client, req message.Request) (message.Response, error)
	ServerAddress() *message.Address
	Invite(ctx context.Context, from dialog.From, to dialog.To, sdp string, updateMsg func(msg message.Message)) (dialog.Dialog, error) // 呼出
	Receive() chan dialog.Dialog                                                                                                        // 接收呼叫
}
