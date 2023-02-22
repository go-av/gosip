package gb28181

import "encoding/xml"

type Query struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  CmdType  `xml:"CmdType"`
	SN       int64    `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
}
