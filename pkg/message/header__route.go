package message

func NewRouteHeader(router string) *RouteHeader {
	ua := RouteHeader(router)
	return &ua
}

type RouteHeader string

func (router *RouteHeader) Name() string {
	return "Route"
}

func (router RouteHeader) Value() string {
	return string(router)
}

func (ua *RouteHeader) Clone() Header {
	return ua
}

func init() {
	defaultHeaderParsers.Register(NewRouteHeader(""))
}

func (RouteHeader) Parse(data string) (Header, error) {
	return NewRouteHeader(data), nil
}
