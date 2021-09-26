package tcp

type Msg interface {}

type Encoder interface {
	Encode(Msg) (Msg, error)
	Decode(Msg) (Msg, error)
}

type StringEncodeDecoder struct {
}

func NewByteToStringConverter() *StringEncodeDecoder {
	return &StringEncodeDecoder{}
}

func (s *StringEncodeDecoder) Encode(v Msg) (Msg, error) {
	return []byte(v.(string)), nil
}

func (s *StringEncodeDecoder) Decode(m Msg) (Msg, error) {
	return string(m.([]byte)), nil
}

