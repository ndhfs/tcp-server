package server

import "fmt"

type Msg interface {}

type Encoder interface {
	Encode(v Msg) ([]byte, error)
	Decode([]byte) (Msg, error)
}

type StringEncodeDecoder struct {
}

func NewStringEncodeDecoder() *StringEncodeDecoder {
	return &StringEncodeDecoder{}
}

func (s *StringEncodeDecoder) Encode(v Msg) ([]byte, error) {
	switch vv := v.(type) {
	case []byte:
		return vv, nil
	case string:
		return []byte(vv), nil
	default:
		return nil, fmt.Errorf("failed encode message. %w", ErrInvalidPackage)
	}
}

func (s *StringEncodeDecoder) Decode(d []byte) (Msg, error) {
	return string(d), nil
}

