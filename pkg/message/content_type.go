package message

var ContentTypeXML = ContentType("Application/MANSCDP+xml")

type ContentType string

const (
	ContentType__MANSCDP_XML ContentType = "Application/MANSCDP+xml"
	ContentType__SDP         ContentType = "application/sdp"
)
