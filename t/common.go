// Copyright 2010-2024 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	Data    T      `json:"data"`
}

func (s *ResultBean[T]) String() string {
	return fmt.Sprintf("ResultBean{Code: %d, Message: %s, Data: %+v}", s.Code, s.Message, s.Data)
}
