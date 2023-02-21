package gb28181

import (
	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

type Keepalive struct {
	CmdType      CmdType `xml:"CmdType"`
	SN           int     `xml:"SN"`
	DeviceID     string  `xml:"DeviceID"`
	Status       string  `xml:"Status"`
	InfoDeviceID string  `xml:"Info>DeviceID"`
}

func (g *GB28181) Keepalive(body []byte) (*server.Response, error) {
	kl := &Keepalive{}
	if err := utils.XMLDecode(body, kl); err != nil {
		return nil, err
	}

	return g.handler.Keepalive(kl)
}
