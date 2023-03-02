package sdp

import (
	"github.com/pion/sdp/v3"
	"github.com/sirupsen/logrus"
)

func ParseSDP(str []byte) (*SDP, error) {
	sd := &SDP{}
	err := sd.Unmarshal([]byte(str))
	if err != nil {
		return nil, err
	}
	return sd, nil
}

func NewSDP() *SDP {
	return &SDP{}
}

type SDP sdp.SessionDescription

func (SDP) ContentType() string {
	return "application/sdp"
}

func (sd *SDP) Body() []byte {
	data, err := (*sdp.SessionDescription)(sd).Marshal()
	if err != nil {
		logrus.Error(err)
		return nil
	}
	return data
}

func (sd *SDP) Marshal() string {
	data, err := (*sdp.SessionDescription)(sd).Marshal()
	if err != nil {
		logrus.Error(err)
		return ""
	}
	return string(data)
}

func (sd *SDP) Unmarshal(value []byte) error {
	return (*sdp.SessionDescription)(sd).Unmarshal(value)
}
