package utils

import (
	"bytes"
	"encoding/xml"

	"golang.org/x/net/html/charset"
)

func XMLDecode(data []byte, v interface{}) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charset.NewReaderLabel
	return decoder.Decode(v)
}
