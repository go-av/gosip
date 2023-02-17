package controller

import (
	"fmt"
	"time"

	"github.com/go-av/gosip/pkg/server"
)

type Client struct {
	server   server.Server
	user     string
	protocol string
	address  string
	auth     bool
	deviceID string
	sn       string
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

			resp, err := c.server.SendMessage(c, server.NewContent("Application/MANSCDP+xml", []byte(`<Query>
			<CmdType>Catalog</CmdType>
			<SN>`+fmt.Sprintf("%d", time.Now().Unix())+`</SN>
			<DeviceID>`+c.user+`</DeviceID>
		</Query>`)))

			if err != nil {
				fmt.Println("senterr", err)
			}
			if resp != nil {
				fmt.Println("resp", resp.ContentType(), resp.Data())
			}
		}()
	}
	return nil
}

func (c *Client) IsAuth() bool {
	return c.auth
}

func (c *Client) SetKeepalive() error {
	return nil
}

func (c *Client) Logout() error {
	c.auth = false
	return nil
}
