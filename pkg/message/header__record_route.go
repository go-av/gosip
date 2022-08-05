package message

func NewRecordRouteHeader(route string) *RecordRouteHeader {
	ua := RecordRouteHeader(route)
	return &ua
}

type RecordRouteHeader string

func (ua *RecordRouteHeader) Name() string {
	return "Record-Route"
}

func (route RecordRouteHeader) Value() string {
	return string(route)
}

func (route *RecordRouteHeader) Clone() Header {
	return route
}

func init() {
	defaultHeaderParsers.Register(NewRecordRouteHeader(""))
}

func (RecordRouteHeader) Parse(data string) (Header, error) {
	return NewRecordRouteHeader(data), nil
}
