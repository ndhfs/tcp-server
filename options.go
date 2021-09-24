package tcp

import "time"

const (
	defaultReadTimeout  = 30 * time.Second
	defaultWriteTimeout = 30 * time.Second

	defaultWorkersNum        = 100
	defaultWorkerWaitTimeout = 15 * time.Second

	defaultReadBufferSize = 4096

	defaultAcceptThreshold = 50
)

type Options struct {
	processor         Processor
	readTimeout       time.Duration
	writeTimeout      time.Duration
	logger            Logger
	debug             bool
	workersNum        int
	workerWaitTimeout time.Duration
	acceptThreshold   int
	decoders          []Decoder
}

type Option func(options *Options)

var defaultOptions = Options{
	readTimeout:       defaultReadTimeout,
	writeTimeout:      defaultWriteTimeout,
	logger:            NewStdLogger(),
	workersNum:        defaultWorkersNum,
	workerWaitTimeout: defaultWorkerWaitTimeout,
	acceptThreshold:   defaultAcceptThreshold,
	decoders:          []Decoder{NewByteToStringConverter()},
	processor:         NewRowSocketProcessor(),
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

func WithDecoders(encoders ...Decoder) Option {
	return func(options *Options) {
		options.decoders = encoders
	}
}
