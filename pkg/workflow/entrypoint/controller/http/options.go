package http

const DefaultWorkflowRoutePrefix = "/workflows"

type Option interface {
	GetPrefix() string
}

type options struct {
	Prefix string
}

func (o *options) GetPrefix() string {
	return o.Prefix
}

type prefixOption string

func (c prefixOption) apply(opts *options) {
	opts.Prefix = string(c)
}

func WithPrefix(p string) OptionApply {
	return prefixOption(p)
}

type OptionApply interface {
	apply(*options)
}

func GetOptions(opts ...OptionApply) Option {
	op := &options{
		Prefix: DefaultWorkflowRoutePrefix,
	}

	for _, o := range opts {
		o.apply(op)
	}

	return op
}
