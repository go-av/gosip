package sip

import (
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-av/gosip/pkg/gb28181"
	"github.com/go-av/gosip/pkg/server"
)

func (d *SipHandler) DeviceInfo(ctx context.Context, client server.Client, msg *gb28181.DeviceInfo) (*server.Response, error) {
	fmt.Println("xxxxxxxxxxxxxx")
	spew.Dump(msg)
	fmt.Println("xxxxxxxxxxxxxx")
	return server.NewResponse(200, "Success"), nil
}

func (d *SipHandler) DeviceStatus(ctx context.Context, client server.Client, msg *gb28181.DeviceStatus) (*server.Response, error) {
	fmt.Println("xxxxxxxxxxxxxx")
	spew.Dump(msg)
	fmt.Println("xxxxxxxxxxxxxx")
	return server.NewResponse(200, "Success"), nil
}

func (d *SipHandler) PresetQuery(ctx context.Context, client server.Client, msg *gb28181.PresetQuery) (*server.Response, error) {
	fmt.Println("xxxxxxxxxxxxxx")
	spew.Dump(msg)
	fmt.Println("xxxxxxxxxxxxxx")
	return server.NewResponse(200, "Success"), nil
}
