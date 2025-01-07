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
	"io"
	"log"
	"os"
	"os/signal"
	"reflect"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/logger"
	"github.com/stella-go/siu/autoconfig"
	"github.com/stella-go/siu/common"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/inject"
	"github.com/stella-go/siu/interfaces"
	"github.com/stella-go/siu/middleware"
)

const (
	defaultBanner = `
	     _______. __   __    __  
	    /       ||  | |  |  |  | 
	   |   (----'|  | |  |  |  | 
	    \   \    |  | |  |  |  | 
	.----)   |   |  | |  '--'  | 
	|_______/    |__|  \______/  
	        Version %s
`
)

const (
	loggerEnvKey             = "logger"
	loggerUseEnvKey          = loggerEnvKey + ".siu"
	loggerTagEnvKey          = loggerEnvKey + ".tag"
	loggerLevelEnvKey        = loggerEnvKey + ".level"
	loggerPatternEnvKey      = loggerEnvKey + ".pattern"
	loggerDaliyEnvKey        = loggerEnvKey + ".daliy"
	loggerPathEnvKey         = loggerEnvKey + ".path"
	loggerFileEnvKey         = loggerEnvKey + ".file"
	loggerMaxFilesEnvKey     = loggerEnvKey + ".maxFiles"
	loggerMaxFileSizesEnvKey = loggerEnvKey + ".maxFileSize"

	BuildinRegisterOrder = 0
)

type buildinLogger struct {
	l        *log.Logger
	logLevel logger.Level
	tag      string
}

func newBuildinLogger(logLevel logger.Level, tag string, writer io.Writer) *buildinLogger {
	l := log.New(writer, "", log.LstdFlags)
	return &buildinLogger{l: l, logLevel: logLevel, tag: tag}
}

func (p *buildinLogger) DEBUG(format string, arr ...interface{}) {
	if p.logLevel <= logger.DebugLevel {
		if len(arr) > 0 {
			if _, ok := arr[len(arr)-1].(error); ok {
				format += " %v"
			}
		}
		msg := fmt.Sprintf(format, arr...)
		p.l.Printf("DEBUG - %s %s", p.tag, msg)
	}
}

func (p *buildinLogger) INFO(format string, arr ...interface{}) {
	if p.logLevel <= logger.InfoLevel {
		if len(arr) > 0 {
			if _, ok := arr[len(arr)-1].(error); ok {
				format += " %v"
			}
		}
		msg := fmt.Sprintf(format, arr...)
		p.l.Printf("INFO  - %s %s", p.tag, msg)
	}
}

func (p *buildinLogger) WARN(format string, arr ...interface{}) {
	if p.logLevel <= logger.WarnLevel {
		if len(arr) > 0 {
			if _, ok := arr[len(arr)-1].(error); ok {
				format += " %v"
			}
		}
		msg := fmt.Sprintf(format, arr...)
		p.l.Printf("WARN  - %s %s", p.tag, msg)
	}
}

func (p *buildinLogger) ERROR(format string, arr ...interface{}) {
	if p.logLevel <= logger.ErrorLevel {
		if len(arr) > 0 {
			if _, ok := arr[len(arr)-1].(error); ok {
				format += " %v"
			}
		}
		msg := fmt.Sprintf(format, arr...)
		p.l.Printf("ERROR - %s %s", p.tag, msg)
	}
}

func (p *buildinLogger) Level() logger.Level {
	return p.logLevel
}

func (p *buildinLogger) Tag() string {
	return p.tag
}

type context struct {
	environment config.TypedConfig
	logger      interfaces.Logger

	registers     []interfaces.InjectRegister
	auto          []interfaces.AutoFactory
	middleware    []interfaces.OrderedMiddleware
	routers       []interfaces.Router
	shutdownHooks []interfaces.ShutdownHook

	store *sync.Map

	server *gin.Engine
}

func newContext(environment config.TypedConfig, contextLogger interfaces.Logger, server *gin.Engine) *context {
	ctx := &context{environment, contextLogger, make([]interfaces.InjectRegister, 0), make([]interfaces.AutoFactory, 0), make([]interfaces.OrderedMiddleware, 0), make([]interfaces.Router, 0), make([]interfaces.ShutdownHook, 0), &sync.Map{}, server}
	if leveledLogger, ok := contextLogger.(interfaces.LeveledLogger); ok {
		common.SetLevel(leveledLogger.Level())
	}
	if tagedLogger, ok := contextLogger.(interfaces.TagedLogger); ok {
		common.SetTag(tagedLogger.Tag())
	}
	ctx.Register(&buildinRegister{ctx})
	ctx.AutoFactory(&autoconfig.AutoMysql{}, &autoconfig.AutoGorm{}, &autoconfig.AutoRedis{}, &autoconfig.AutoZookeeper{})
	ctx.Use(&middleware.MiddlewareRewrite{}, &middleware.MiddlewareAccess{}, &middleware.MiddlewareCROS{}, &middleware.MiddlewareErrorlog{}, &middleware.MiddlewareResource{}, &middleware.MiddlewareSession{}, &middleware.MiddlewareJwt{})
	return ctx
}

