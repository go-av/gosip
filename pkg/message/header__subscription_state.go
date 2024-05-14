package message

func NewSubscriptionStateHeader(state string) *SubscriptionStateHeader {
	id := SubscriptionStateHeader(state)
	return &id
}

type SubscriptionStateHeader string

func (SubscriptionStateHeader) Name() string {
	return "Subscription-State"
}

func (state SubscriptionStateHeader) Value() string {
	return string(state)
}

func (state *SubscriptionStateHeader) Clone() Header {
	return state
}

func init() {
	defaultHeaderParsers.Register(NewSubscriptionStateHeader(""))
}

func (SubscriptionStateHeader) Parse(data string) (Header, error) {
	return NewSubscriptionStateHeader(data), nil
}
