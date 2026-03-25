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
	"github.com/robfig/cron/v3"
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
	loggerEnvKey            = "logger"
	loggerUseEnvKey         = loggerEnvKey + ".siu"
	loggerTagEnvKey         = loggerEnvKey + ".tag"
	loggerLevelEnvKey       = loggerEnvKey + ".level"
	loggerPatternEnvKey     = loggerEnvKey + ".pattern"
	loggerDailyEnvKey       = loggerEnvKey + ".daily"
	loggerPathEnvKey        = loggerEnvKey + ".path"
	loggerFileEnvKey        = loggerEnvKey + ".file"
	loggerMaxFilesEnvKey    = loggerEnvKey + ".maxFiles"
	loggerMaxFileSizeEnvKey = loggerEnvKey + ".maxFileSize"
	loggerSyslogEnvKey      = loggerEnvKey + ".syslog"

	BuildinRegisterOrder = 0
	BeanRegisterOrder
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

func (p *buildinLogger) logf(lv logger.Level, levelStr string, format string, arr ...any) {
	if p.logLevel <= lv {
		if len(arr) > 0 {
			if _, ok := arr[len(arr)-1].(error); ok {
				format += " %v"
			}
		}
		msg := fmt.Sprintf(format, arr...)
		p.l.Printf("%s - %s %s", levelStr, p.tag, msg)
	}
}

func (p *buildinLogger) DEBUG(format string, arr ...any) {
	p.logf(logger.DebugLevel, "DEBUG", format, arr...)
}

func (p *buildinLogger) INFO(format string, arr ...any) {
	p.logf(logger.InfoLevel, "INFO ", format, arr...)
}

func (p *buildinLogger) WARN(format string, arr ...any) {
	p.logf(logger.WarnLevel, "WARN ", format, arr...)
}

func (p *buildinLogger) ERROR(format string, arr ...any) {
	p.logf(logger.ErrorLevel, "ERROR", format, arr...)
}

func (p *buildinLogger) Level() logger.Level {
	return p.logLevel
}

func (p *buildinLogger) Tag() string {
	return p.tag
}

type cronLogger struct {
	logger interfaces.Logger
}

func (p *cronLogger) Printf(format string, v ...any) {
	p.logger.ERROR(format, v...)
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

	cron *cron.Cron
}

func newContext(environment config.TypedConfig, contextLogger interfaces.Logger, server *gin.Engine) *context {
	ctx := &context{
		environment:   environment,
		logger:        contextLogger,
		registers:     make([]interfaces.InjectRegister, 0),
		auto:          make([]interfaces.AutoFactory, 0),
		middleware:    make([]interfaces.OrderedMiddleware, 0),
		routers:       make([]interfaces.Router, 0),
		shutdownHooks: make([]interfaces.ShutdownHook, 0),
		store:         &sync.Map{},
		server:        server,
		cron: cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
		)), cron.WithChain(cron.Recover(cron.PrintfLogger(&cronLogger{logger: contextLogger})))),
	}
	if leveledLogger, ok := contextLogger.(interfaces.LeveledLogger); ok {
		common.SetLevel(leveledLogger.Level())
	}
	if taggedLogger, ok := contextLogger.(interfaces.TaggedLogger); ok {
		common.SetTag(taggedLogger.Tag())
	}
	ctx.Register(&buildinRegister{ctx})
	ctx.AutoFactory(&autoconfig.AutoMysql{}, &autoconfig.AutoGorm{}, &autoconfig.AutoRedis{}, &autoconfig.AutoZookeeper{}, &autoconfig.AutoOss{}, &autoconfig.AutoCipher{})
	ctx.Use(&middleware.MiddlewareRewrite{}, &middleware.MiddlewareAccess{}, &middleware.MiddlewareCORS{}, &middleware.MiddlewareErrorlog{}, &middleware.MiddlewareResource{}, &middleware.MiddlewareSession{}, &middleware.MiddlewareJwt{})
	return ctx
}