func newEnvironmentContext(environment config.TypedConfig) *context {
	logUse := environment.GetBoolOr(loggerUseEnvKey, true)
	tag := environment.GetStringOr(loggerTagEnvKey, "[SIU]")
	logLevel := logger.Parse(environment.GetStringOr(loggerLevelEnvKey, "info"))
	logPattern := environment.GetStringOr(loggerPatternEnvKey, "%d{06-01-02.15:04:05.000} [%g] %p %c - %m")
	daily := environment.GetBoolOr(loggerDaliyEnvKey, true)
	filePath := environment.GetStringOr(loggerPathEnvKey, ".")
	fileName := environment.GetStringOr(loggerFileEnvKey, "stdout")
	maxFiles := environment.GetIntOr(loggerMaxFilesEnvKey, 30)
	maxFileSize := environment.GetIntOr(loggerMaxFileSizesEnvKey, 200)

	cfg := &logger.RotateConfig{
		Enable:      true,
		Daily:       daily,
		MaxFiles:    maxFiles,
		MaxFileSize: int64(maxFileSize) * logger.FileSizeM,
		FilePath:    filePath,
		FileName:    fileName,
	}
	writer, err := logger.NewConfigRotateWriter(cfg)
	if err != nil {
		panic(err)
	}
	var contextLogger interfaces.Logger
	if logUse {
		contextLogger = logger.NewRootLogger(logLevel, &logger.PatternFormatter{Pattern: logPattern}, writer).GetLogger(tag)
	} else {
		contextLogger = newBuildinLogger(logLevel, tag, writer)
		common.INFO("use buildin logger")
	}

	ctx := newContext(environment, contextLogger, nil)
	return ctx
}

func newDefaultContext() *context {
	environment := &config.ConfigurationEnvironment{}
	return newEnvironmentContext(environment)
}

func (c *context) banner() {
	if bannerFile, ok := c.environment.GetString("banner.file"); ok {
		bannerBts, err := os.ReadFile(bannerFile)
		if err != nil {
			c.logger.INFO(string(bannerBts))
			return
		}
	}
	c.logger.INFO(fmt.Sprintf(defaultBanner, VERSION))
}

func (c *context) DEBUG(format string, arr ...interface{}) {
	c.logger.DEBUG(format, arr...)
}

func (c *context) INFO(format string, arr ...interface{}) {
	c.logger.INFO(format, arr...)
}

func (c *context) WARN(format string, arr ...interface{}) {
	c.logger.WARN(format, arr...)
}

func (c *context) ERROR(format string, arr ...interface{}) {
	c.logger.ERROR(format, arr...)
}

func (c *context) Register(registers ...interfaces.InjectRegister) {
	c.registers = append(c.registers, registers...)
}

func (c *context) AutoFactory(auto ...interfaces.AutoFactory) {
	c.auto = append(c.auto, auto...)
}

func (c *context) Use(middleware ...interfaces.OrderedMiddleware) {
	c.middleware = append(c.middleware, middleware...)
}

func (c *context) Route(router ...interfaces.Router) {
	c.routers = append(c.routers, router...)
}

func (c *context) Shutdown(shutdown ...interfaces.ShutdownHook) {
	c.shutdownHooks = append(c.shutdownHooks, shutdown...)
}

func (c *context) Forward(ctx *gin.Context, path string) {
	url := ctx.Request.URL.Path
	ctx.Request.URL.Path = path
	ctx.Request.RequestURI = strings.Replace(ctx.Request.RequestURI, url, path, 1)
	c.server.HandleContext(ctx)
	ctx.Abort()
}

func (c *context) Get(key string) (interface{}, bool) {
	return c.store.Load(key)
}

func (c *context) Set(key string, value interface{}) {
	c.store.Store(key, value)
}

type resolver struct {
	env config.Config
}

func (r *resolver) Resolve(key string) (interface{}, bool) {
	return r.env.Get(key)
}

type buildinRegister struct {
	c *context
}

func (p *buildinRegister) Named() map[string]interface{} {
	return map[string]interface{}{
		"environment": p.c.environment,
		"logger":      p.c.logger,
		"server":      p.c.server,
	}
}

func (p *buildinRegister) Typed() map[reflect.Type]interface{} {
	return map[reflect.Type]interface{}{
		reflect.TypeOf((*config.TypedConfig)(nil)).Elem(): p.c.environment,
		reflect.TypeOf((*interfaces.Logger)(nil)).Elem():  p.c.logger,
		reflect.TypeOf((*gin.Engine)(nil)):                p.c.server,
	}
}

