package tcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ndhfs/gopool"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	opts         Options
	l            net.Listener
	runCtx       context.Context
	doneFn       func()
	wg           sync.WaitGroup
	mu           sync.RWMutex
	conns        map[string]*connContext
	workerPool   *gopool.GoPool
	shuttingDown bool

	handler      Handler
	subscriber   Subscriber
	middlewares  []Middleware
	errorHandler ErrorHandler
	metrics      *metricsManger
}

func (s *Server) Subscribe(subscriber Subscriber) {
	s.subscriber = subscriber
}

func (s *Server) SetErrorHandler(errorHandler ErrorHandler) {
	s.errorHandler = errorHandler
}

func (s *Server) Serve(ctx context.Context, network string, addr string) error {
	s.logInfo("Start serving at %s %s", network, addr)
	defer func() {
		s.logInfo("Stop serving at %s %s", network, addr)
	}()
	var err error

	s.metrics.init(s.opts)

	s.mu.Lock()
	s.l, err = s.opts.processor.Listen(ctx, network, addr)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed start listener. %w", err)
	}
	s.mu.Unlock()

	acceptPool := gopool.NewGoPool(s.opts.acceptThreshold)

	for {
		conn, err := s.l.Accept()
		if err != nil {
			select {
			case <-s.runCtx.Done():
				return nil
			default:
				var operr *net.OpError
				if errors.As(err, &operr) && (operr.Temporary() || operr.Timeout()) {
					s.logDebug("failed accept conn. %s", err)
				}
				s.logErr(fmt.Sprintf("failed accept conn. %s", err))
				continue
			}
		}

		err = acceptPool.Do(func() {
			s.metrics.acceptCount.Inc()
			c := s.newContext(conn)
			if err := s.handleEvent(EventTypeAccepted, c); err != nil {
				return
			}
			go s.handleConnection(c)
		})

		if err != nil {
			s.metrics.acceptDeclineInc()
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func (s *Server) newContext(conn net.Conn) *connContext {
	c := &connContext{
		id:        uuid.New().String(),
		conn:      conn,
		createdAt: time.Now(),
		s:         s,
		store:     make(Map),
	}

	c.doneCtx, c.doneFn = context.WithCancel(s.runCtx)

	s.mu.Lock()
	s.conns[c.id] = c
	s.mu.Unlock()
	s.metrics.connectionsInc()
	return c
}

func (s *Server) Send(cid string, m Msg) error {
	conn, err := s.getConn(cid)
	if err != nil {
		return fmt.Errorf("failed find connect by id %s. %w", cid, err)
	}
	return conn.Send(m)
}

func (s *Server) CloseConn(cid string) error {
	conn, err := s.getConn(cid)
	if err != nil {
		return fmt.Errorf("failed find connect by id %s. %w", cid, err)
	}
	return conn.Close()
}

func (s *Server) CloseConnWithError(cid string, err error) error {
	conn, err := s.getConn(cid)
	if err != nil {
		return fmt.Errorf("failed find connect by id %s. %w", cid, err)
	}
	return conn.CloseWithErr(err)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logInfo("Shutting down tcp server")
	defer func() {
		s.logInfo("Shutting down tcp server complete")
	}()

	s.mu.Lock()
	if s.shuttingDown {
		s.mu.Unlock()
		return nil
	}
	s.shuttingDown = true
	s.mu.Unlock()

	// signal all connections for shutting down
	s.doneFn()

	// stop accept new conns
	s.l.Close()

	return nil
}

func (s *Server) GracefulShutdown(ctx context.Context) error {
	s.logInfo("Graceful shutting down tcp server")
	defer func() {
		s.logInfo("Graceful shutting down tcp server complete")
	}()

	s.mu.Lock()
	if s.shuttingDown {
		s.mu.Unlock()
		return nil
	}
	s.shuttingDown = true
	s.mu.Unlock()

	// signal all connections for shutting down
	s.doneFn()

	// stop accept new conns
	s.l.Close()

	var waitCh = make(chan struct{})
	go func() {
		s.logInfo("waiting active connections closed. Num cons: %d", s.NumConns())
		s.wg.Wait()
		close(waitCh)
	}()

	for {
		select {
		case <-waitCh:
			s.logInfo("waiting active connections closed. Success")
			return nil
		case <-time.Tick(5 * time.Second):
			s.logInfo("waiting active connections closed. Num cons: %d", s.NumConns())
		case <-ctx.Done():
			s.logErr("waiting active connections closed. Timout expired. Force closed")
			return nil
		}
	}
}

func (s *Server) handleConnection(c *connContext) {
	s.wg.Add(1)
	s.logDebug("New connection %s", c.conn.RemoteAddr())
	defer func() {
		c.Close()
		s.mu.Lock()
		delete(s.conns, c.id)
		s.mu.Unlock()
		s.metrics.connectionsDec()
		s.logDebug("Connection %s closed.", c.conn.RemoteAddr())
		s.wg.Done()
	}()

	for {
		// Если завершились, больше не слушать
		select {
		case <-c.doneCtx.Done():
			return
		default:
		}

		if s.opts.readTimeout > 0 {
			c.conn.SetReadDeadline(time.Now().Add(s.opts.readTimeout))
		}

		b, err := s.opts.processor.Read(c.conn)

		// Если завершились, не обрабатываем
		select {
		case <-c.doneCtx.Done():
			return
		default:
		}

		if err != nil {
			var erop *net.OpError
			if errors.Is(err, io.EOF) {
			} else if errors.As(err, &erop) && (erop.Timeout()) {
			} else {
				s.logErr("Err read conn. %s", err)
			}
			return
		}

		if err := s.dispatch(c, b); err != nil {
			s.handleError(c, err)
		}
	}
}

func New(opts ...Option) *Server {
	var o = defaultOptions
	for _, opt := range opts {
		opt(&o)
	}

	ctx, doneFn := context.WithCancel(context.Background())
	return &Server{
		opts:       o,
		runCtx:     ctx,
		doneFn:     doneFn,
		conns:      make(map[string]*connContext, 100),
		workerPool: gopool.NewGoPool(o.workersNum),
		metrics:    new(metricsManger),
	}
}

func (s *Server) SetHandler(handler Handler, middlewares ...Middleware) {
	s.handler = handler
	s.middlewares = middlewares
}

func (s *Server) logErr(format string, v ...interface{}) {
	s.opts.logger.Printf("[ERR] "+format, v...)
}

func (s *Server) logDebug(format string, v ...interface{}) {
	if s.opts.debug {
		s.opts.logger.Printf("[DBG] "+format, v...)
	}
}

// logInfo
func (s *Server) logInfo(format string, a ...interface{}) {
	s.opts.logger.Printf("[INF] "+format, a...)
}

func (s *Server) NumConns() interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.conns)
}

