package message

type Body interface {
	Data() []byte
	ContentType() string
}

func NewBody(contentType string, data []byte) Body {
	return &body{
		contentType: contentType,
		data:        data,
	}
}

type body struct {
	data        []byte
	contentType string
}

func (b *body) Data() []byte {
	return b.data
}

func (b *body) ContentType() string {
	return b.contentType
}
