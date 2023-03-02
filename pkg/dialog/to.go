package dialog

import "github.com/go-av/gosip/pkg/utils"

// 远端
type To interface {
	User() string
	HostAndPort() *utils.HostAndPort
}

func NewTo(user string, hostAndPort string) To {
	hp, _ := utils.ParseHostAndPort(hostAndPort)

	return &to{
		user:        user,
		hostAndPort: hp,
	}
}

type to struct {
	user        string
	hostAndPort *utils.HostAndPort
}

func (t *to) User() string {
	return t.user
}

func (t *to) HostAndPort() *utils.HostAndPort {
	return t.hostAndPort
}
