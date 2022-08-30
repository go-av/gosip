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
	client := sip.NewClient("蜗牛", "snail_in", "abc", "172.20.30.52", 5060)
	ctx, _ := context.WithCancel(context.Background())
	client.Start(ctx, "udp", "172.20.50.12", 5060)
	time.Sleep(1 * time.Second)
	client.SetSDP(func(*sdp.SDP) *sdp.SDP {
		str := `v=0
o=- 3868331676 3868331676 IN IP4 172.20.30.52
s=gosip (MacOSX)
t=0 0
m=audio 50006 RTP/AVP 8 0 101
c=IN IP4 172.20.30.52
a=rtcp:50007
a=rtpmap:8 PCMA/8000
a=rtpmap:0 PCMU/8000
a=rtpmap:101 telephone-event/8000
m=video 50006 RTP/AVP 96
c=IN IP4 172.20.30.52
a=rtcp:50009
a=rtpmap:96 VP8/90000
`
		sd, err := sdp.ParseSDP(str)
		if err != nil {
			fmt.Println("err=====", err)
		}
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
			fmt.Println("sdp", dl.SDP())
			if state == dialog.Answered {
				go func() {
					time.Sleep(10 * time.Second)
					dl.Hangup()
				}()
			}
			if state == dialog.Hangup {
				displayName, _ := dl.User()
				fmt.Printf("结束与%s通话\n", displayName)
				return
			}
		}
	}
}
