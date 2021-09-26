package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ndhfs/tcp-server"
	"github.com/ndhfs/tcp-server/examples/processors/websocket"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type SuffixEncoder struct {
	suffix string
}

func NewSuffixDecoder(suffix string) *SuffixEncoder {
	return &SuffixEncoder{suffix: suffix}
}

func (s *SuffixEncoder) Encode(msg tcp.Msg) (tcp.Msg, error) {
	return append(msg.([]byte), []byte(s.suffix)...), nil
}

func (s *SuffixEncoder) Decode(msg tcp.Msg) (tcp.Msg, error) {
	return []byte(strings.TrimSuffix(string(msg.([]byte)), s.suffix)), nil
}

func main() {

	s := tcp.New(
		tcp.WithProcessor(websocket.NewProcessor(
			websocket.WithBackoffProcessor(tcp.NewRowSocketProcessor()),
		)),
		tcp.WithReadTimeout(15 * time.Second),
		tcp.WithDebugMode(true),
		tcp.WithEncoder(tcp.NewByteToStringConverter()),
	)

	s.SetHandler(newMessageHandler, func(handler tcp.Handler) tcp.Handler {
		return func(c tcp.Conn, m tcp.Msg) error {
			log.Println(" -> ", m)
			return handler(c, m)
		}
	})
	s.Subscribe(func(e tcp.EventType, c tcp.Conn) error {
		log.Println("EVENT: ", e, c.ID())
		return nil
	})
	s.SetErrorHandler(func(c tcp.Conn, err error) {
		c.Send("Err " + err.Error())
		c.Close()
	})

	defer func() {
		ctx, cancelFn := context.WithTimeout(context.Background(), 20 * time.Second)
		defer cancelFn()
		s.GracefulShutdown(ctx)
	}()
	go func() {
		if err := s.Serve(context.Background(), "tcp4", ":8080"); err != nil {
			panic(fmt.Errorf("failed server tcp server. %w", err))
		}
	}()

	var sigCh = make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGQUIT, syscall.SIGINT)

	<-sigCh
}

func newMessageHandler(c tcp.Conn, m tcp.Msg) error {
	data := m.(string)
	log.Println("NEW MESSAGE", data)

	cmd := strings.TrimSpace(data)

	switch cmd {
	case "wait20":
		c.Send("waiting")
		go func() {
			select {
			case <-time.After(20 * time.Second):
				c.Send("waiting done")
			case <-c.Context().Done():
				fmt.Println("waiting aborted")
			}
		}()
	case "close":
		return c.Close()
	case "err":
		return errors.New("test error")
	}

	return nil
}

