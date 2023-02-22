package gb28181

import (
	"encoding/xml"
	"time"

	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

type DeviceStatus struct {
	XMLName    xml.Name `xml:"Response"`
	CmdType    CmdType  `xml:"CmdType"`
	SN         int64    `xml:"SN"`                   // 命令序列号 (必选 )
	DeviceID   string   `xml:"DeviceID"`             // 目标设备/区域/系统的编码 (必选 )
	Result     string   `xml:"Result"`               // 查询结果标志 (必选 )
	Online     string   `xml:"Online"`               // 是否在线 (必选 )
	Status     string   `xml:"Status"`               // 是否正常工作 (必选 )
	Reason     string   `xml:"Reason,omitempty"`     // 不正常工作原因(可选)
	Encode     string   `xml:"Encode,omitempty"`     // 是否编码 (可选 )
	Record     string   `xml:"Record,omitempty"`     // 是否录像 (可选 )
	DeviceTime string   `xml:"DeviceTime,omitempty"` // 设备时间和日期(可选)
	// todo 设备告警信息
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

func (g *GB28181) GetDeviceStatus(client server.Client, deviceID string) (int64, error) {
	sn := time.Now().Unix()
	_, err := g.SendMessage(client, &Query{
		CmdType:  CmdType__DeviceStatus,
		SN:       sn,
		DeviceID: deviceID,
	})
	if err != nil {
		return 0, err
	}
	return sn, nil
}

func (g *GB28181) DeviceStatus(body []byte) (*server.Response, error) {
	deviceStatus := &DeviceStatus{}
	if err := utils.XMLDecode(body, deviceStatus); err != nil {
		return nil, err
	}
	return g.handler.DeviceStatus(deviceStatus)
}
