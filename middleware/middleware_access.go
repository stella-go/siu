// Copyright 2010-2025 the original author or authors.

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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/interfaces"
)

const (
	AccessMiddleDisableKey = "middleware.access.disable"
	AccessMaxLengthKey     = "middleware.access.max-length"
	AccessMiddleOrder      = 10
)

type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *CustomResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

type MiddlewareAccess struct {
	Conf      config.TypedConfig `@siu:"name='environment',default='type'"`
	Logger    interfaces.Logger  `@siu:"name='logger',default='type'"`
	debug     bool
	maxLength int
}

func (p *MiddlewareAccess) Init() {
	debug := p.Conf.GetStringOr("logger.level", "info")
	if strings.ToLower(debug) == "debug" {
		p.debug = true
	}
	p.maxLength = p.Conf.GetIntOr(AccessMaxLengthKey, 2048)
}

func (p *MiddlewareAccess) Condition() bool {
	if v, ok := p.Conf.GetBool(AccessMiddleDisableKey); ok && v {
		return false
	}
	return true
}

func (p *MiddlewareAccess) Function() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !p.debug {
			start := time.Now()
			method := c.Request.Method
			path := c.Request.URL.Path
			query := c.Request.URL.RawQuery
			if query != "" {
				path = path + "?" + query
			}

			c.Next()

			latency := time.Since(start) / time.Millisecond
			ip := c.ClientIP()
			size := c.Writer.Size()
			status := c.Writer.Status()
			printLogger(p.Logger.INFO, "%s %3d %s %s %dms %dbytes", method, status, path, ip, latency, size)
		} else {
			start := time.Now()
			method := c.Request.Method
			path := c.Request.URL.Path
			query := c.Request.URL.RawQuery
			if query != "" {
				path = path + "?" + query
			}
			proto := c.Request.Proto
			headers := p.headerString(c.Request.Header)
			bts, _ := io.ReadAll(c.Request.Body)
			sb := &strings.Builder{}
			if len(bts) > 0 {
				s := string(bts)
				if p.isPrintable(s) {
					if len(bts) > p.maxLength {
						s := fmt.Sprintf("\n=============::Request::=============\n%s %s %s\n\n%s\n%s...\n", method, path, proto, headers, bts[:p.maxLength])
						sb.WriteString(s)
					} else {
						s := fmt.Sprintf("\n=============::Request::=============\n%s %s %s\n\n%s\n%s\n", method, path, proto, headers, bts)
						sb.WriteString(s)
					}
				} else {
					s := fmt.Sprintf("\n=============::Request::=============\n%s %s %s\n\n%s\n<binary data>\n", method, path, proto, headers)
					sb.WriteString(s)
				}
			} else {
				s := fmt.Sprintf("\n=============::Request::=============\n%s %s %s\n\n%s\n", method, path, proto, headers)
				sb.WriteString(s)
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bts))
			writer := &CustomResponseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
			c.Writer = writer

			c.Next()

			latency := time.Since(start) / time.Millisecond
			status := c.Writer.Status()
			statusText := http.StatusText(status)
			headers = p.headerString(c.Writer.Header())
			ip := c.ClientIP()
			size := c.Writer.Size()

			bts = writer.body.Bytes()
			if len(bts) > 0 {
				s := string(bts)
				if p.isPrintable(s) {
					if len(bts) > p.maxLength {
						s := fmt.Sprintf("=============::Response::============\n%s %d %s\n\n%s\n%s...\n", proto, status, statusText, headers, bts[:p.maxLength])
						sb.WriteString(s)
					} else {
						s := fmt.Sprintf("=============::Response::============\n%s %d %s\n\n%s\n%s\n", proto, status, statusText, headers, bts)
						sb.WriteString(s)
					}
				} else {
					s := fmt.Sprintf("=============::Response::============\n%s %d %s\n\n%s\n<binary data>\n", proto, status, statusText, headers)
					sb.WriteString(s)
				}
			} else {
				s := fmt.Sprintf("=============::Response::============\n%s %d %s\n\n%s\n", proto, status, statusText, headers)
				sb.WriteString(s)
			}
			sb.WriteString("=============::End::=================")

			if _, ok := c.Get(ContextResourceKey); !ok {
				printLogger(p.Logger.DEBUG, "%s", sb.String())
			}
			printLogger(p.Logger.INFO, "%s %3d %s %s %dms %dbytes", method, status, path, ip, latency, size)
		}
	}
}

func (p *MiddlewareAccess) Order() int {
	return AccessMiddleOrder
}

func (p *MiddlewareAccess) headerString(header http.Header) string {
	sb := &strings.Builder{}
	for k, va := range header {
		if len(va) > 0 {
			for _, v := range va {
				sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
			}
		} else {
			sb.WriteString(fmt.Sprintf("%s: %s\n", k, va))
		}
	}
	return sb.String()
}

func (p *MiddlewareAccess) isPrintable(s string) bool {
	max := len(s)
	if max == 0 {
		return true
	}
	not_print := 0
	for _, r := range s {
		if !unicode.IsPrint(r) {
			not_print++
		}
		if float64(not_print)/float64(max) > 0.2 {
			return false
		}
	}
	return float64(not_print)/float64(max) <= 0.2
}
