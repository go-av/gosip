package message

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-av/gosip/pkg/types"
)

func NewAddress(user string, host string, port types.Port) *Address {
	return &Address{
		User: user,
		Host: host,
		Port: port,
	}
}

type Address struct {
	Encrypted bool
	User      string
	Host      string
	Port      types.Port
}

func (s *Address) String() string {
	buf := bytes.NewBuffer(nil)
	if s.Encrypted {
		buf.WriteString("sips")
		buf.WriteString(":")
	} else {
		buf.WriteString("sip")
		buf.WriteString(":")
	}
	buf.WriteString(s.User)
	buf.WriteString("@")
	buf.WriteString(s.Host)
	if s.Port > 0 {
		buf.WriteString(":" + s.Port.String())
	}
	return buf.String()
}

func (address *Address) Clone() *Address {
	return &Address{
		Encrypted: address.Encrypted,
		User:      address.User,
		Host:      address.Host,
		Port:      address.Port,
	}
}

func ParseAddressAndParam(text string) (display string, address *Address, param *Params, err error) {
	param = NewParams()
	if len(text) == 0 {
		err = fmt.Errorf("address-type header has empty body")
		return
	}

	text = strings.TrimSpace(text)
	index := strings.Index(text, ";")

	if index == -1 {
		display, address, err = ParamAddress(text)
	} else {
		display, address, err = ParamAddress(text[:index])
		param = ParseParams(text[index+1:])
	}
	return
}

func ParamAddress(text string) (display string, address *Address, err error) {
	text = strings.TrimSpace(text)
	address = &Address{}
	i := strings.Index(text, "<")
	if i > 1 {
		display = strings.TrimSpace(text[:i])
		text = text[i+1:]
	}
	if i := strings.Index(text, ">"); i != -1 {
		text = text[:i]
	}

	if display != "0" {
		display = strings.Replace(display, "\"", "", -1)
	}

	// parse uri
	colonIdx := strings.Index(text, ":")
	if colonIdx == -1 {
		err = fmt.Errorf("no ':' in URI %s", text)
		return
	}

	switch strings.ToLower(text[:colonIdx]) {
	case "sips":
		address.Encrypted = true
	default:
		address.Encrypted = false
	}
	text = text[colonIdx+1:]
	i = strings.Index(text, "@")
	if i > 0 {
		if n := strings.Index(text[:i], ":"); n > 0 {
			address.User = text[:n]
		} else {
			address.User = text[:i]
		}
		text = text[i+1:]
	}
	if n := strings.Index(text, ":"); n > 0 {
		address.Host = text[:n]
		port, _ := strconv.ParseUint(text[n+1:], 10, 64)
		address.Port = types.Port(port)
	} else {
		address.Host = text
	}

	return
}
