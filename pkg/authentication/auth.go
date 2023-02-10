package authentication

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"regexp"
	"strings"

	"github.com/go-av/gosip/pkg/utils"
)

type Authorization struct {
	realm     string
	nonce     string
	algorithm string
	username  string
	password  string
	uri       string
	response  string
	method    string
	qop       string
	nc        string
	cnonce    string
	opaque    string
	other     map[string]string
}

func Parse(value string) *Authorization {
	auth := &Authorization{
		algorithm: "MD5",
		other:     make(map[string]string),
	}

	re := regexp.MustCompile(`([\w]+)="([^"]+)"`)
	matches := re.FindAllStringSubmatch(value, -1)
	for _, match := range matches {
		switch match[1] {
		case "realm":
			auth.realm = match[2]
		case "algorithm":
			auth.algorithm = match[2]
		case "nonce":
			auth.nonce = match[2]
		case "username":
			auth.username = match[2]
		case "uri":
			auth.uri = match[2]
		case "response":
			auth.response = match[2]
		case "qop":
			for _, v := range strings.Split(match[2], ",") {
				v = strings.Trim(v, " ")
				if v == "auth" || v == "auth-int" {
					auth.qop = "auth"
					break
				}
			}
		case "nc":
			auth.nc = match[2]
		case "cnonce":
			auth.cnonce = match[2]
		case "opaque":
			auth.opaque = match[2]
		default:
			auth.other[match[1]] = match[2]
		}
	}

	return auth
}

func (auth *Authorization) Realm() string {
	return auth.realm
}

func (auth *Authorization) Nonce() string {
	return auth.nonce
}

func (auth *Authorization) Algorithm() string {
	return auth.algorithm
}

func (auth *Authorization) Username() string {
	return auth.username
}

func (auth *Authorization) Opaque() string {
	return auth.opaque
}

func (auth *Authorization) SetUsername(username string) *Authorization {
	auth.username = username
	return auth
}

func (auth *Authorization) SetPassword(password string) *Authorization {
	auth.password = password
	return auth
}

func (auth *Authorization) Uri() string {
	return auth.uri
}

func (auth *Authorization) SetUri(uri string) *Authorization {
	auth.uri = uri
	return auth
}

func (auth *Authorization) SetMethod(method string) *Authorization {
	auth.method = method
	return auth
}

func (auth *Authorization) Response() string {
	return auth.response
}

func (auth *Authorization) SetResponse(response string) {
	auth.response = response
}

func (auth *Authorization) Qop() string {
	return auth.qop
}

func (auth *Authorization) SetQop(qop string) {
	auth.qop = qop
}

func (auth *Authorization) Nc() string {
	return auth.nc
}

func (auth *Authorization) SetNc(nc string) {
	auth.nc = nc
}

func (auth *Authorization) CNonce() string {
	return auth.cnonce
}

func (auth *Authorization) SetCNonce(cnonce string) {
	auth.cnonce = cnonce
}

func (auth *Authorization) String() string {
	if auth == nil {
		return "<nil>"
	}

	str := fmt.Sprintf(`Digest realm="%s"`, auth.realm)

	if auth.algorithm != "" {
		str += fmt.Sprintf(`,algorithm="%s"`, auth.algorithm)
	}

	if auth.nonce != "" {
		str += fmt.Sprintf(`,nonce="%s"`, auth.nonce)
	}

	if auth.username != "" {
		str += fmt.Sprintf(`,username="%s"`, auth.username)
	}

	if auth.uri != "" {
		str += fmt.Sprintf(`,uri="%s"`, auth.uri)
	}

	if auth.response != "" {
		str += fmt.Sprintf(`,response="%s"`, auth.response)
	}

	if auth.qop != "" {
		str += fmt.Sprintf(`,qop="%s"`, auth.qop)
	}

	if auth.nc != "" {
		str += fmt.Sprintf(`,nc=%s`, auth.nc)
	}

	if auth.cnonce != "" {
		str += fmt.Sprintf(`,cnonce="%s"`, auth.cnonce)
	}

	if auth.opaque != "" {
		str += fmt.Sprintf(`,opaque="%s"`, auth.opaque)
	}

	return str
}

func (auth *Authorization) Clone() *Authorization {
	return &Authorization{
		realm:     auth.realm,
		nonce:     auth.nonce,
		algorithm: auth.algorithm,
		username:  auth.username,
		password:  auth.password,
		uri:       auth.uri,
		response:  auth.response,
		method:    auth.method,
		qop:       auth.qop,
		nc:        auth.nc,
		cnonce:    auth.cnonce,
		opaque:    auth.opaque,
		other:     auth.other,
	}
}

func (auth *Authorization) Auth(username string, password string, method string, uri string) *Authorization {
	newAuth := auth.Clone()
	newAuth.SetUsername(username)
	newAuth.SetPassword(password)
	if newAuth.cnonce == "" {
		newAuth.SetCNonce(utils.RandString(10))
	}
	newAuth.SetNc("00000001")

	if uri != "" {
		newAuth.SetUri(uri)
	}
	if method != "" {
		newAuth.SetMethod(method)
	}

	response := CalcResponse(
		newAuth.algorithm,
		newAuth.username,
		newAuth.realm,
		newAuth.password,
		newAuth.method,
		newAuth.uri,
		newAuth.nonce,
		newAuth.qop,
		newAuth.cnonce,
		newAuth.nc,
	)

	newAuth.SetResponse(response)
	return newAuth
}

// CalcResponse Authorization response https://www.ietf.org/rfc/rfc2617.txt
func CalcResponse(algorithm, username, realm, password, method, uri, nonce, qop, cnonce, nc string) string {
	calcA1 := func() string {
		encoder := GetEncoder(algorithm)
		encoder.Write([]byte(username + ":" + realm + ":" + password))
		return hex.EncodeToString(encoder.Sum(nil))
	}
	calcA2 := func() string {
		encoder := GetEncoder(algorithm)
		encoder.Write([]byte(method + ":" + uri))
		return hex.EncodeToString(encoder.Sum(nil))
	}

	encoder := GetEncoder(algorithm)
	encoder.Write([]byte(calcA1() + ":" + nonce + ":"))
	if qop != "" {
		encoder.Write([]byte(nc + ":" + cnonce + ":" + qop + ":"))
	}
	encoder.Write([]byte(calcA2()))

	return hex.EncodeToString(encoder.Sum(nil))
}

func GetEncoder(algorithm string) hash.Hash {
	var encoder hash.Hash
	switch strings.ToUpper(algorithm) {
	case "SHA265":
		encoder = sha256.New()
	case "SHA512":
		encoder = sha512.New()
	default:
		encoder = md5.New()
	}
	return encoder
}
