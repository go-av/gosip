package server

import (
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
	SetServer(Server)
	GetClient(user string) (Client, error)
	Realm() string
	ReceiveMessage(message.Body) (*Response, error)
}

type Server interface {
	Send(protocol string, address string, msg message.Message) error
	SendMessage(client Client, content message.Body) (message.Body, error)
	ServerAddress() *message.Address
}
