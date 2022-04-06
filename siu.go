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

package siu

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/logger"
	"github.com/stella-go/siu/autoconfig"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/middleware"
)

type Logger interface {
	DEBUG(format string, arr ...interface{})
	INFO(format string, arr ...interface{})
	WARN(format string, arr ...interface{})
	ERROR(format string, arr ...interface{})
}

type LeveledLogger interface {
	Logger
	Level() logger.Level
}

type Router interface {
	Router() map[string]gin.HandlerFunc
}

type MiddlewareRouter interface {
	Router
	Middleware() []gin.HandlerFunc
}

var ctx *context
var once sync.Once

// export method

func LoadConfig(files ...string) {
	config.LoadConfig(files...)
}

func New(environment config.TypedConfig, contextLogger Logger, server *gin.Engine) {
	once.Do(func() {
		ctx = newContext(environment, contextLogger, server)
	})
}

func NewWithEnvironment(environment config.TypedConfig) {
	once.Do(func() {
		ctx = newEnvironmentContext(environment)
	})
}

func Default() {
	once.Do(func() {
		ctx = newDefaultContext()
	})
}

func DEBUG(format string, arr ...interface{}) {
	Default()
	ctx.DEBUG(format, arr...)
}

func INFO(format string, arr ...interface{}) {
	Default()
	ctx.INFO(format, arr...)
}

func WARN(format string, arr ...interface{}) {
	Default()
	ctx.WARN(format, arr...)
}

func ERROR(format string, arr ...interface{}) {
	Default()
	ctx.ERROR(format, arr...)
}

func AutoFactory(auto ...autoconfig.AutoFactory) {
	Default()
	ctx.AutoFactory(auto...)
}

func Use(middleware ...middleware.OrderedMiddleware) {
	Default()
	ctx.Use(middleware...)
}

func Route(router ...Router) {
	Default()
	ctx.Route(router...)
}

func Get(key string) (interface{}, bool) {
	Default()
	return ctx.Get(key)
}

func Set(key string, value interface{}) {
	Default()
	ctx.Set(key, value)
}

func Run() {
	Default()
	ctx.Run()
}
