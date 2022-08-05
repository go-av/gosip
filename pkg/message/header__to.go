package message

import (
	"bytes"
	"fmt"
)

func NewToHeader(displayName string, address *Address, param *Params) *ToHeader {
	return &ToHeader{
		DisplayName: displayName,
		Address:     address,
		Params:      param,
	}
}

type ToHeader struct {
	DisplayName string
	Address     *Address
	Params      *Params
}

func (ToHeader) Name() string {
	return "To"
}

func (to *ToHeader) Value() string {
	buf := bytes.NewBuffer(nil)
	if to.DisplayName != "" {
		buf.WriteString(fmt.Sprintf("\"%s\" ", to.DisplayName))
	}

	if to.Address != nil {
		buf.WriteString(fmt.Sprintf("<%s>", to.Address.String()))
	}

	if to.Params != nil && to.Params.Length() > 0 {
		buf.WriteString(";")
		buf.WriteString(to.Params.ToString(";"))
	}

	return buf.String()
}

func (to *ToHeader) Clone() Header {
	return &ToHeader{
		DisplayName: to.DisplayName,
		Address:     to.Address,
		Params:      to.Params,
	}
}

func init() {
	defaultHeaderParsers.Register(&ToHeader{})
}

func (ToHeader) Parse(data string) (Header, error) {
	displayName, address, param, err := ParseAddressAndParam(data)
	if err != nil {
		return nil, err
	}

	return &ToHeader{
		DisplayName: displayName,
		Address:     address,
		Params:      param,
	}, nil
}
