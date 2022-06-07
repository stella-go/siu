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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/config"
)

const (
	CROSMiddleDisableKey = "middleware.cros.disable"
	CROSMiddleOrder      = 10
)

type MiddlewareCROS struct {
	Conf config.TypedConfig `@siu:"name='environment',default='type'"`
}

func (p *MiddlewareCROS) Condition() bool {
	if v, ok := p.Conf.GetBool(CROSMiddleDisableKey); ok && v {
		return false
	}
	return true
}

func (p *MiddlewareCROS) Function() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		headers := c.GetHeader("Access-Control-Allow-Headers")
		if headers == "" {
			headers = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", headers)
		c.Header("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "*")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		c.Next()
	}
}
func (p *MiddlewareCROS) Order() int {
	return CROSMiddleOrder
}
