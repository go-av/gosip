package types

import (
	"strconv"
)

type Port int

func (port *Port) String() string {
	if port == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*port), 10)
}
