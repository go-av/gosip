package ptz_test

import (
	"fmt"
	"testing"

	"github.com/go-av/gosip/pkg/utils/ptz"
)

func TestAA(t *testing.T) {
	fmt.Println(ptz.PTZCmd(ptz.LeftDown, 0, 0))
	// A5 0F 01 06 00 00 00 0B
	//,A50F01067D7D00B5
}
