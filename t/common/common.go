package common

import "github.com/stella-go/siu/t/n"

type RequestBean[T any] struct {
	Timestamp *n.Int64 `json:"timestamp"`
	Data      T        `json:"data"`
}

type ResultBean[T any] struct {
	Code    *n.Int    `json:"code"`
	Message *n.String `json:"message"`
	Data    T         `json:"data,omitempty"`
}
