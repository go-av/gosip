package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/log"
	"github.com/go-av/gosip/pkg/sdp"
	"github.com/go-av/gosip/pkg/sip"
	"github.com/go-cmd/cmd"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/sirupsen/logrus"
	"github.com/youpy/go-wav"
)

func main() {
	log.EnablePrintMSG(true)
	client := sip.NewClient("蜗牛", "snail_out", "abc", "172.20.30.52", 5063)
	client.SetSDP(func() *sdp.SDP {
		body := "v=0\r\n"
		body += "o=- 3868331676 3868331676 IN IP4 172.20.30.52\r\n"
		body += "s=Gosip 1.0.0 (MacOSX)\r\n"
		body += "t=0 0\r\n"
		body += "m=audio 50006 RTP/AVP 8 0 101\r\n"
		body += "c=IN IP4 172.20.30.52\r\n"
		body += "a=rtcp:50007\r\n"
		body += "a=rtpmap:8 PCMA/8000\r\n"
		body += "a=rtpmap:0 PCMU/8000\r\n"
		body += "a=rtpmap:101 telephone-event/8000\r\n"
		body += "a=sendrecv\r\n"
		sd, _ := sdp.ParseSDP(body)
		return sd
	})
	ctx, _ := context.WithCancel(context.Background())
	client.Start(ctx, "udp", "172.20.50.12", 5060)
	time.Sleep(1 * time.Second)
	fmt.Println("呼叫")
	dl, err := client.Call("snail_in")
	if err != nil {
		fmt.Println("err", err)
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
						Wav2RTP("./test.wav", fmt.Sprintf("%s:%d", sp.Origin.UnicastAddress, media.MediaName.Port.Value))
						// stop := Audio2RTP(ctx, "./test.wav", fmt.Sprintf("rtp://%s:%d", sp.Origin.UnicastAddress, media.MediaName.Port.Value))
						// <-stop
						dl.Hangup()
					}
				}
			}
			if state == dialog.Hangup {
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
