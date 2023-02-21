package controller

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/go-av/gosip/pkg/gb28181"
)

func (d *ServerHandler) Catalog(catalog *gb28181.Catalog) error {
	spew.Dump(catalog)
	// client, err := d.GetClient("34020000001110000002")
	// if err != nil {
	// 	return nil
	// }

	// _, err = d.gb28181.GetDeviceInfo(client, catalog.DeviceID)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// for _, item := range catalog.Item {
	// 	_, err = d.gb28181.GetDeviceInfo(client, item.DeviceID)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	return nil
}
