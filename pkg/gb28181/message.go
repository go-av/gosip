package gb28181

import (
	"encoding/xml"

	"github.com/go-av/gosip/pkg/message"
)

type MessageNotify struct {
	CmdType  string `xml:"CmdType"`
	SN       int    `xml:"SN"`
	DeviceID string `xml:"DeviceID"`
	Status   string `xml:"Status"`
	Info     string `xml:"Info"`
}

type MessageHandler struct {
}

func (handler *MessageHandler) ContentType() string {
	return "Application/MANSCDP+xml"
}

func (handler *MessageHandler) Handler(msg message.Message) (bool, message.Message, error) {
	m := &MessageNotify{}
	data := []byte(msg.Body())
	err := xml.Unmarshal(data, m)

	return false, nil, err
}
