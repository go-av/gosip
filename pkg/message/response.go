package message

import (
	"bytes"
	"fmt"

	"github.com/go-av/gosip/pkg/method"
)

type StatusCode int

type Response interface {
	Message

	StatusCode() StatusCode
	Reason() string
	IsProvisional() bool

	IsSuccess() bool
	IsRedirection() bool
	IsClientError() bool
	IsServerError() bool
	IsGlobalError() bool

	IsAck() bool
	IsCancel() bool
}

func NewResponse(req Message, statusCode StatusCode, reason string) Response {
	resp := new(response)
	resp.startLine = resp.StartLine
	resp.statusCode = statusCode
	resp.reason = reason
	resp.headers = &headers{
		headers: map[string][]Header{},
	}

	if req != nil {
		CopyHeaders(req, resp, "Via", "Call-ID", "From", "To", "CSeq", "Max-Forwards")
		resp.SetHeader(NewUserAgentHeader(userAgent))
	}

	return resp
}

type response struct {
	message
	statusCode StatusCode
	reason     string
}

func (res *response) StartLine() string {
	buffer := bytes.NewBuffer(nil)

	buffer.WriteString(
		fmt.Sprintf(
			"%s %d %s",
			"SIP/2.0",
			res.StatusCode(),
			res.Reason(),
		),
	)
	return buffer.String()
}

func (res *response) StatusCode() StatusCode {
	return res.statusCode
}

func (res *response) Reason() string {
	return res.reason
}

func (res *response) IsProvisional() bool {
	return res.StatusCode() < 200
}

func (res *response) IsSuccess() bool {
	return res.StatusCode() >= 200 && res.StatusCode() < 300
}

func (res *response) IsRedirection() bool {
	return res.StatusCode() >= 300 && res.StatusCode() < 400
}

func (res *response) IsClientError() bool {
	return res.StatusCode() >= 400 && res.StatusCode() < 500
}

func (res *response) IsServerError() bool {
	return res.StatusCode() >= 500 && res.StatusCode() < 600
}

func (res *response) IsGlobalError() bool {
	return res.StatusCode() >= 600
}

func (res *response) IsAck() bool {
	if cseq, ok := res.CSeq(); ok {
		return cseq.Method == method.ACK
	}
	return false
}

func (res *response) IsCancel() bool {
	if cseq, ok := res.CSeq(); ok {
		return cseq.Method == method.CANCEL
	}
	return false
}
