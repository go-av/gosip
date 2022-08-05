package message

func NewAuthorizationHeader(auth string) *AuthorizationHeader {
	au := AuthorizationHeader(auth)
	return &au
}

type AuthorizationHeader string

func (callId *AuthorizationHeader) Name() string {
	return "Authorization"
}

func (auth AuthorizationHeader) Value() string {
	return string(auth)
}

func (auth *AuthorizationHeader) Clone() Header {
	return auth
}
