package server

type Client interface {
	SetTransport(protocol string, address string)
	Transport() (protocol string, address string)
	User() string
	Password() string
	SetAuth(bool) error
	IsAuth() bool
	Logout() error
}
