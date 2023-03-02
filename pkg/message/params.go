package message

import (
	"bytes"
	"strings"
	"sync"
)

type Params struct {
	params sync.Map
}

func NewParams() *Params {
	return &Params{}
}
func (p *Params) Get(key string) (string, bool) {
	value, ok := p.params.Load(key)
	if !ok {
		return "", false
	}
	return value.(string), true
}

func (p *Params) Set(key string, value string) *Params {
	p.params.Store(key, value)
	return p
}

func (p *Params) Del(key string) *Params {
	p.params.Delete(key)
	return p
}

func (p *Params) Length() int {
	i := 0
	p.params.Range(func(key, value any) bool {
		i++
		return true
	})
	return i
}

func (p *Params) ToString(sep string) string {
	var buf = bytes.NewBuffer(nil)
	first := true
	p.params.Range(func(key, value any) bool {
		if !first {
			buf.WriteString(sep)
		}
		first = false
		if value.(string) != "" {
			buf.WriteString(key.(string) + "=" + value.(string))
		} else {
			buf.WriteString(key.(string))
		}

		return true
	})
	return buf.String()
}

func (p *Params) Clone() *Params {
	new := &Params{}
	p.params.Range(func(key, value any) bool {
		new.Set(key.(string), value.(string))
		return true
	})
	return new
}

func ParseParams(data string) *Params {
	data = strings.TrimSpace(data)
	param := NewParams()
	if data == "" {
		return param
	}
	list := strings.Split(data, ";")
	for _, p := range list {
		kv := strings.Split(p, "=")
		if len(kv) == 1 {
			param.Set(kv[0], "")
		} else {
			param.Set(kv[0], kv[1])
		}
	}
	return param
}
