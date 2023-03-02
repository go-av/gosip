package sip

import (
	"fmt"
	"sync"

	"github.com/go-av/gosip/pkg/gb28181"
	"github.com/go-av/gosip/pkg/message"
	"github.com/go-av/gosip/pkg/server"
)

func NewSipHandler(sipID string, realm string) *SipHandler {
	handler := &SipHandler{
		sipID:   sipID,
		realm:   realm,
		clients: &sync.Map{},
		gb28181: &gb28181.GB28181{},
	}

	return handler
}

type SipHandler struct {
	sipID   string
	realm   string
	server  server.Server
	clients *sync.Map
	gb28181 *gb28181.GB28181
}

func (server *SipHandler) SetServer(s server.Server) {
	server.server = s
	server.gb28181 = gb28181.NewGB28181(s, server)
	go func() {
		for {
			select {
			case dl := <-s.Receive():
				fmt.Println(dl.From().User(), "呼叫", dl.To().User())
				// dd, err := sdp.ParseSDP([]byte(sdp2))
				// if err != nil {
				// 	panic(err)
				// }
				// dl.Answer(dd.Marshal())
				// // 99920000002000000000
				// time.Sleep(5 * time.Second)
				// dl.Bye()
			}
		}
	}()
}

var sdp2 = `v=0
o=- 3868331676 3868331676 IN IP4 172.20.30.61
s=gosip 1.0.0
t=0 0
m=audio 40026 RTP/AVP 8 0 101
c=IN IP4 172.20.30.61
a=rtcp:50007
a=rtpmap:8 PCMA/8000
a=rtpmap:0 PCMU/8000
a=rtpmap:101 telephone-event/8000
m=video 40026 RTP/AVP 96
c=IN IP4 172.20.30.61
a=rtcp:50009
a=rtpmap:96 VP8/90000
a=sendrecv
`

func (d *SipHandler) GetClient(user string) (server.Client, error) {
	client, ok := d.clients.LoadOrStore(user, &Client{
		user:   user,
		auth:   false,
		server: d,
	})
	if !ok {
		fmt.Println("发现新用户-----", user)
	}
	return client.(*Client), nil
}

func (d *SipHandler) Realm() string {
	return d.realm
}

func (d *SipHandler) ServerSIPID() string {
	return d.sipID
}

func (d *SipHandler) ReceiveMessage(body message.Body) (*server.Response, error) {
	if body.ContentType() == "Application/MANSCDP+xml" {
		return d.gb28181.Handler(body.Data())
	}
	return server.NewResponse(400, "unknown"), nil
}
