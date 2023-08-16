package gb28181

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

// type
/*

<?xml version="1.0" encoding="gb2312"?>
<Response>
<CmdType>PresetQuery</CmdType>
<SN>216025925</SN>
<DeviceID>34020000001320000061</DeviceID>
<PresetList Num="6">
<Item>
<PresetID>1</PresetID>
<PresetName>1</PresetName>
</Item>
<Item>
<PresetID>2</PresetID>
<PresetName>2</PresetName>
</Item>
<Item>
<PresetID>3</PresetID>
<PresetName>3</PresetName>
</Item>
<Item>
<PresetID>4</PresetID>
<PresetName>4</PresetName>
</Item>
<Item>
<PresetID>5</PresetID>
<PresetName>5</PresetName>
</Item>
<Item>
<PresetID>7</PresetID>
<PresetName>7</PresetName>
</Item>
</PresetList>
</Response>

*/

type PresetQuery struct {
	XMLName    xml.Name   `xml:"Response"`
	CmdType    CmdType    `xml:"CmdType"`
	SN         int64      `xml:"SN"`       // 命令序列号 (必选 )
	DeviceID   string     `xml:"DeviceID"` // 目标设备/区域/系统的编码 (必选 )
	PresetList PresetList `xml:"PresetList"`
}

type PresetList struct {
	Num  int      `xml:"Num,attr"`
	Item []Preset `xml:"Item"`
}

type Preset struct {
	PresetID   string `xml:"PresetID"`   // 预置位编码(必选)
	PresetName string `xml:"PresetName"` // 预置位名称(必选)
}

func (g *GB28181) GetPresetQuery(client server.Client, deviceID string) (int64, error) {
	sn := time.Now().Unix()
	_, err := g.SendMessage(client, &Query{
		CmdType:  CmdType__PresetQuery,
		SN:       sn,
		DeviceID: deviceID,
	})
	if err != nil {
		return sn, err
	}
	return sn, nil
}

func (g *GB28181) PresetQuery(ctx context.Context, client server.Client, body []byte) (*server.Response, error) {
	presetQuery := &PresetQuery{}
	if err := utils.XMLDecode(body, presetQuery); err != nil {
		return nil, err
	}

	return g.handler.PresetQuery(ctx, client, presetQuery)
}
