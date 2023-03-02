package main

import (
	"fmt"

	"github.com/pion/sdp/v3"
)

func main() {
	sd := sdp.SessionDescription{}

	sd.MediaDescriptions = append(sd.MediaDescriptions, &sdp.MediaDescription{
		MediaName: sdp.MediaName{
			Media: "video",
			Port: sdp.RangedPort{
				Value: 3088,
			},
			Formats: []string{"96", "98", "97"},
		},

		Attributes: []sdp.Attribute{
			sdp.NewAttribute("rtpmap", "96 PS/90000"),
			sdp.NewAttribute("rtpmap", "98 H264/90000"),
			sdp.NewAttribute("rtpmap", "97 MPEG4/90000"),
		},
	})

	aa, _ := sd.Marshal()

	fmt.Println(string(aa))
}
