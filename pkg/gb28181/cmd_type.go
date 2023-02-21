package gb28181

type CmdType string

const (
	CmdType__Catalog    CmdType = "Catalog"
	CmdType__Keepalive  CmdType = "Keepalive"
	CmdType__RecordInfo CmdType = "RecordInfo"
	CmdType__DeviceInfo CmdType = "DeviceInfo"
)
