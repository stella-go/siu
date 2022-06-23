// Copyright 2010-2022 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/interfaces"
	"github.com/stella-go/siu/t/stackerror"
)

const (
	ErrorlogMiddleDisableKey = "middleware.error-log.disable"
	ErrorlogMiddleOrder      = 30
)

var (
	err500 = errors.New("internal server error")
)

type MiddlewareErrorlog struct {
	Conf   config.TypedConfig `@siu:"name='environment',default='type'"`
	Logger interfaces.Logger  `@siu:"name='logger',default='type'"`
}

func (p *MiddlewareErrorlog) Condition() bool {
	if v, ok := p.Conf.GetBool(ErrorlogMiddleDisableKey); ok && v {
		return false
	}
	return true
}

func (p *MiddlewareErrorlog) Function() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				err = stackerror.NewError(3, err.(error))
				p.Logger.ERROR("", err)
				c.AbortWithError(500, err500)
			}
		}()
		c.Next()
	}
}
func (p *MiddlewareErrorlog) Order() int {
	return ErrorlogMiddleOrder
}
