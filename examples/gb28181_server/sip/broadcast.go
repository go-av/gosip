package sip

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/go-av/gosip/pkg/gb28181"
	"github.com/go-av/gosip/pkg/server"
)

func (d *SipHandler) Broadcast(client server.Client, bl *gb28181.BroadcastResponse) {
	spew.Dump(bl)
}

func (d *SipHandler) StartBroadcast(client server.Client, sourceID string, targetID string) (int64, error) {
	return d.gb28181.StartBroadcast(client, sourceID, targetID)
}
