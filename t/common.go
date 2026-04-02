// Copyright 2010-2026 the original author or authors.

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

// Common content type constants for RouteDef.ContentType.
const (
	ContentTypeJSON           = "application/json"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeMultipart      = "multipart/form-data"
)

type RequestBean[T any] struct {
	Timestamp int64 `json:"timestamp" @meta:"description=unix timestamp in milliseconds,required"`
	Data      T     `json:"data" @meta:"description=the actual request data"`
}

func (s *RequestBean[T]) String() string {
	return fmt.Sprintf("RequestBean{Timestamp: %d, Data: %+v}", s.Timestamp, s.Data)
}

type ResultBean[T any] struct {
	Code    int    `json:"code" @meta:"description=status code (200 for success, non-200 for error)"`
	Message string `json:"message" @meta:"description=status message"`
	Data    T      `json:"data" @meta:"description=the actual response data"`
}

func (s *ResultBean[T]) String() string {
	return fmt.Sprintf("ResultBean{Code: %d, Message: %s, Data: %+v}", s.Code, s.Message, s.Data)
}

type PageableResult[T any] struct {
	Count int `json:"count" @meta:"description=total count of items"`
	List  []T `json:"list" @meta:"description=list of items"`
}

func (s *PageableResult[T]) String() string {
	return fmt.Sprintf("PageableResult{Count: %d, List: %+v}", s.Count, s.List)
}
