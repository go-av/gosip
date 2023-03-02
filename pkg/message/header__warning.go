package message

func NewWarningHeader(warning string) *WarningHeader {
	w := WarningHeader(warning)
	return &w
}

type WarningHeader string

func (w *WarningHeader) Name() string {
	return "Warning"
}

func (w WarningHeader) Value() string {
	return string(w)
}

func (w *WarningHeader) Clone() Header {
	return w
}

func init() {
	defaultHeaderParsers.Register(NewWarningHeader(""))
}

func (WarningHeader) Parse(data string) (Header, error) {
	return NewWarningHeader(data), nil
}
