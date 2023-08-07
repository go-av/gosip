package gb28181

import (
	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

/*
<Notify>

	<CmdType>Keepalive</CmdType>
	<SN>67</SN>
	<DeviceID>34020000001110000005</DeviceID>
	<Status>OK</Status>
	<Info>
	<DeviceID>34020000001320000051</DeviceID>
	</Info>

</Notify>
*/
type Keepalive struct {
	CmdType       CmdType  `xml:"CmdType"`
	SN            int      `xml:"SN"`
	DeviceID      string   `xml:"DeviceID"`
	Status        string   `xml:"Status"`
	InfoDeviceIDs []string `xml:"Info>DeviceID"`
}

func (g *GB28181) Keepalive(client server.Client, body []byte) (*server.Response, error) {
	kl := &Keepalive{}
	if err := utils.XMLDecode(body, kl); err != nil {
		return nil, err
	}
	return g.handler.Keepalive(client, kl)
}
