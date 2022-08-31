package message

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-av/gosip/pkg/method"
	"github.com/go-av/gosip/pkg/sdp"
	"github.com/sirupsen/logrus"
)

func Parse(src []byte) (Message, error) {
	var msg Message
	var body string
	bodysplit := strings.Split(string(src), "\r\n\r\n")
	if len(bodysplit) > 1 {
		body = bodysplit[1]
	}

	lines := strings.Split(bodysplit[0], "\r\n")
	if len(lines) == 0 {
		return nil, errors.New("src is not message")
	}

	if isRequest(lines[0]) {
		md, recipient, _, err := ParseRequestLine(lines[0])
		if err != nil {
			return nil, err
		}
		msg = NewRequestMessage("", md, recipient)
	} else if isResponse(lines[0]) {
		_, statusCode, statusDesc, err := ParseStatusLine(lines[0])
		if err != nil {
			return nil, err
		}
		msg = NewResponse(nil, statusCode, statusDesc)
	} else {
		return nil, errors.New("failed to read start line ")
	}

	msg.SetSrc(src)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if i == 0 {
			continue
		}
		index := strings.Index(line, ":")
		if index <= 0 {
			logrus.Debugf("header illegal: %s", line)
			continue
		}
		parser, ok := defaultHeaderParsers.Get(line[:index])
		if !ok {
			fmt.Println("not found parser:", line)
			logrus.Debugf("%s not found parser", line[:index])
			continue
		}
		h, err := parser.Parse(strings.TrimSpace(line[index+1:]))
		if err != nil {
			logrus.Debugf("%s parse is error:%s", line[:index], err.Error())
			continue
		}
		msg.AppendHeader(h)
	}

	if contentType, ok := msg.ContentType(); ok {
		sdp := &sdp.SDP{}
		switch contentType.Value() {
		case sdp.ContentType():
			if err := sdp.Unmarshal([]byte(body)); err == nil {
				msg.SetBody(sdp)
			} else {
				logrus.Errorf("sdp.Unmarshal err:%s", err)
			}
		}
	}

	return msg, nil
}

func isResponse(startLine string) bool {
	if strings.Count(startLine, " ") < 2 {
		return false
	}

	parts := strings.Split(startLine, " ")
	if len(parts) < 3 {
		return false
	} else if len(parts[0]) < 3 {
		return false
	} else {
		return strings.ToUpper(parts[0][:3]) == "SIP"
	}
}

func isRequest(startLine string) bool {
	if strings.Count(startLine, " ") != 2 {
		return false
	}

	parts := strings.Split(startLine, " ")
	if len(parts) < 3 {
		return false
	} else if len(parts[2]) < 3 {
		return false
	} else {
		return strings.ToUpper(parts[2][:3]) == "SIP"
	}
}

func ParseRequestLine(requestLine string) (md method.Method, recipient *Address, sipVersion string, err error) {
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		err = fmt.Errorf("request line should have 2 spaces: '%s'", requestLine)
		return
	}

	md = method.Method(strings.ToUpper(parts[0]))
	_, recipient, err = ParamAddress(parts[1])
	sipVersion = parts[2]
	return
}

func ParseStatusLine(statusLine string) (sipVersion string, statusCode StatusCode, statusDesc string, err error) {
	parts := strings.Split(statusLine, " ")
	if len(parts) < 3 {
		err = fmt.Errorf("status line has too few spaces: '%s'", statusLine)
		return
	}

	sipVersion = parts[0]
	statusCodeRaw, err := strconv.ParseUint(parts[1], 10, 16)
	statusCode = StatusCode(statusCodeRaw)
	statusDesc = strings.Join(parts[2:], " ")
	return
}
