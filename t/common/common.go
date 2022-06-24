package common

type RequestBean[T any] struct {
	Timestamp int64 `json:"timestamp"`
	Data      T     `json:"data"`
}

type ResultBean[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}
