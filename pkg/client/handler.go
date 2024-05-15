package client

import (
	"context"

	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/method"
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

func (r *Response) WithHeader(header message.Headers) *Response {
	r.header = header
	return r
}

type Response struct {
	code        message.StatusCode
	reason      string
	body        []byte
	contentType message.ContentType
	header      message.Headers
}

type Handler interface {
	ReceiveMessage(context.Context, method.Method, message.Headers, message.Body) (*Response, error)
}
