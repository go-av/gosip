package message

import (
	"github.com/go-av/gosip/pkg/authentication"
)

func NewWWWAuthenticateHeader() *WWWAuthenticateHeader {
	auth := &WWWAuthenticateHeader{}
	return auth
}

type WWWAuthenticateHeader authentication.Authorization

func (auth *WWWAuthenticateHeader) Name() string {
	return "WWW-Authenticate"
}

func (auth WWWAuthenticateHeader) Value() string {
	a := authentication.Authorization(auth)
	return a.String()
}

func (auth *WWWAuthenticateHeader) Clone() Header {
	return auth
}

func (auth WWWAuthenticateHeader) Auth(username string, password string, uri string) *AuthorizationHeader {
	a := authentication.Authorization(auth)
	return NewAuthorizationHeader(a.Auth(username, password, "REGISTER", uri).String())
}

func init() {
	defaultHeaderParsers.Register(NewWWWAuthenticateHeader())
}

func (WWWAuthenticateHeader) Parse(data string) (Header, error) {
	aa := authentication.Parse(data)
	dd := WWWAuthenticateHeader(*aa)
	return &dd, nil
}
