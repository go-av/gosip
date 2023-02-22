package message

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func NewAddress(user string, host string, port uint16) *Address {
	return &Address{
		User: user,
		Host: host,
		Port: port,
	}
}

type Address struct {
	Domain    string
	Encrypted bool
	User      string
	Host      string
	Port      uint16
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
	if s.Domain != "" {
		buf.WriteString(s.Domain)
	} else {
		buf.WriteString(s.Host)
		if s.Port > 0 {
			buf.WriteString(":" + strconv.FormatUint(uint64(s.Port), 10))
		}
	}
	return buf.String()
}

func (s *Address) WithDomain(domain string) *Address {
	s.Domain = domain
	return s
}

func (address *Address) Clone() *Address {
	return &Address{
		Domain:    address.Domain,
		Encrypted: address.Encrypted,
		User:      address.User,
		Host:      address.Host,
		Port:      address.Port,
	}
}

func (address *Address) SetUser(user string) *Address {
	address.User = user
	return address
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
		address.Port = uint16(port)
	} else {
		address.Host = text
	}

	return
}
