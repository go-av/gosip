package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-av/gosip/pkg/dialog"
	"github.com/go-av/gosip/pkg/log"
	"github.com/go-av/gosip/pkg/sdp"
	"github.com/go-av/gosip/pkg/sip"
)

func main() {
	log.EnablePrintMSG(true)
	client := sip.NewClient("蜗牛", "snail_in", "abc", "172.20.30.52", 5062)
	ctx, _ := context.WithCancel(context.Background())
	client.Start(ctx, "udp", "172.20.50.12", 5060)
	time.Sleep(1 * time.Second)
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
	for {
		select {
		case dl := <-client.Dialog():
			go doDialog(dl)
		}
	}
}

func doDialog(dl dialog.Dialog) {
	fmt.Println("收到：")
	user, _ := dl.User()
	fmt.Println("user:", user)
	dl.SetState(dialog.Ringing)
	time.Sleep(2 * time.Second)
	dl.SetState(dialog.Answered)
	for {
		select {
		case state := <-dl.State():
			fmt.Println("in state", state)
			if state == dialog.Answered {
				go func() {
					// time.Sleep(10 * time.Second)
					// dl.Hangup()
				}()
			}
			if state == dialog.Hangup {
				return
			}
		}
	}
}
