package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/sdp"
	"github.com/go-av/gosip/pkg/sip"
	"github.com/go-cmd/cmd"
)

func main() {
	// log.EnablePrintMSG(true)
	client := sip.NewClient("蜗牛", "snail", "abc", "172.20.50.12", 5062)
	client.SetSDP(func() *sdp.SDP {
		body := "v=0\r\n"
		body += "o=- 3868331676 3868331676 IN IP4 172.20.30.52\r\n"
		body += "s=Gosip 1.0.0 (MacOSX)\r\n"
		body += "t=0 0\r\n"
		body += "m=audio 50006 RTP/AVP 9 8 0 101\r\n"
		body += "c=IN IP4 172.20.30.52\r\n"
		body += "a=rtcp:50007\r\n"
		body += "a=rtpmap:8 PCMA/8000\r\n"
		body += "a=rtpmap:0 PCMU/8000\r\n"
		body += "a=rtpmap:101 telephone-event/8000\r\n"
		body += "a=sendrecv\r\n"
		sd, _ := sdp.ParseSDP(body)
		return sd
	})
	ctx, cancel := context.WithCancel(context.Background())
	client.Start(ctx, "udp", "172.20.50.12", 5060)
	time.Sleep(1 * time.Second)
	fmt.Println("呼叫")
	_ = cancel
	dl, err := client.Call("snail")
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
			fmt.Println("dl1 state=======", state)
			if state == dialog.Answered {
				sp := dl.SDP()
				for _, media := range sp.MediaDescriptions {
					if media.MediaName.Media == "audio" {
						rtpURI := fmt.Sprintf("rtp://%s:%d", sp.Origin.UnicastAddress, media.MediaName.Port.Value)
						stop := Audio2RTP(ctx, "./test.wav", rtpURI)
						<-stop
						dl.Hangup()
					}
				}
			}
			if state == dialog.Hangup {
				fmt.Println("Hangup")
				return
			}
			if state == dialog.Error {
				fmt.Println("Error")
				fmt.Println(dl.Reason())
				fmt.Println(dl.StatusCode())
				return
			}
		}
	}
}

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
