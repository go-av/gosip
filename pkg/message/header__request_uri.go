package message

import (
	"bytes"
	"fmt"
)

func NewRequestURIHeader(displayName string, address *Address, transport string, param *Params) *RequestURIHeader {
	return &RequestURIHeader{
		DisplayName: displayName,
		Address:     address,
		Params:      param,
		Transport:   transport,
	}
}

type RequestURIHeader struct {
	DisplayName string
	Address     *Address
	Params      *Params
	Transport   string
}

func (contact *RequestURIHeader) Name() string {
	return "Request-URI"
}

func (req *RequestURIHeader) Value() string {
	buf := bytes.NewBuffer(nil)

	if req.DisplayName != "" {
		buf.WriteString("\"" + req.DisplayName + "\"" + " ")
	}

	if req.Address != nil {
		s := req.Address.String()
		if req.Transport != "" {
			s += fmt.Sprintf(";transport=%s", req.Transport)
		}

		buf.WriteString(fmt.Sprintf("<%s>", s))
	} else {
		buf.WriteString("*")
	}

	if (req.Params != nil) && (req.Params.Length() > 0) {
		buf.WriteString(";")
		buf.WriteString(req.Params.ToString(";"))
	}

	return buf.String()
}

func (req *RequestURIHeader) Clone() Header {
	var newCnt *RequestURIHeader
	if req == nil {
		return newCnt
	}

	newCnt = &RequestURIHeader{
		DisplayName: req.DisplayName,
	}
	if req.Address != nil {
		newCnt.Address = req.Address.Clone()
	}
	if req.Params != nil {
		newCnt.Params = req.Params.Clone()
	}

	return newCnt
}

func init() {
	defaultHeaderParsers.Register(&RequestURIHeader{})
}

func (RequestURIHeader) Parse(data string) (Header, error) {
	// todo <xxx>;xxx=xxx, <xxx>;xxx=xxx
	displayName, address, params, err := ParseAddressAndParam(data)
	if err != nil {
		return nil, err
	}

	return &RequestURIHeader{
		DisplayName: displayName,
		Address:     address,
		Params:      params,
	}, nil
}
