package tiniyo

import "time"

type Options struct {
	HostAPI        string
	AuthID         string
	AuthToken      string
	RequestTimeout time.Duration

	From string
	To   string
	Body string
}

type MessageRequestBody struct {
	AddressRetention     string `json:"AddressRetention,omitempty"`
	ApplicationSID       string `json:"ApplicationSid,omitempty"`
	Attempt              string `json:"Attempt,omitempty"`
	Body                 string `json:"Body"`
	ContentRetention     string `json:"ContentRetention,omitempty"`
	ForceDelivery        string `json:"ForceDelivery,omitempty"`
	From                 string `json:"From"`
	MaxPrice             string `json:"MaxPrice,omitempty"`
	MediaURL             string `json:"MediaUrl,omitempty"`
	MessagingServiceSID  string `json:"MessagingServiceSid,omitempty"`
	PeID                 string `json:"PeId,omitempty"`
	PersistentAction     string `json:"PersistentAction,omitempty"`
	ProvideFeedback      string `json:"ProvideFeedback,omitempty"`
	SmartEncoded         string `json:"SmartEncoded,omitempty"`
	StatusCallback       string `json:"StatusCallback,omitempty"`
	StatusCallbackMethod string `json:"StatusCallbackMethod,omitempty"`
	TemplateID           string `json:"TemplateId,omitempty"`
	To                   string `json:"To"`
	ValidityPeriod       string `json:"ValidityPeriod,omitempty"`
}

type MessageResponseBody struct {
	AccountSID          string           `json:"account_sid,omitempty"`
	APIVersion          string           `json:"api_version,omitempty"`
	Body                string           `json:"body,omitempty"`
	DateCreated         string           `json:"date_created,omitempty"`
	DateSent            string           `json:"date_sent,omitempty"`
	DateUpdated         string           `json:"date_updated,omitempty"`
	Direction           string           `json:"direction,omitempty"`
	ErrorCode           string           `json:"error_code,omitempty"`
	ErrorMessage        string           `json:"error_message,omitempty"`
	From                string           `json:"from,omitempty"`
	MessagingServiceSID string           `json:"messaging_service_sid,omitempty"`
	NumMedia            string           `json:"num_media,omitempty"`
	NumSegments         string           `json:"num_segments,omitempty"`
	Price               string           `json:"price,omitempty"`
	PriceUnit           string           `json:"price_unit,omitempty"`
	SID                 string           `json:"sid,omitempty"`
	Status              string           `json:"status,omitempty"`
	SubresourceURIs     *SubresourceURIs `json:"subresource_uris,omitempty"`
	To                  string           `json:"to,omitempty"`
	URI                 string           `json:"uri,omitempty"`
}

type SubresourceURIs struct {
	Media string `json:"media,omitempty"`
}

type MessageErrorResponseBody struct {
	Message  string `json:"message,omitempty"`
	Code     int    `json:"code,omitempty"`
	Status   int    `json:"status,omitempty"`
	MoreInfo string `json:"more_info,omitempty"`
}
