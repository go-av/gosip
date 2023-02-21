package controller

import (
	"fmt"
	"time"
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
			time.Sleep(20 * time.Second)
			deviceIDs := []string{c.user, "34020000001320000051", "34020000001320000011", "34020000001320000002", "34020000001320000003", "34020000001320000041"}
			time.Sleep(1 * time.Second)
			for _, i := range deviceIDs {
				c.server.gb28181.GetDeviceInfo(c, i)

			}
		}()
	}
	return nil
}

func (c *Client) IsAuth() bool {
	return c.auth
}

func (c *Client) Logout() error {
	fmt.Println("用户注销")
	c.auth = false
	return nil
}
