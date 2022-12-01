package gen

type (
	options struct {
		patterns  []string
		tags      []string
		recursive bool
		deps      []string
	}
	Option     interface{ apply(*options) }
	optionFunc func(g *options)
)

func (f optionFunc) apply(o *options) {
	f(o)
}

func DefaultOptions() options {
	return options{
		patterns:  []string{"."},
		tags:      []string{},
		deps:      []string{},
		recursive: true,
	}
}

func WithPatterns(patterns ...string) Option {
	return optionFunc(
		func(o *options) {
			patterns = filterEmptyStr(patterns...)
			if len(patterns) > 0 {
				o.patterns = patterns
			}
		})
}

func WithTags(tags ...string) Option {
	return optionFunc(
		func(o *options) {
			tags = filterEmptyStr(tags...)
			if len(tags) > 0 {
				o.tags = tags
			}
		})
}

func WithRecursive(recursive bool) Option {
	return optionFunc(
		func(o *options) {
			o.recursive = recursive
		})
}

func WithDeps(deps ...string) Option {
	return optionFunc(
		func(o *options) {
			deps = filterEmptyStr(deps...)
			if len(deps) > 0 {
				o.deps = deps
			}
		})
}
