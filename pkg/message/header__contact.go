package message

import (
	"bytes"
	"fmt"
)

func NewContactHeader(displayName string, address *Address, transport string, param *Params) *ContactHeader {
	return &ContactHeader{
		DisplayName: displayName,
		Address:     address,
		Params:      param,
		Transport:   transport,
	}
}

type ContactHeader struct {
	DisplayName string
	Address     *Address
	Params      *Params
	Transport   string
}

func (contact *ContactHeader) Name() string {
	return "Contact"
}

func (contact *ContactHeader) Value() string {
	buf := bytes.NewBuffer(nil)

	if contact.DisplayName != "" {
		buf.WriteString("\"" + contact.DisplayName + "\"" + " ")
	}

	if contact.Address != nil {
		s := contact.Address.String()
		if contact.Transport != "" {
			s += fmt.Sprintf(";transport=%s", contact.Transport)
		}

		buf.WriteString(fmt.Sprintf("<%s>", s))
	} else {
		buf.WriteString("*")
	}

	if (contact.Params != nil) && (contact.Params.Length() > 0) {
		buf.WriteString(";")
		buf.WriteString(contact.Params.ToString(";"))
	}

	return buf.String()
}

func (contact *ContactHeader) Clone() Header {
	var newCnt *ContactHeader
	if contact == nil {
		return newCnt
	}

	newCnt = &ContactHeader{
		DisplayName: contact.DisplayName,
	}
	if contact.Address != nil {
		newCnt.Address = contact.Address.Clone()
	}
	if contact.Params != nil {
		newCnt.Params = contact.Params.Clone()
	}

	return newCnt
}

func init() {
	defaultHeaderParsers.Register(&ContactHeader{})
}

func (ContactHeader) Parse(data string) (Header, error) {
	// todo <xxx>;xxx=xxx, <xxx>;xxx=xxx
	displayName, address, params, err := ParseAddressAndParam(data)
	if err != nil {
		return nil, err
	}

	return &ContactHeader{
		DisplayName: displayName,
		Address:     address,
		Params:      params,
	}, nil
}
