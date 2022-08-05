package message

import (
	"bytes"
	"fmt"

	"github.com/go-av/gosip/pkg/method"
)

type Request interface {
	Message
	Method() method.Method
}

func NewRequestMessage(transport string, method method.Method, recipient *Address) Message {
	req := new(request)
	req.headers = &headers{
		headers: map[string][]Header{},
	}
	req.transport = transport
	req.method = method
	req.recipient = recipient
	req.message.startLine = req.StartLine
	// req.body = body

	req.SetHeader(NewUserAgentHeader("gosip"))

	return req
}

type request struct {
	message
	transport string
	method    method.Method
	recipient *Address
}

func (req *request) StartLine() string {
	buffer := bytes.NewBuffer(nil)

	buffer.WriteString(
		fmt.Sprintf(
			"%s %s %s",
			string(req.Method()),
			req.Recipient(),
			"SIP/2.0",
		),
	)

	return buffer.String()
}

func (req *request) Method() method.Method {
	return req.method
}

func (req *request) Recipient() *Address {
	return req.recipient
}
