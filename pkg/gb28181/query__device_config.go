package gb28181

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/go-av/gosip/pkg/server"
)

type DeviceConfig struct {
	XMLName  xml.Name `xml:"Response"`
	CmdType  CmdType  `xml:"CmdType"`
	SN       int64    `xml:"SN"`       // 命令序列号 (必选 )
	DeviceID string   `xml:"DeviceID"` // 目标设备/区域/系统的编码 (必选 )
}

/*
<?xml version="1.0" encoding="gb2312"?>
<Response>
	<CmdType>DeviceStatus</CmdType>
	<SN>1676968400</SN>
	<DeviceID>34020000001320000041</DeviceID>
	<Result>OK</Result>
	<Online>ONLINE</Online>
	<Status>OK</Status>
	<Encode>ON</Encode>
	<Record>ON</Record>
</Response>
*/

func (g *GB28181) GetDeviceConfig(client server.Client, deviceID string) (int64, error) {
	sn := time.Now().Unix()
	_, err := g.SendMessage(client, &Query{
		CmdType:  CmdType__ConfigDownload,
		SN:       sn,
		DeviceID: deviceID,
	})
	if err != nil {
		return 0, err
	}
	return sn, nil
}

func (g *GB28181) ConfigDownload(ctx context.Context, client server.Client, body []byte) (*server.Response, error) {
	fmt.Println("xxxxxxxxxxx")
	fmt.Println(string(body))
	fmt.Println("xxxxxxxxxxx")
	return server.NewResponse(200, "OK"), nil
}
