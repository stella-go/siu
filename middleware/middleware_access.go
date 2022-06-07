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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/interfaces"
)

const (
	AccessMiddleDisableKey = "middleware.access.disable"
	AccessMiddleOrder      = 9999
)

type MiddlewareAccess struct {
	Conf   config.TypedConfig `@siu:"name='environment',default='type'"`
	Logger interfaces.Logger  `@siu:"name='logger',default='type'"`
}

func (p *MiddlewareAccess) Condition() bool {
	if v, ok := p.Conf.GetBool(AccessMiddleDisableKey); ok && v {
		return false
	}
	return true
}

func (p *MiddlewareAccess) Function() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		raw := c.Request.URL.RawQuery
		path := c.Request.URL.Path
		if raw != "" {
			path = path + "?" + raw
		}
		c.Next()
		latency := time.Now().Sub(start) / time.Millisecond
		ip := c.ClientIP()
		method := c.Request.Method
		status := c.Writer.Status()
		size := c.Writer.Size()
		p.Logger.INFO("%-4s %3d %s %s %dms %dB", method, status, path, ip, latency, size)
	}
}
func (p *MiddlewareAccess) Order() int {
	return AccessMiddleOrder
}
