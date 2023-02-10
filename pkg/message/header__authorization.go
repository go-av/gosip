package message

func NewAuthorizationHeader(data string) *AuthorizationHeader {
	header := AuthorizationHeader(data)
	return &header
}

type AuthorizationHeader string

func (auth *AuthorizationHeader) Name() string {
	return "Authorization"
}

func (auth AuthorizationHeader) Value() string {
	return string(auth)
}

func (auth *AuthorizationHeader) Clone() Header {
	return auth
}

func init() {
	defaultHeaderParsers.Register(NewAuthorizationHeader(""))
}

func (AuthorizationHeader) Parse(data string) (Header, error) {
	return NewAuthorizationHeader(data), nil

}
