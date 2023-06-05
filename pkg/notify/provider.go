package notify

type Provider interface {
	AddSMSProvider(provider BySMS)
	AddEmailProvider(provider ByEmail)
	HasProvider(tn TypeNotify, name string) bool
	GetProviders(tn TypeNotify) map[string]interface{}
	GetProvider(tn TypeNotify, name string) interface{}
}

type provider struct {
	data map[TypeNotify]map[string]interface{}
}

func (p *provider) AddSMSProvider(provider BySMS) {
	if len(p.data[TypeSMS]) == 0 {
		p.data[TypeSMS] = map[string]interface{}{
			provider.UID(): provider,
		}
	}

	p.data[TypeSMS][provider.UID()] = provider
}

func (p *provider) AddEmailProvider(provider ByEmail) {
	if len(p.data[TypeEmail]) == 0 {
		p.data[TypeEmail] = map[string]interface{}{
			provider.UID(): provider,
		}
	}

	p.data[TypeEmail][provider.UID()] = provider
}

func (p *provider) HasProvider(tn TypeNotify, name string) bool {
	if _, ok := p.data[tn][name]; !ok {
		return false
	}

	return true
}

func (p *provider) GetProviders(tn TypeNotify) map[string]interface{} {
	return p.data[tn]
}

func (p *provider) GetProvider(tn TypeNotify, name string) interface{} {
	return p.data[tn][name]
}

func New() Provider {
	return &provider{
		data: make(map[TypeNotify]map[string]interface{}),
	}
}
