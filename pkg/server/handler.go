package server

import "github.com/go-av/gosip/pkg/message"

type ServerHandler interface {
	SetServer(Server)
	GetClient(user string) (Client, error)
	Realm() string
	ReceiveMessage(Content) (int, string)
}

type Server interface {
	Send(protocol string, address string, msg message.Message) error
	SendMessage(client Client, content Content) (Content, error)
	ServerAddress() *message.Address
}

type Content interface {
	Data() []byte
	ContentType() string
}

func NewContent(contentType string, data []byte) Content {
	return &content{
		contentType: contentType,
		data:        data,
	}
}

type content struct {
	data        []byte
	contentType string
}

func (b *content) Data() []byte {
	return b.data
}

func (b *content) ContentType() string {
	return b.contentType
}