func newEnvironmentContext(environment config.TypedConfig) *context {
	logUse := environment.GetBoolOr(loggerUseEnvKey, true)
	tag := environment.GetStringOr(loggerTagEnvKey, "[SIU]")
	logLevel := logger.Parse(environment.GetStringOr(loggerLevelEnvKey, "info"))
	logPattern := environment.GetStringOr(loggerPatternEnvKey, "%d{06-01-02.15:04:05.000} [%g] %p %c - %m")
	daily := environment.GetBoolOr(loggerDailyEnvKey, true)
	filePath := environment.GetStringOr(loggerPathEnvKey, ".")
	fileName := environment.GetStringOr(loggerFileEnvKey, "stdout")
	maxFiles := environment.GetIntOr(loggerMaxFilesEnvKey, 30)
	maxFileSize := environment.GetIntOr(loggerMaxFileSizeEnvKey, 200)

	var w io.Writer
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
	w = writer
	if syslog, ok := environment.GetString(loggerSyslogEnvKey); ok {
		hostname, err := os.Hostname()
		if err != nil || hostname == "" {
			hostname = "unknown"
		}
		protocol := ""
		if strings.HasPrefix(syslog, "udp:") {
			protocol = "udp"
			syslog = strings.TrimPrefix(syslog, "udp:")
		} else if strings.HasPrefix(syslog, "tcp:") {
			protocol = "tcp"
			syslog = strings.TrimPrefix(syslog, "tcp:")
		} else {
			protocol = "udp"
		}
		sysWriter, err := logger.NewConfigSyslogWriter(&logger.SyslogConfig{
			Facility: logger.LOG_F_LOCAL5,
			Severity: logger.LOG_S_INFO,
			Hostname: hostname,
			Addr:     syslog,
			Protocol: protocol,
			Tag:      tag,
		})

		if err != nil {
			panic(err)
		}
		w = io.MultiWriter(writer, sysWriter)
	}

	var contextLogger interfaces.Logger
	if logUse {
		contextLogger = logger.NewRootLogger(logLevel, &logger.PatternFormatter{Pattern: logPattern}, w).GetLogger(tag)
	} else {
		contextLogger = newBuildinLogger(logLevel, tag, w)
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
		if err == nil {
			c.logger.INFO(string(bannerBts))
			return
		}
	}
	c.logger.INFO(fmt.Sprintf(defaultBanner, VERSION))
}

func (c *context) DEBUG(format string, arr ...any) {
	c.logger.DEBUG(format, arr...)
}

func (c *context) INFO(format string, arr ...any) {
	c.logger.INFO(format, arr...)
}

func (c *context) WARN(format string, arr ...any) {
	c.logger.WARN(format, arr...)
}

func (c *context) ERROR(format string, arr ...any) {
	c.logger.ERROR(format, arr...)
}

type beanRegister struct {
	obj  any
	name string
	typ  reflect.Type
}

func (p *beanRegister) Named() map[string]any {
	return map[string]any{
		p.name: p.obj,
	}
}

func (p *beanRegister) Typed() map[reflect.Type]any {
	return map[reflect.Type]any{
		p.typ: p.obj,
	}
}

func (p *beanRegister) Order() int {
	return BeanRegisterOrder
}

func (c *context) RegisterBean(name string, typ reflect.Type, obj any) {
	c.registers = append(c.registers, &beanRegister{obj, name, typ})
}

func (c *context) GetBeanByName(name string) (any, bool) {
	return inject.GetNamed(name)
}

func (c *context) GetBeanByType(typ reflect.Type) (any, bool) {
	return inject.GetTyped(typ)
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

func (c *context) Get(key string) (any, bool) {
	return c.store.Load(key)
}

func (c *context) Set(key string, value any) {
	c.store.Store(key, value)
}

func (c *context) Cron(spec string, cmd func()) error {
	_, err := c.cron.AddFunc(spec, cmd)
	return err
}

type buildinRegister struct {
	c *context
}

func (p *buildinRegister) Named() map[string]any {
	return map[string]any{
		"environment": p.c.environment,
		"logger":      p.c.logger,
		"server":      p.c.server,
	}
}

func (p *buildinRegister) Typed() map[reflect.Type]any {
	return map[reflect.Type]any{
		reflect.TypeOf((*config.TypedConfig)(nil)).Elem(): p.c.environment,
		reflect.TypeOf((*interfaces.Logger)(nil)).Elem():  p.c.logger,
		reflect.TypeOf((*gin.Engine)(nil)):                p.c.server,
	}
}

func (p *buildinRegister) Order() int {
	return BuildinRegisterOrder
}

func (c *context) register(resolver inject.ValueResolver) {
	rs := interfaces.OrderSlice[interfaces.InjectRegister](c.registers)
	sort.Sort(rs)
	for _, register := range rs {
		s := make(map[any]struct{})
		for k, v := range register.Named() {
			if _, ok := inject.GetNamed(k); !ok {
				if _, ok := s[v]; !ok && register.Order() != BuildinRegisterOrder {
					err := inject.Inject(resolver, v)
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
					err := inject.Inject(resolver, v)
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
	c.banner()

	if c.server == nil {
		mode := c.environment.GetStringOr("server.mode", "release")
		gin.SetMode(mode)
		c.server = gin.New()
		c.server.SetTrustedProxies(nil)
	}
	c.cron.Start()
	defer c.cron.Stop()

	resolver := &inject.ConfigResolver{C: c.environment}

	c.register(resolver)

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
	inject.InvokeReady()

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