func (p *buildinRegister) Order() int {
	return BuildinRegisterOrder
}

func (c *context) register() {
	rs := interfaces.OrderSlice[interfaces.InjectRegister](c.registers)
	sort.Sort(rs)
	for _, register := range rs {
		s := make(map[interface{}]struct{})
		for k, v := range register.Named() {
			if _, ok := inject.GetNamed(k); !ok {
				if _, ok := s[v]; !ok && register.Order() != BuildinRegisterOrder {
					err := inject.Inject(nil, v)
					if err != nil {
						panic(err)
					}
				}
				err := inject.RegisterNamed(k, v)
				if err != nil {
					panic(err)
				}
				s[v] = struct{}{}
			} else if register.Order() != BuildinRegisterOrder {
				panic(fmt.Errorf("named object \"%s\" is already registered", k))
			}
		}
		for k, v := range register.Typed() {
			if _, ok := inject.GetTyped(k); !ok {
				if _, ok := s[v]; !ok && register.Order() != BuildinRegisterOrder {
					err := inject.Inject(nil, v)
					if err != nil {
						panic(err)
					}
				}
				err := inject.RegisterTyped(k, v)
				if err != nil {
					panic(err)
				}
				s[v] = struct{}{}
			} else if register.Order() != BuildinRegisterOrder {
				panic(fmt.Errorf("typed object %s is already registered", k))
			}
		}
	}
}

func (c *context) Run() {
	ctx.banner()

	if c.server == nil {
		mode := c.environment.GetStringOr("server.mode", "release")
		gin.SetMode(mode)
		c.server = gin.New()
		c.server.SetTrustedProxies(nil)
	}

	c.register()

	resolver := &resolver{c.environment}
	fs := interfaces.OrderSlice[interfaces.AutoFactory](c.auto)
	sort.Sort(fs)
	for _, a := range fs {
		err := inject.Inject(resolver, a)
		if err != nil {
			panic(err)
		}
		if a.Condition() {
			common.DEBUG("%s is starting", a.Name())
			err := a.OnStart()
			if err != nil {
				common.ERROR("", err)
				panic(err)
			}
			common.DEBUG("%s is start", a.Name())
			for k, v := range a.Named() {
				err := inject.RegisterNamed(k, v)
				if err != nil {
					panic(err)
				}
			}
			for k, v := range a.Typed() {
				err := inject.RegisterTyped(k, v)
				if err != nil {
					panic(err)
				}
			}
		} else {
			common.DEBUG("%s is disabled", a.Name())
		}
	}

	defer func() {
		for i := len(fs) - 1; i >= 0; i-- {
			if fs[i].Condition() {
				common.DEBUG("%s is stoping", fs[i].Name())
				err := fs[i].OnStop()
				if err != nil {
					common.ERROR("", err)
				}
				common.DEBUG("%s is stop", fs[i].Name())
			}
		}
		c.logger.INFO("Server is stop")
	}()

	ms := interfaces.OrderSlice[interfaces.OrderedMiddleware](c.middleware)
	sort.Sort(ms)
	for _, m := range ms {
		err := inject.Inject(resolver, m)
		if err != nil {
			panic(err)
		}
		if m.Condition() {
			c.server.Use(m.Function())
		} else {
			common.DEBUG("%s is disabled", reflect.TypeOf(m))
		}
	}
	for _, router := range c.routers {
		err := inject.Inject(resolver, router)
		if err != nil {
			panic(err)
		}
	}
	prefix := c.environment.GetStringOr("server.prefix", "")
	base := c.server.Group(prefix)
	for _, router := range c.routers {
		rs := router.Router()
		group := base.Group("")
		if mr, ok := router.(interfaces.MiddlewareRouter); ok {
			if ms := mr.Middleware(); ms != nil {
				group.Use(ms...)
			}
		}
		for name, function := range rs {
			tokens := strings.Split(name, " ")
			methods := strings.Split(tokens[0], ",")
			for _, method := range methods {
				group.Handle(strings.ToUpper(method), tokens[1], function)
			}
		}
	}
	ip := c.environment.GetStringOr("server.ip", "0.0.0.0")
	port := c.environment.GetStringOr("server.port", "8080")
	go func() {
		err := c.server.Run(fmt.Sprintf("%s:%s", ip, port))
		if err != nil {
			panic(err)
		}
	}()
	c.logger.INFO("Listening on: %s:%s", ip, port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	hs := interfaces.OrderSlice[interfaces.ShutdownHook](c.shutdownHooks)
	sort.Sort(hs)
	for i := len(hs) - 1; i >= 0; i-- {
		hs[i].Function()()
		common.DEBUG("%s is stop", hs[i].Name())
	}
	c.logger.INFO("Server stoping...")
}
