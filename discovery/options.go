package discovery

type Options struct {
	ServiceType string
	ServiceName string
	RawKey      string
}

func (options Options) String() string {
	if options.RawKey != "" {
		return options.RawKey
	}

	if options.ServiceType == "*" && options.ServiceName == "*" {
		return "*"
	}

	return options.ServiceType + ":" + options.ServiceName + ":*"
}

type funcDiscoveryOption struct {
	f func(*Options)
}

func (fdo *funcDiscoveryOption) Apply(do *Options) {
	fdo.f(do)
}

type Option interface {
	Apply(*Options)
}

func TypeOption(t string) Option {
	return &funcDiscoveryOption{
		f: func(options *Options) {
			if t == "" {
				t = "*"
			}
			options.ServiceType = t
		},
	}
}

func NameOption(t string) Option {
	return &funcDiscoveryOption{
		f: func(options *Options) {
			if t == "" {
				t = "*"
			}
			options.ServiceName = t
		},
	}
}

func RawKeyOption(t string) Option {
	return &funcDiscoveryOption{
		f: func(options *Options) {
			options.RawKey = t
		},
	}
}
