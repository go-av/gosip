package message

type ContentType string

const (
	ContentType__MANSCDP_XML ContentType = "Application/MANSCDP+xml"
	ContentType__SDP         ContentType = "application/sdp"
	ContentType__XML         ContentType = "application/xml"
)
