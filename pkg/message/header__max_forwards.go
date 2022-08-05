package message

import (
	"fmt"
	"strconv"
)

func NewMaxForwardsHeader(maxForwards uint32) *MaxForwardsHeader {
	l := MaxForwardsHeader(maxForwards)
	return &l
}

type MaxForwardsHeader uint32

func (maxForwards *MaxForwardsHeader) Name() string {
	return "Max-Forwards"
}

func (maxForwards MaxForwardsHeader) Value() string {
	return fmt.Sprintf("%d", maxForwards)
}

func (maxForwards *MaxForwardsHeader) Clone() Header {
	return maxForwards
}

func init() {
	defaultHeaderParsers.Register(NewMaxForwardsHeader(0))
}

func (MaxForwardsHeader) Parse(data string) (Header, error) {
	num, _ := strconv.ParseUint(data, 10, 64)
	return NewMaxForwardsHeader(uint32(num)), nil
}
