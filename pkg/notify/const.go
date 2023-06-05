package notify

const (
	TypeEmail TypeNotify = "email"
	TypeSMS   TypeNotify = "sms"

	ProviderPostmark = "postmark"
	ProviderTiniyo   = "tiniyo"
	ProviderMailgun  = "mailgun"
	ProviderTwilio   = "twilio"

	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusSent       = "SENT"
	StatusFailed     = "FAILED"
	StatusDelivered  = "DELIVERED"
	StatusSeen       = "SEEN"
	StatusReplied    = "REPLIED"
	StatusQueue      = "QUEUED"
)
