package authentication_test

import (
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
		opaque="5ccc069c403ebaf9f0171e9517f40e41"
	`)
	auth.SetCNonce("test")
	ss := auth.Auth("user", "pssword", "GET", "/dir/index.html")
	require.Equal(t, ss.Response(), "5e103ea9b42446be725d20c8265b0638")
}
