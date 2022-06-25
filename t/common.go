package t

import "fmt"

type RequestBean[T any] struct {
	Timestamp int64 `json:"timestamp"`
	Data      T     `json:"data"`
}

func (s *RequestBean[T]) String() string {
	return fmt.Sprintf("RequestBean{Timestamp: %d, Data: %+v}", s.Timestamp, s.Data)
}

type ResultBean[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

func (s *ResultBean[T]) String() string {
	return fmt.Sprintf("ResultBean{Code: %d, Message: %s, Data: %+v}", s.Code, s.Message, s.Data)
}
