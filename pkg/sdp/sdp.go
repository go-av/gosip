package sdp

import "github.com/pion/sdp/v2"

func ParseSDP(data string) (*SDP, error) {
	sd := &SDP{}
	err := sd.Unmarshal([]byte(data))
	if err != nil {
		return nil, err
	}
	return sd, nil
}

func NewSDP() *SDP {
	return &SDP{}
}

type SDP struct {
	sdp.SessionDescription
}

func (SDP) ContentType() string {
	return "application/sdp"
}

func (sdp *SDP) Body() string {
	data, err := sdp.Marshal()
	if err != nil {
		return ""
	}
	return string(data)
}