func (s *Server) dispatch(c *connContext, b []byte) error {
	err := s.workerPool.DoWithTimeout(s.opts.workerWaitTimeout, func() {
		s.metrics.workersIdleDec()
		defer func() {
			s.metrics.workersIdleInc()
		}()
		s.dispatchAsync(c, b)
	})
	if err != nil {
		return ErrServerIsBusy
	}
	return nil
}

func (s *Server) handleError(c *connContext, err error) {
	if s.errorHandler != nil {
		s.errorHandler(c, err)
	} else {
		s.logErr("failed dispatch client message. %s", err)
	}
}

func (s *Server) dispatchAsync(c *connContext, b []byte) {
	defer func() {
		if err := recover(); err != nil {
			switch er := err.(type) {
			case error:
				s.handleError(c, er)
			case string:
				s.handleError(c, errors.New(er))
			default:
				s.handleError(c, fmt.Errorf("unexpected error: %v", er))
			}
		}
	}()
	s.logDebug("Received msg: %s", string(b))
	if s.handler == nil {
		s.logInfo("Handler not registered. Skipping")
		return
	}

	var msg Msg = b
	if enc := c.s.opts.encoder; enc != nil {
		var err error
		msg, err = enc.Decode(msg)
		if err != nil {
			s.handleError(c, fmt.Errorf("failed decode message. %w", ErrInvalidPackage))
			return
		}
	}

	if s.handler == nil {
		s.handleError(c, ErrHandlerNotRegistered)
		return
	}

	handler := s.handler
	for _, middleware := range s.middlewares {
		handler = middleware(handler)
	}

	if err := handler(c, msg); err != nil {
		s.handleError(c, err)
		return
	}
}

func (s *Server) handleEvent(e EventType, c *connContext) error {
	if s.subscriber != nil {
		return s.subscriber(e, c)
	}
	return nil
}

func (s *Server) getConn(cid string) (*connContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if c, ok := s.conns[cid]; ok {
		return c, nil
	}
	return nil, ErrUndefinedConn
}
