package message

func NewSubjectHeader(sub string) *SubjectHeader {
	ua := SubjectHeader(sub)
	return &ua
}

type SubjectHeader string

func (sub *SubjectHeader) Name() string {
	return "Subject"
}

func (sub SubjectHeader) Value() string {
	return string(sub)
}

func (sub *SubjectHeader) Clone() Header {
	return sub
}

func init() {
	defaultHeaderParsers.Register(NewSubjectHeader(""))
}

func (SubjectHeader) Parse(data string) (Header, error) {
	return NewSubjectHeader(data), nil
}
