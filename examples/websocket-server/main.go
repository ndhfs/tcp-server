package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ndhfs/tcp-server"
	"github.com/ndhfs/tcp-server/processors/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
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
		tcp.WithPrometheus(prometheus.DefaultRegisterer),
	)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		go http.ListenAndServe(":8081", nil)

	}()

	s.SetHandler(createMessageHandler(s), func(handler tcp.Handler) tcp.Handler {
		return func(c tcp.Context, m tcp.Msg) error {
			c.Set("username", "Mike")
			return handler(c, m)
		}
	})
	s.Subscribe(func(e tcp.EventType, c tcp.Context) error {
		log.Println("EVENT: ", e, c.ID())
		return nil
	})
	s.SetErrorHandler(func(c tcp.Context, err error) {
		c.Send("Err " + err.Error())
		c.Close()
	})

	defer func() {
		ctx, cancelFn := context.WithTimeout(context.Background(), 20 * time.Second)
		defer cancelFn()
		s.GracefulShutdown(ctx)
	}()
	go func() {
		if err := s.Serve(context.Background(), "tcp4", ":8082"); err != nil {
			panic(fmt.Errorf("failed server tcp server. %w", err))
		}
	}()

	var sigCh = make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGQUIT, syscall.SIGINT)

	<-sigCh
}

func createMessageHandler(s *tcp.Server) tcp.Handler {
	return func(c tcp.Context, m tcp.Msg) error {
		data := m.(string)
		log.Println("NEW MESSAGE", data)

		cmd := strings.TrimSpace(data)

		switch cmd {
		case "wait20":
			c.Send("waiting")
			username := c.Get("username").(string)
			go func(cid string) {
				select {
				case <-time.After(20 * time.Second):
					s.Send(cid, "waiting done; " + username)
				case <-c.Context().Done():
					fmt.Println("waiting aborted")
				}
			}(c.ID())
		case "lock20":
			time.Sleep(20 * time.Second)
		case "close":
			return c.Close()
		case "close_w_err":
			return c.CloseWithErr(errors.New("closed with error"))
		case "err":
			return errors.New("test error")
		case "panic":
			panic("I am panic")
		}

		return nil
	}
}

