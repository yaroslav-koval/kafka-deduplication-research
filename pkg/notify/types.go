package notify

import "context"

type Status interface {
	GetStatus(ctx context.Context, req MessageStatusRequest) (*StatusResponse, error)
}

type UID interface {
	UID() string
}

type ByEmail interface {
	UID
	Status
	WithBaseURL(baseURL string)
	From(from string) ByEmail
	WithSubject(subject string) ByEmail
	UseHTML() ByEmail
	WithBody(body string) ByEmail
	WithTag(tag string) ByEmail
	WithTrackOpens() ByEmail
	To(to string, cc ...string) ByEmail
	Send(ctx context.Context) (*MessageResponse, error)
}

type BySMS interface {
	UID
	Status
	From(from string) BySMS
	WithBody(body string) BySMS
	To(to string, cc ...string) BySMS
	Send(ctx context.Context) (*MessageResponse, error)
}

type TypeNotify string

func (tn TypeNotify) ToString() string {
	return string(tn)
}
