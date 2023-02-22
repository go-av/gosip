package controller

import (
	"fmt"
	"time"

	"github.com/go-av/gosip/pkg/utils/ptz"
)

type Client struct {
	server   *ServerHandler
	user     string
	protocol string
	address  string
	auth     bool
}

func (c *Client) SetTransport(protocol string, address string) {
	if c.protocol != protocol || c.address != address {
		c.address = address
		c.protocol = protocol
		c.auth = false
	}
}

func (c *Client) Transport() (protocol string, address string) {
	return c.protocol, c.address
}

func (c *Client) User() string {
	return c.user
}

func (c *Client) Password() string {
	return "12345678"
}

func (c *Client) SetAuth(auth bool) error {
	c.auth = auth
	if auth {
		go func() {
			time.Sleep(1 * time.Second)
			c.server.gb28181.GetCatalog(c)
			// time.Sleep(5 * time.Second)
			// deviceIDs := []string{c.user, "71020001001320000001"}
			// time.Sleep(1 * time.Second)
			// for _, deviceID := range deviceIDs {
			// 	// c.server.gb28181.GetDeviceInfo(c, deviceID)
			// 	// c.server.gb28181.GetDeviceStatus(c, deviceID)
			// 	c.server.gb28181.GetPresetQuery(c, i)
			// 	// c.server.gb28181.GetDeviceConfig(c, deviceID)
			// }

			deviceID := c.user
			// // 预制点位调试
			// all := []ptz.PTZ_Type{ptz.Right, ptz.Left, ptz.Left, ptz.Up, ptz.Down, ptz.LeftUp, ptz.LeftDown, ptz.RightUp, ptz.RightDown, ptz.Stop}
			// for _, a := range all {
			// 	fmt.Println("方位调整", string(a))
			// 	c.server.gb28181.PTZControl(c, deviceID, ptz.PTZCmd(a, 2, 0))
			// 	time.Sleep(5 * time.Second)
			// }
			c.server.gb28181.PTZControl(c, deviceID, ptz.PTZCmd(ptz.Left, 0, 1))
			time.Sleep(2 * time.Second)
			c.server.gb28181.GetPresetQuery(c, deviceID)
			fmt.Println("调用预制点位")
			c.server.gb28181.PTZControl(c, deviceID, ptz.PTZCmd(ptz.CalPos, 0, 1))
		}()
	}
	return nil
}

func (c *Client) IsAuth() bool {
	return c.auth
}

func (c *Client) Logout() error {
	fmt.Println("用户注销-----")
	c.auth = false
	return nil
}
