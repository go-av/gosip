package message

import (
	"bytes"
	"fmt"
)

func NewFromHeader(displayName string, address *Address, param *Params) *FromHeader {
	return &FromHeader{
		DisplayName: displayName,
		Address:     address,
		Params:      param,
	}
}

type FromHeader struct {
	DisplayName string
	Address     *Address
	Params      *Params
}

func (from FromHeader) Name() string {
	return "From"
}

func (from *FromHeader) Value() string {
	buf := bytes.NewBuffer(nil)
	if from.DisplayName != "" {
		buf.WriteString(fmt.Sprintf("\"%s\" ", from.DisplayName))
	}

	if from.Address != nil {
		buf.WriteString(fmt.Sprintf("<%s>", from.Address.String()))
	}

	if from.Params != nil && from.Params.Length() > 0 {
		buf.WriteString(";")
		buf.WriteString(from.Params.ToString(";"))
	}

	return buf.String()
}

func (from *FromHeader) Clone() Header {
	return &FromHeader{
		DisplayName: from.DisplayName,
		Address:     from.Address,
		Params:      from.Params,
	}
}

func init() {
	defaultHeaderParsers.Register(&FromHeader{})
}

func (FromHeader) Parse(data string) (Header, error) {
	displayName, address, param, err := ParseAddressAndParam(data)
	if err != nil {
		return nil, err
	}

	return &FromHeader{
		DisplayName: displayName,
		Address:     address,
		Params:      param,
	}, nil
}
