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
	listener          Listener
	readTimeout       time.Duration
	writeTimeout      time.Duration
	logger            Logger
	debug             bool
	workersNum        int
	workerWaitTimeout time.Duration

	acceptThreshold int

	reader Reader
	writer Writer

	decoders []Decoder
}

var defaultOptions = Options{
	listener:          NewSimpleListener(),
	readTimeout:       defaultReadTimeout,
	writeTimeout:      defaultWriteTimeout,
	logger:            NewStdLogger(),
	workersNum:        defaultWorkersNum,
	workerWaitTimeout: defaultWorkerWaitTimeout,
	reader:            NewLineReader(defaultReadBufferSize),
	writer:            NewLineWriter(),
	acceptThreshold:   defaultAcceptThreshold,
	decoders:          []Decoder{NewByteToStringConverter()},
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

func WithReader(reader Reader) Option {
	return func(options *Options) {
		options.reader = reader
	}
}

func WithWriter(writer Writer) Option {
	return func(options *Options) {
		options.writer = writer
	}
}

func WithListener(l Listener) Option {
	return func(options *Options) {
		options.listener = l
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
