package util

import (
	"encoding/json"
)

type EncoderDecoder[T any] interface {
	Encode(value T) ([]byte, error)
	Decode(data []byte) (*T, error)
}

type JsonEncDec[T any] struct{}

var _ EncoderDecoder[any] = new(JsonEncDec[any])

func NewJsonEncoderDecoder[T any]() *JsonEncDec[T] {
	return &JsonEncDec[T]{}
}
func (encdec *JsonEncDec[T]) Encode(value T) ([]byte, error) {
	res, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (encdec *JsonEncDec[T]) Decode(data []byte) (*T, error) {
	var res T
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
