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

package middleware

import (
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/interfaces"
)

const (
	RewriteMiddleDisableKey = "middleware.rewrite.disable"
	RewriteMiddleMatchKey   = "middleware.rewrite.match"
	RewriteMiddleRewriteKey = "middleware.rewrite.rewrite"
	RewriteMiddleOrder      = 5
)

type MiddlewareRewrite struct {
	Server      *gin.Engine        `@siu:"name='server',default='type'"`
	Conf        config.TypedConfig `@siu:"name='environment',default='type'"`
	Logger      interfaces.Logger  `@siu:"name='logger',default='type'"`
	rewriteFunc func(string) (bool, string)
}

func (p *MiddlewareRewrite) Init() {
	match, ok1 := p.Conf.GetString(RewriteMiddleMatchKey)
	rewrite, ok2 := p.Conf.GetString(RewriteMiddleRewriteKey)
	if ok1 && ok2 && match != "" && rewrite != "" {
		re := regexp.MustCompile(match)
		p.rewriteFunc = func(s string) (bool, string) {
			if re.MatchString(s) {
				return true, re.ReplaceAllString(s, rewrite)
			}
			return false, s
		}
	} else {
		p.rewriteFunc = func(s string) (bool, string) {
			return false, s
		}
	}
}

func (p *MiddlewareRewrite) Condition() bool {
	if v := p.Conf.GetBoolOr(RewriteMiddleDisableKey, true); v {
		return false
	}
	return true
}

func (p *MiddlewareRewrite) Function() gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := c.Request.URL.Path
		if match, s := p.rewriteFunc(uri); match {
			p.Logger.DEBUG("request path rewrite: %s -> %s", uri, s)
			c.Request.URL.Path = s
			c.Request.RequestURI = s
			p.Server.HandleContext(c)
			c.Abort()
		}
	}
}

func (p *MiddlewareRewrite) Order() int {
	return RewriteMiddleOrder
}
