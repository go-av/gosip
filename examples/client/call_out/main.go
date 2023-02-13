package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/go-av/gosip/pkg/client"
	"github.com/go-av/gosip/pkg/client/dialog"
	"github.com/go-av/gosip/pkg/sdp"
	"github.com/go-av/gosip/pkg/utils"
	"github.com/go-cmd/cmd"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/sirupsen/logrus"
	"github.com/youpy/go-wav"
)

func main() {
	localIP := utils.LocalIp()
	protocol := flag.String("protocol", "udp", "protocol:[udp , tcp],default=udp")
	localAddr := flag.String("local-addr", fmt.Sprintf("%s:5060", localIP), "SIP IP")
	serverAddr := flag.String("server-addr", "172.20.50.12:5060", "SIP 服务端地址")
	to := flag.String("to", "snail_in", "call to user")
	flag.Parse()

	client, err := client.NewClient("蜗牛", "34030000001110000002", "12345678", *protocol, *localAddr)
	if err != nil {
		panic(err)
	}
	client.SetSDP(func(*sdp.SDP) *sdp.SDP {
		str := `v=0
o=- 3868331676 3868331676 IN IP4 %s
s=gosip 1.0.0
t=0 0
m=audio 50006 RTP/AVP 8 0 101
c=IN IP4 %s
a=rtcp:50007
a=rtpmap:8 PCMA/8000
a=rtpmap:0 PCMU/8000
a=rtpmap:101 telephone-event/8000
m=video 50006 RTP/AVP 96
c=IN IP4 %s
a=rtcp:50009
a=rtpmap:96 VP8/90000
a=sendrecv
`
		sd, err := sdp.ParseSDP(fmt.Sprintf(str, localIP, localIP, localIP))
		if err != nil {
			panic(err)
		}
		return sd
	})

	ctx, cancel := context.WithCancel(context.Background())
	// err := client.Start(ctx, "172.20.50.12", 5060)
	// err := client.Start(ctx, "172.20.30.92", 5060)
	err = client.Start(ctx, *serverAddr)
	if err != nil {
		panic(err)
	}

	fmt.Println("呼叫", *to)
	dl, err := client.Call(*to)
	if err != nil {
		panic(err)
	}

	defer dl.Hangup()

	for {
		select {
		case <-ctx.Done():
			dl.Hangup()
			return
		case state := <-dl.State():
			if state == dialog.Answered {
				sp := dl.SDP()
				for _, media := range sp.MediaDescriptions {
					fmt.Println(media.MediaName.Media, sp.Origin.UnicastAddress, media.MediaName.Port.Value)
				}

				for _, media := range sp.MediaDescriptions {
					if media.MediaName.Media == "audio" {
						stop := Audio2RTP(ctx, "./test.wav", fmt.Sprintf("rtp://%s:%d", sp.Origin.UnicastAddress, media.MediaName.Port.Value))
						<-stop
						dl.Hangup()
					}
				}
			}
			if state == dialog.Hangup {
				cancel()
				fmt.Println("Hangup")
				return
			}
			if state == dialog.Error {
				fmt.Println("Error", dl.StatusCode(), dl.Reason())
				return
			}
		}
	}
}

// 使用 ffmpeg 推送
func Audio2RTP(ctx context.Context, audioUrl string, rtpURI string) chan bool {
	stopNotice := make(chan bool, 1)
	args := []string{"-re", "-i", audioUrl, "-vn", "-c:a", "pcm_alaw", "-f", "alaw", "-ac", "1", "-ar", "8000",
		"-f", "rtp", rtpURI,
	}

	ffmpegCMD := cmd.NewCmd("ffmpeg", args...)
	statusChan := ffmpegCMD.Start()
	go func() {
		defer func() {
			fmt.Println("AudioPlay end")
			stopNotice <- true
		}()
		for {
			select {
			case <-ctx.Done():
				ffmpegCMD.Stop()
			case _ = <-statusChan:
			case cmdErr := <-ffmpegCMD.Stderr:
				fmt.Println("cmdErr:", cmdErr)
			case _ = <-ffmpegCMD.Done():
				fmt.Println("ffmpeg live CMD.Done()")
				return
			}
		}
	}()
	return stopNotice
}

// 使用直接推送
func Wav2RTP(wavpath string, rtpaddress string) {
	file, _ := os.Open(wavpath)
	wavReader := wav.NewReader(file)

	conn, err := net.Dial("udp", rtpaddress)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	seq := rtp.NewRandomSequencer()
	ssrc := rand.Uint32()
	payloader := &codecs.G711Payloader{}
	const ulawSamplingRate = 8000 // ulaw sampling rate
	var pt uint8 = 0

	packetrizer := rtp.NewPacketizer(1200, pt, ssrc, payloader, seq, ulawSamplingRate)

	data := make([]byte, 4096)

	// 1/8000 = 125 mciroseconds
	// data byte * 125

	tickDuration := time.Microsecond * time.Duration(len(data)*125)
	ticker := time.NewTicker(tickDuration)
	for ; true; <-ticker.C {
		l, err := wavReader.Read(data)
		if err == io.EOF {
			fmt.Println("end")
			break
		} else if err != nil {
			logrus.Errorf("Could not read the sample. err: %v", err)
			return
		}

		// packetize to the RTP
		packets := packetrizer.Packetize(data, uint32(l))
		for _, packet := range packets {
			b, err := packet.Marshal()
			if err != nil {
				continue
			}

			_, err = conn.Write(b)
			if err != nil {
				logrus.Errorf("Could not send the rtp correctly. err: %v", err)
			}
		}
	}

}
