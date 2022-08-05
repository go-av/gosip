package message

func NewContentTypeHeader(contentType string) *ContentTypeHeader {
	l := ContentTypeHeader(contentType)
	return &l
}

type ContentTypeHeader string

func (ct *ContentTypeHeader) Name() string {
	return "Content-Type"
}

func (ct ContentTypeHeader) Value() string {
	return string(ct)
}

func (ct *ContentTypeHeader) Clone() Header {
	return ct
}

func init() {
	defaultHeaderParsers.Register(NewContentTypeHeader(""))
}

func (ContentTypeHeader) Parse(data string) (Header, error) {
	return NewContentTypeHeader(data), nil
}
