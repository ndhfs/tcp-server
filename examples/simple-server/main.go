package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ndhfs/tcp-server"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	s := tcp.New(
		tcp.WithReadTimeout(15 * time.Second),
		tcp.WithDebugMode(true),

	)

	s.SetHandler(newMessageHandler)
	s.SetErrorHandler(func(ctx tcp.Conn, err error) {
		ctx.Send("Err " + err.Error())
		ctx.Close()
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
