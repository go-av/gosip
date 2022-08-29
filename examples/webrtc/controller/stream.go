package controller

import (
	"fmt"
	"io"
	"sync"

	reuse "github.com/libp2p/go-reuseport"
)

func NewStreamMgr(ip string, port int) *StreamMgr {
	mgr := &StreamMgr{}
	go mgr.run(ip, port)
	return mgr
}

type StreamMgr struct {
	streams sync.Map
}

func (mgr *StreamMgr) run(ip string, port int) {
	fmt.Println("ListenPacket:", fmt.Sprintf("%s:%d", ip, port))
	conn, err := reuse.ListenPacket("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 20480)
	i := 0
	for {
		i++
		n, addr, err := conn.ReadFrom(buf)
		if err == nil {
			mgr.LoadOrCreate(addr.String()).write(buf[:n])
		}
	}
}

func (mgr *StreamMgr) LoadOrCreate(address string) *stream {
	if s, ok := mgr.streams.Load(address); ok {
		return s.(*stream)
	}

	s := &stream{}

	mgr.streams.Store(address, s)
	return s
}

type stream struct {
	writer io.Writer
}

func (s *stream) write(buf []byte) {
	if s.writer != nil {
		_, err := s.writer.Write(buf)
		if err != nil {
			fmt.Println("writer err", err)
		}
	}
}

func (s *stream) SetWriter(writer io.Writer) {
	s.writer = writer
}
