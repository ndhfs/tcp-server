package tcp

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	defaultReadTimeout  = 30 * time.Second
	defaultWriteTimeout = 30 * time.Second

	defaultWorkersNum        = 100
	defaultWorkerWaitTimeout = 15 * time.Second

	defaultReadBufferSize = 4096

	defaultAcceptThreshold = 50
)

type Options struct {
	name              string
	processor         Processor
	readTimeout       time.Duration
	writeTimeout      time.Duration
	logger            Logger
	debug             bool
	workersNum        int
	workerWaitTimeout time.Duration
	acceptThreshold   int
	encoder           Encoder
	prom              prometheus.Registerer
	rl                RateLimiterFactory
}

type Option func(options *Options)

var defaultOptions = Options{
	name:              "default",
	readTimeout:       defaultReadTimeout,
	writeTimeout:      defaultWriteTimeout,
	logger:            NewStdLogger(),
	workersNum:        defaultWorkersNum,
	workerWaitTimeout: defaultWorkerWaitTimeout,
	acceptThreshold:   defaultAcceptThreshold,
	processor:         NewRowSocketProcessor(),
	rl:                NewUnlimitedRateLimitFactory(),
}

//WithReadTimeout sets read deadline for income connects
func WithReadTimeout(duration time.Duration) Option {
	return func(options *Options) {
		options.readTimeout = duration
	}
}

func WithWriteTimeout(duration time.Duration) Option {
	return func(options *Options) {
		options.writeTimeout = duration
	}
}

func WithLogger(l Logger) Option {
	return func(options *Options) {
		options.logger = l
	}
}

func WithProcessor(processor Processor) Option {
	return func(options *Options) {
		options.processor = processor
	}
}

func WithDebugMode(on bool) Option {
	return func(options *Options) {
		options.debug = on
	}
}

func WithEncoder(enc Encoder) Option {
	return func(options *Options) {
		options.encoder = enc
	}
}

func WithPrometheus(prom prometheus.Registerer) Option {
	return func(options *Options) {
		options.prom = prom
	}
}

func WithRateLimiter(rl RateLimiterFactory) Option {
	return func(options *Options) {
		options.rl = rl
	}
}
