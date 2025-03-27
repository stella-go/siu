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

package siu

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/common"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/interfaces"
)

var ctx *context
var once sync.Once

// export method

func LoadConfig(files ...string) {
	config.LoadConfig(files...)
}

func New(environment config.TypedConfig, contextLogger interfaces.Logger, server *gin.Engine) {
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
	if strings.Contains(format, "__sLINE__") {
		file, line := stack()
		format = strings.ReplaceAll(format, "__sLINE__", fmt.Sprintf("__LINE:%s:%d__", file, line))
	}
	if ctx == nil {
		common.DEBUG(format, arr...)
	} else {
		ctx.DEBUG(format, arr...)
	}
}

func INFO(format string, arr ...interface{}) {
	if strings.Contains(format, "__sLINE__") {
		file, line := stack()
		format = strings.ReplaceAll(format, "__sLINE__", fmt.Sprintf("__LINE:%s:%d__", file, line))
	}
	if ctx == nil {
		common.INFO(format, arr...)
	} else {
		ctx.INFO(format, arr...)
	}
}

func WARN(format string, arr ...interface{}) {
	if strings.Contains(format, "__sLINE__") {
		file, line := stack()
		format = strings.ReplaceAll(format, "__sLINE__", fmt.Sprintf("__LINE:%s:%d__", file, line))
	}
	if ctx == nil {
		common.WARN(format, arr...)
	} else {
		ctx.WARN(format, arr...)
	}
}

func ERROR(format string, arr ...interface{}) {
	if strings.Contains(format, "__sLINE__") {
		file, line := stack()
		format = strings.ReplaceAll(format, "__sLINE__", fmt.Sprintf("__LINE:%s:%d__", file, line))
	}
	if ctx == nil {
		common.ERROR(format, arr...)
	} else {
		ctx.ERROR(format, arr...)
	}
}

func RegisterBean(name string, typ reflect.Type, obj interface{}) {
	Default()
	ctx.RegisterBean(name, typ, obj)
}

func GetBeanByName(name string) (interface{}, bool) {
	Default()
	return ctx.GetBeanByName(name)
}

func GetBeanByType(typ reflect.Type) (interface{}, bool) {
	Default()
	return ctx.GetBeanByType(typ)
}

func Register(registers ...interfaces.InjectRegister) {
	Default()
	ctx.Register(registers...)
}

func AutoFactory(auto ...interfaces.AutoFactory) {
	Default()
	ctx.AutoFactory(auto...)
}

func Use(middleware ...interfaces.OrderedMiddleware) {
	Default()
	ctx.Use(middleware...)
}

func Route(router ...interfaces.Router) {
	Default()
	ctx.Route(router...)
}

func Shutdown(shutdown ...interfaces.ShutdownHook) {
	Default()
	ctx.Shutdown(shutdown...)
}

func Forward(c *gin.Context, path string) {
	Default()
	ctx.Forward(c, path)
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

func stack() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "", 0
	}
	file = file[strings.LastIndex(file, "/")+1:]
	return file, line
}
