package authentication_test

import (
	"fmt"
	"testing"

	"github.com/go-av/gosip/pkg/authentication"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	auth := authentication.Parse(`
		Digest realm="testrealm@host.com",
		qop="auth",
		nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093",
		opaque="5ccc069c403ebaf9f0171e9517f40e41"
	`)
	require.Equal(t, auth.Realm(), "testrealm@host.com")
	require.Equal(t, auth.Qop(), "auth")
	require.Equal(t, auth.Nonce(), "dcd98b7102dd2f0e8b11d0f600bfb0c093")
	require.Equal(t, auth.Opaque(), "5ccc069c403ebaf9f0171e9517f40e41")
}

func TestAuth(t *testing.T) {
	auth := authentication.Parse(`
		Digest realm="testrealm@host.com",
		qop="auth,auth-int",
		nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093",
		nc=0000001,
		opaque="5ccc069c403ebaf9f0171e9517f40e41"
	`)
	auth.SetCNonce("test")
	ss := auth.Auth("user", "pssword", "GET", "/dir/index.html")
	require.Equal(t, ss.Response(), "5e103ea9b42446be725d20c8265b0638")
}

func TestAAAAAA(t *testing.T) {
	str := `Digest username="50010600002000000003",realm="5001020001",nonce="caSY5M8se7r9ZoxeaDah",uri="sip:11111111@172.20.30.57:5060",response="f116fe8aa4efdfba3f070fd45d5f3877",algorithm=MD5,qop=auth,cnonce="1b14aac3-5613-460f-a23b-385331d3bdba",nc="00000001"`
	auth := authentication.Parse(str)
	fmt.Println(auth)
}
