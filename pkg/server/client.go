package server

type Client interface {
	SetTransport(protocol string, address string)
	Transport() (protocol string, address string)
	User() string
	Password() string
	SetAuth(bool) error
	SetKeepalive() error
	IsAuth() bool
	Logout() error
}
