package websocket

import (
	"fmt"
	"github.com/gobwas/ws"
	"io"
)

type Reader struct {

}

func NewReader() *Reader {
	return &Reader{}
}

func (r *Reader) Read(in io.Reader) ([]byte, error) {
	hdr, err := ws.ReadHeader(in)
	if err != nil {
		return nil, fmt.Errorf("failed read websocket header. %w", err)
	}

	payload := make([]byte, hdr.Length)
	_, err = io.ReadFull(in, payload)
	if err != nil {
		// handle error
	}
	if hdr.Masked {
		ws.Cipher(payload, hdr.Mask, 0)
	}
	return payload, err
}
