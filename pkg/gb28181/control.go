package gb28181

import (
	"time"

	"github.com/go-av/gosip/pkg/server"
)

type Control struct {
	CmdType   CmdType `xml:"CmdType"`
	SN        int64   `xml:"SN"`
	DeviceID  string  `xml:"DeviceID"`
	PTZCmd    string  `xml:"PTZCmd,omitempty"`    // 球机/云台控制命令(可选,控制码应符合附录 A 中的 A.3中的规定)
	TeleBoot  string  `xml:"TeleBoot,omitempty"`  // 远 程 启 动 控 制 命 令 (可 选 )
	RecordCmd string  `xml:"RecordCmd,omitempty"` // 录 像 控 制 命 令 (可 选 )
	GuardCmd  string  `xml:"GuardCmd,omitempty"`  // 报警布防/撤防命令(可选)
}

func (g *GB28181) DeviceControl(client server.Client,) (int64, error) {
	sn := time.Now().Unix()
	_, err := g.SendMessage(client, &Query{
		CmdType:  CmdType__Catalog,
		SN:       sn,
		DeviceID: client.User(),
	})

	if err != nil {
		return 0, err
	}

	return sn, nil
}
