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

type SocketHandler struct {

}

func (s *SocketHandler) MessageReceived(c server.Conn, _data server.Msg) error {
	data := _data.(string)
	log.Println("NEW MESSAGE", data)

	cmd := strings.TrimSpace(data)

	switch cmd {
	case "wait20":
		c.Send([]byte(`waiting`))
		go func() {
			select {
			case <-time.After(20 * time.Second):
				c.Send([]byte(`waiting done\n`))
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

func (s *SocketHandler) ConnectionClosed(c server.Conn) error {
	panic("implement me")
}

func main() {
	s := server.New(
		server.WithReadTimeout(15 * time.Second),
		server.WithDebugMode(true),

	)

	s.SetHandler(new(SocketHandler))
	s.SetErrorHandler(func(ctx server.Conn, err error) {
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

