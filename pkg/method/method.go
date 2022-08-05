package method

type Method string

const (
	REGISTER  Method = "REGISTER"
	INVITE    Method = "INVITE"
	ACK       Method = "ACK"
	BYE       Method = "BYE"
	CANCEL    Method = "CANCEL"
	UPDATE    Method = "UPDATE"
	REFER     Method = "REFER"
	PRACK     Method = "PRACK"
	SUBSCRIBE Method = "SUBSCRIBE"
	NOTIFY    Method = "NOTIFY"
	PUBLISH   Method = "PUBLISH"
	MESSAGE   Method = "MESSAGE"
	INFO      Method = "INFO"
	OPTIONS   Method = "OPTIONS"
)
