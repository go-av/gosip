package message

func NewDateHeader(date string) *DateHeader {
	id := DateHeader(date)
	return &id
}

type DateHeader string

func (date *DateHeader) Name() string {
	return "Date"
}

func (date DateHeader) Value() string {
	return string(date)
}

func (date *DateHeader) Clone() Header {
	return date
}

func init() {
	defaultHeaderParsers.Register(NewDateHeader(""))
}

func (DateHeader) Parse(date string) (Header, error) {
	return NewDateHeader(date), nil
}
