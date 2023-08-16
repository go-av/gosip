package gb28181

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/go-av/gosip/pkg/server"
	"github.com/go-av/gosip/pkg/utils"
)

type Broadcast struct {
	XMLName  xml.Name `xml:"Notify"`
	CmdType  CmdType  `xml:"CmdType"`
	SN       int64    `xml:"SN"`
	SourceID string   `xml:"SourceID"`
	TargetID string   `xml:"TargetID"`
}

type BroadcastResponse struct {
	XMLName  xml.Name `xml:"Response"`
	SN       int64    `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Result   string   `xml:"Result"`
}

func (g *GB28181) StartBroadcast(client server.Client, sourceID string, targetID string) (int64, error) {
	sn := time.Now().Unix()
	_, err := g.SendMessage(client, &Broadcast{
		CmdType:  CmdType__Broadcast,
		SN:       sn,
		SourceID: sourceID,
		TargetID: targetID,
	})

	if err != nil {
		return 0, err
	}

	return sn, nil
}

func (g *GB28181) Broadcast(ctx context.Context, client server.Client, body []byte) (*server.Response, error) {
	cl := &BroadcastResponse{}
	if err := utils.XMLDecode(body, cl); err != nil {
		return nil, err
	}

	g.handler.Broadcast(ctx, client, cl)

	return server.NewResponse(200, "success."), nil
}
