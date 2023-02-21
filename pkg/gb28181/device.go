package gb28181

import (
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

type DeviceInfo struct {
	CmdType      CmdType `xml:"CmdType"`
	SN           int64   `xml:"SN"`           // 命令序列号 (必选 )
	DeviceID     string  `xml:"DeviceID"`     // 目标设备/区域/系统的编码 (必选 )
	DeviceName   string  `xml:"DeviceName"`   // 目标设备/区域/系统的名称(可选)
	Result       string  `xml:"Result"`       // 查询结果 (必选 )
	DeviceType   string  `xml:"DeviceType"`   // 设备类型
	Manufacturer string  `xml:"Manufacturer"` // 设备生产商 (可选 )
	Model        string  `xml:"Model"`        // 设备型号 (可选 )
	Firmware     string  `xml:"Firmware"`     // 设备固件版本 (可选 )
	Channel      int     `xml:"Channel"`      // 视频输入通道数(可选)
	MaxCamera    int     `xml:"MaxCamera"`
	MaxAlarm     int     `xml:"MaxAlarm"`
}

/*
<Response>

	<CmdType>DeviceInfo</CmdType>
	<SN>1676965736</SN>
	<DeviceID>34020000001110000005</DeviceID>
	<DeviceName>Network Video Recorder</DeviceName>
	<Result>OK</Result>
	<DeviceType>DVR</DeviceType>
	<Manufacturer>HIKVISION</Manufacturer>
	<Model>DS-7916N-R4</Model>
	<Firmware>V4.72.107</Firmware>
	<MaxCamera>16</MaxCamera>
	<MaxAlarm>16</MaxAlarm>

</Response>
*/
func (g *GB28181) GetDeviceInfo(client server.Client, deviceID string) (int64, error) {
	sn := time.Now().Unix()
	_, err := g.SendMessage(client, &Query{
		CmdType:  CmdType__DeviceInfo,
		SN:       sn,
		DeviceID: deviceID,
	})

	if err != nil {
		return 0, err
	}

	return sn, nil
}

func (g *GB28181) DeviceInfo(body []byte) (*server.Response, error) {
	deviceInfo := &DeviceInfo{}
	if err := utils.XMLDecode(body, deviceInfo); err != nil {
		return nil, err
	}
	fmt.Println("xxxxxxxxxxxxx")
	spew.Dump(deviceInfo)
	fmt.Println("xxxxxxxxxxxxx")
	return g.handler.DeviceInfo(deviceInfo)
}
