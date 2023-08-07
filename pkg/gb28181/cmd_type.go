package gb28181

type CmdType string

const (
	CmdType__Catalog        CmdType = "Catalog"
	CmdType__Keepalive      CmdType = "Keepalive"
	CmdType__RecordInfo     CmdType = "RecordInfo"
	CmdType__DeviceInfo     CmdType = "DeviceInfo"
	CmdType__ConfigDownload CmdType = "ConfigDownload"
	CmdType__Broadcast      CmdType = "Broadcast"
	CmdType__DeviceStatus   CmdType = "DeviceStatus"
	CmdType__PresetQuery    CmdType = "PresetQuery"
	CmdType__DeviceControl  CmdType = "DeviceControl"
)
