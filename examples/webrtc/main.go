package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"text/template"
	"time"

	"github.com/go-av/gosip/examples/webrtc/controller"
	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/sdp"
	"github.com/go-av/gosip/pkg/sip"
	"github.com/go-av/gosip/pkg/types"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
)

/**
 * 临时 demo,未进行优化
 */
func decode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		return err
	}
	return nil
}

func encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

type udpConn struct {
	conn    net.Conn
	address string
	port    int
	formats []string
}

func resp(w http.ResponseWriter, code int, msg string, data any) {
	resp := map[string]any{
		"code": code,
		"msg":  msg,
		"data": data,
	}
	d, _ := json.Marshal(resp)
	w.Write(d)
}

var sipClient *sip.Client
var streamMgr *controller.StreamMgr

func main() {
	httpAddress := flag.String("httpAddress", ":80", "Address to host the HTTP server on.")
	localAddress := flag.String("localSIPAddress", controller.ResolveLocalIP().String(), "sip udp address")
	loadlPort := flag.Int("localSIPPort", 5060, "sip udp port")
	mediaPort := flag.Int("mediaPort", 50000, "media port")

	userName := flag.String("userName", "snail", "用户名")
	displayName := flag.String("displayName", "snail", "显示名")
	password := flag.String("password", "admin", "admin")

	sipServerAddress := flag.String("sipServerAddress", "172.20.50.12", "sip server address")
	sipServerPort := flag.Int("sipServerPort", 5060, "sip server port")

	flag.Parse()

	sipClient = sip.NewClient(*userName, *displayName, *password, *localAddress, types.Port(*loadlPort))
	sipClient.SetSDP(func(*sdp.SDP) *sdp.SDP {
		sdpTmp := `v=0
o=- 1661500261 1 IN IP4 {{.ip}}
s=gosip 1.0.0
c=IN IP4 {{.ip}}
t=0 0
m=audio {{.port}} RTP/AVP 111 0 8
a=rtpmap:111 opus/48000/2
a=fmtp:111 minptime=10;useinbandfec=1
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=mid:audio
a=sendrecv
m=video {{.port}} RTP/AVP 96 97
a=rtpmap:96 VP8/90000
a=rtpmap:97 H264/90000
a=mid:video
a=sendrecv
`

		tmpl, err := template.New("sip").Parse(sdpTmp)
		if err != nil {
			panic(err)
		}
		buf := bytes.NewBuffer(nil)
		if err := tmpl.Execute(buf, map[string]any{
			"ip":   *localAddress,
			"port": *mediaPort,
		}); err != nil {
			panic(err)
		}

		fmt.Println(buf.String())
		sd, _ := sdp.ParseSDP(buf.String())
		return sd
	})

	streamMgr = controller.NewStreamMgr(*localAddress, *mediaPort)

	ctx, _ := context.WithCancel(context.Background())
	sipClient.Start(ctx, "udp", *sipServerAddress, *sipServerPort)

	logrus.Println("Listening on", *httpAddress)

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("views"))))
	http.HandleFunc("/call", call)

	err := http.ListenAndServe(*httpAddress, nil)
	if err != nil {
		logrus.Error("Failed to serve: %v", err)
		return
	}
}

func call(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query()

	userID := data.Get("userID")
	sd := data.Get("sdp")

	offer := &webrtc.SessionDescription{}
	err := decode(sd, offer)
	if err != nil {
		resp(w, 400, "SDP 错误", nil)
		return
	}

	if userID == "" {
		resp(w, 400, "对方用户名错误", nil)
		return
	}

	fmt.Println("call", userID)
	dl, err := sipClient.Call(userID)
	if err != nil {
		resp(w, 500, err.Error(), nil)
		return
	}

	timer := time.NewTimer(30 * time.Second) // 30秒未接，将自动挂断
	for {
		select {
		case <-timer.C:
			resp(w, 500, "呼叫超时", nil)
			return
		case state := <-dl.State():
			fmt.Println("state=======", state)
			if state == dialog.Answered {
				sp := dl.SDP()
				udpConns := map[string]*udpConn{}
				for _, media := range sp.MediaDescriptions {
					if media.MediaName.Port.Value > 0 {
						udpConns[media.MediaName.Media] = &udpConn{
							address: sp.Origin.UnicastAddress,
							port:    media.MediaName.Port.Value,
							formats: media.MediaName.Formats,
						}
					}
				}

				for _, c := range udpConns {
					if c.conn, err = net.Dial("udp", fmt.Sprintf("%s:%d", c.address, c.port)); err != nil {
						panic(err)
					}
					fmt.Println("dial", fmt.Sprintf("%s:%d", c.address, c.port))
				}

				answer, stop := do(context.Background(), w, udpConns, offer)
				resp(w, 200, "success", encode(answer))
				go func() {
					<-stop
					fmt.Println("退出")
					dl.Hangup()
				}()

				return
			}
			if state == dialog.Hangup {
				return
			}
			if state == dialog.Error {
				resp(w, int(dl.StatusCode()), dl.Reason(), nil)
				return
			}
		}
	}
}

func do(ctx context.Context, w http.ResponseWriter, udpConns map[string]*udpConn, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, chan bool) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})

	mediaForwarding(peerConnection, udpConns)

	if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}
	if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	stop := make(chan bool, 1)
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("OnTrack", track.Kind().String())
		c, ok := udpConns[track.Kind().String()]
		if !ok {
			return
		}

		go func() {
			ticker := time.NewTicker(time.Second * 2)
			for range ticker.C {
				if rtcpErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}}); rtcpErr != nil {
					fmt.Println(rtcpErr)
				}
			}
		}()

		b := make([]byte, 1500)

		for {
			n, _, err := track.Read(b)
			if err != nil {
				fmt.Println(err)
			}
			if _, err := c.conn.Write(b[:n]); err != nil {
				fmt.Println(err)
			}
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateFailed {
			stop <- true
		}
		if connectionState == webrtc.ICEConnectionStateClosed {
			stop <- true
		}
	})

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())
		if s == webrtc.PeerConnectionStateFailed {
			fmt.Println("Peer Connection has gone to failed exiting")
			stop <- true
		}
		if s == webrtc.PeerConnectionStateDisconnected {
			fmt.Println("Peer Connection has gone to disconnected")
			stop <- true
		}
	})

	if err := peerConnection.SetRemoteDescription(*offer); err != nil {
		panic(err)
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	<-gatherComplete

	return peerConnection.LocalDescription(), stop
}

func mediaForwarding(peerConnection *webrtc.PeerConnection, udpConns map[string]*udpConn) {
	if conn, ok := udpConns["video"]; ok {
		fmt.Println(conn.formats)
		mimeType := webrtc.MimeTypeVP8

		videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: mimeType}, "video", "pion")
		if err != nil {
			panic(err)
		}
		_, err = peerConnection.AddTrack(videoTrack)
		if err != nil {
			panic(err)
		}

		streamMgr.LoadOrCreate(fmt.Sprintf("%s:%d", conn.address, conn.port)).SetWriter(videoTrack)
	}

	if conn, ok := udpConns["audio"]; ok {
		audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
		if err != nil {
			panic(err)
		}

		fmt.Println("add audio Track ")

		_, err = peerConnection.AddTrack(audioTrack)
		if err != nil {
			panic(err)
		}
		streamMgr.LoadOrCreate(fmt.Sprintf("%s:%d", conn.address, conn.port)).SetWriter(audioTrack)
	}
}
