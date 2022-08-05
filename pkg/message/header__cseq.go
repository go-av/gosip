package message

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-av/gosip/pkg/method"
)

func NewCSeqHeader(seqNo uint32, me method.Method) *CSeqHeader {
	return &CSeqHeader{
		SeqNo:  seqNo,
		Method: me,
	}
}

type CSeqHeader struct {
	SeqNo  uint32
	Method method.Method
}

func (cseq *CSeqHeader) Name() string { return "CSeq" }

func (cseq *CSeqHeader) Value() string {
	return fmt.Sprintf("%d %s", cseq.SeqNo, cseq.Method)
}

func (cseq *CSeqHeader) Clone() Header {
	if cseq == nil {
		var newCSeq *CSeqHeader
		return newCSeq
	}

	return &CSeqHeader{
		SeqNo:  cseq.SeqNo,
		Method: cseq.Method,
	}
}

func init() {
	defaultHeaderParsers.Register(&CSeqHeader{})
}

func (CSeqHeader) Parse(data string) (Header, error) {
	list := strings.Split(data, " ")
	if len(list) != 2 {
		return nil, errors.New("CSeq is illegal")
	}
	seqNo, _ := strconv.ParseUint(list[0], 10, 64)
	return &CSeqHeader{
		SeqNo:  uint32(seqNo),
		Method: method.Method(list[1]),
	}, nil
}
