package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/go-av/gosip/pkg/client"
	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/sdp"
	reuse "github.com/libp2p/go-reuseport"
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

var sipClient *client.Client
var streamMgr *StreamMgr
var callsdp = ""
var iceServer = ""

func main() {
	// localIP := utils.LocalIp()
	localIP := "172.20.30.61"
	protocol := flag.String("protocol", "udp", "protocol:[udp , tcp],default=udp")
	localAddr := flag.String("local-addr", fmt.Sprintf("%s:5060", localIP), "SIP 本地监听地址")
	serverAddr := flag.String("server-addr", "172.20.50.12:5060", "SIP 服务端地址")
	ice := flag.String("ice-server", "stun:172.20.50.12:3478", "ICE URL")
	httpAddress := flag.String("http-addr", ":80", "HTTP 服务地址")
	mediaPort := flag.Int("mediaPort", 50000, "媒体端口")
	userName := flag.String("userName", "snail", "用户名")
	displayName := flag.String("displayName", "snail", "显示名")
	password := flag.String("password", "admin", "密码")
	flag.Parse()
	iceServer = *ice
	var err error
	sipClient, err = client.NewClient(*userName, *displayName, *password, *localAddr, nil)
	if err != nil {
		panic(err)
	}
	sdpTmp := `v=0
o=- 1661500261 1 IN IP4 {{.ip}}
s=ps
c=IN IP4 {{.ip}}
t=0 0
m=audio {{.mediaPort1}} RTP/AVP 111 0 8
a=rtpmap:111 opus/48000/2
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=mid:audio
a=sendrecv
m=video {{.mediaPort2}} RTP/AVP 96 97 102 122 127 121 125 107 120
a=rtpmap:96 VP8/90000
a=rtpmap:97 VP8/90000
a=rtpmap:102 H264/90000
a=fmtp:102 profile-level-id=42801F
a=rtpmap:122 H264/90000
a=fmtp:122 profile-level-id=42801F
a=rtpmap:127 H264/90000
a=fmtp:127 profile-level-id=42801F
a=rtpmap:121 H264/90000
a=fmtp:121 profile-level-id=42801F
a=rtpmap:125 H264/90000
a=fmtp:125 profile-level-id=42801F
a=rtpmap:107 H264/90000
a=fmtp:107 profile-level-id=42801F
a=rtpmap:120 VP8/90000
a=mid:video
a=sendrecv
`

	tmpl, err := template.New("sip").Parse(sdpTmp)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, map[string]any{
		"ip":         localIP,
		"mediaPort1": *mediaPort,
		"mediaPort2": *mediaPort,
	}); err != nil {
		panic(err)
	}

	sd, err := sdp.ParseSDP(buf.Bytes())
	if err != nil {
		panic(err)
	}
	callsdp = sd.Marshal()

	streamMgr = NewStreamMgr(localIP, *mediaPort)

	ctx, _ := context.WithCancel(context.Background())
	err = sipClient.Registrar(ctx, *serverAddr, *protocol)
	if err != nil {
		panic(err)
	}

	logrus.Println("Listening on", *httpAddress)

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("views"))))
	http.HandleFunc("/call", call)

	err = http.ListenAndServe(*httpAddress, nil)
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

	ctx, cancel := context.WithCancel(context.Background())
	dl, err := sipClient.Call(ctx, userID, callsdp)
	if err != nil {
		resp(w, 500, err.Error(), nil)
		return
	}

	timer := time.NewTimer(30 * time.Second) // 30秒未接，将自动挂断
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			cancel()
			resp(w, 500, "呼叫超时", nil)
			return
		case <-dl.Context().Done():
			return
		case state := <-dl.State():
			fmt.Println("state:", state.State(), state.Reason())
			if state.State() == dialog.Accepted {
				sp, err := sdp.ParseSDP(dl.SDP())
				if err != nil {
					panic(err)
				}
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
				}

				answer, stop := do(dl.Context(), w, udpConns, offer)
				resp(w, 200, "success", encode(answer))
				go func() {
					<-stop
					fmt.Println("退出")
					// dl.Bye()
				}()
				return
			}

			if state.State() == dialog.Error {
				resp(w, 500, state.Reason(), nil)
				return
			}

		}
	}
}

func do(ctx context.Context, w http.ResponseWriter, udpConns map[string]*udpConn, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, chan bool) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{iceServer},
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
			fmt.Println("on track.Kind().String() ", track.Kind().String())
			return
		}

		go func() {
			for {
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
		fmt.Println("video track->", conn.formats, fmt.Sprintf("%s:%d", conn.address, conn.port))

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
		fmt.Println("audio track->", conn.formats, fmt.Sprintf("%s:%d", conn.address, conn.port))
		audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
		if err != nil {
			panic(err)
		}

		_, err = peerConnection.AddTrack(audioTrack)
		if err != nil {
			panic(err)
		}
		streamMgr.LoadOrCreate(fmt.Sprintf("%s:%d", conn.address, conn.port)).SetWriter(audioTrack)
	}
}

func NewStreamMgr(ip string, port int) *StreamMgr {
	mgr := &StreamMgr{}
	go mgr.run(ip, port)
	return mgr
}

type StreamMgr struct {
	streams sync.Map
}

func (mgr *StreamMgr) run(ip string, port int) {
	fmt.Println("ListenPacket:", fmt.Sprintf("%s:%d", ip, port))
	conn, err := reuse.ListenPacket("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 20480)
	i := 0
	for {
		i++
		n, addr, err := conn.ReadFrom(buf)
		if err == nil {
			mgr.LoadOrCreate(addr.String()).write(buf[:n])
		}
	}
}

func (mgr *StreamMgr) LoadOrCreate(address string) *stream {
	if s, ok := mgr.streams.Load(address); ok {
		return s.(*stream)
	}
	s := &stream{}
	mgr.streams.Store(address, s)
	return s
}

type stream struct {
	writer io.Writer
}

func (s *stream) write(buf []byte) {
	if s.writer != nil {
		_, err := s.writer.Write(buf)
		if err != nil {
			fmt.Println("writer err", err)
		}
	}
}

func (s *stream) SetWriter(writer io.Writer) {
	s.writer = writer
}
