package dialog

import "github.com/go-av/gosip/pkg/utils"

// 本地
type From interface {
	Protocol() string
	HostAndPort() *utils.HostAndPort
	User() string
	DisplayName() string
}

func NewFrom(displayName string, user string, protocol string, hostAndPort string) From {
	hp, _ := utils.ParseHostAndPort(hostAndPort)
	return &from{
		user:        user,
		displayName: displayName,
		protocol:    protocol,
		hostAndPort: hp,
	}
}

type from struct {
	displayName string
	user        string
	protocol    string
	hostAndPort *utils.HostAndPort
}

func (f *from) User() string {
	return f.user
}

func (f *from) DisplayName() string {
	return f.displayName
}

func (f *from) Protocol() string {
	return f.protocol
}

func (f *from) HostAndPort() *utils.HostAndPort {
	return f.hostAndPort
}
