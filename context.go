// Copyright 2010-2021 the original author or authors.

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
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-zookeeper/zk"
	"github.com/stella-go/logger"
	"github.com/stella-go/siu/autoconfig"
	"github.com/stella-go/siu/config"
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
	LoggerEnvKey             = "logger"
	LoggerUseEnvKey          = LoggerEnvKey + ".siu"
	LoggerLevelEnvKey        = LoggerEnvKey + ".level"
	LoggerDaliyEnvKey        = LoggerEnvKey + ".daliy"
	LoggerPathEnvKey         = LoggerEnvKey + ".path"
	LoggerFileEnvKey         = LoggerEnvKey + ".file"
	LoggerMaxFilesEnvKey     = LoggerEnvKey + ".maxFiles"
	LoggerMaxFileSizesEnvKey = LoggerEnvKey + ".maxFileSize"
)

type Router interface {
	Router() map[string]gin.HandlerFunc
}

type MiddlewareRouter interface {
	Router
	Middleware() []gin.HandlerFunc
}

type context struct {
	rootLogger *logger.Logger
	auto       []autoconfig.AutoConfig
	middleware []middleware.OrderedMiddleware
	ctx        map[string]interface{}
	routers    []Router

	server *gin.Engine
}

var ctx *context = &context{
	auto:       make([]autoconfig.AutoConfig, 0),
	middleware: make([]middleware.OrderedMiddleware, 0),
	ctx:        make(map[string]interface{}),
	routers:    make([]Router, 0),
	server:     gin.New(),
}

var logLevel = logger.Parse(ctx.EnvGetStringOr(LoggerLevelEnvKey, "info"))
var logUse = ctx.EnvGetBoolOr(LoggerUseEnvKey, true)

func init() {
	initLogger()
	banner()
	setDefault()
}

func initLogger() {
	daily := ctx.EnvGetBoolOr(LoggerDaliyEnvKey, true)
	filePath := ctx.EnvGetStringOr(LoggerPathEnvKey, ".")
	fileName := ctx.EnvGetStringOr(LoggerFileEnvKey, "stdout")
	maxFiles := ctx.EnvGetIntOr(LoggerMaxFilesEnvKey, 30)
	maxFileSize := ctx.EnvGetIntOr(LoggerMaxFileSizesEnvKey, 200)

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
	if logUse {
		rootLogger := logger.NewRootLogger(logLevel, &logger.DefaultFormatter{}, writer)
		ctx.rootLogger = rootLogger.GetLogger("SIU")
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.SetOutput(writer)
	}
}

func banner() {
	if bannerFile, ok := ctx.EnvGetString("banner.file"); ok {
		bannerBts, err := ioutil.ReadFile(bannerFile)
		if err != nil {
			ctx.INFO(string(bannerBts))
			return
		}
	}
	ctx.INFO(fmt.Sprintf(defaultBanner, VERSION))
}

func setDefault() {
	ctx.AutoConfig(&autoconfig.AutoMysql{}, &autoconfig.AutoRedis{}, &autoconfig.AutoZookeeper{})
	ctx.Use(&middleware.MiddlewareCROS{}, &middleware.MiddlewareResource{})
}

func (c *context) RootLogger() *logger.Logger {
	return c.rootLogger
}

func (c *context) NewLogger(name string) *logger.Logger {
	return c.rootLogger.GetLogger(name)
}

func (c *context) DEBUG(format string, arr ...interface{}) {
	if logUse {
		ctx.RootLogger().DEBUG(format, arr...)
	} else {
		if logLevel <= logger.DebugLevel {
			if len(arr) > 0 {
				if _, ok := arr[len(arr)-1].(error); ok {
					format += " %v"
				}
			}
			msg := fmt.Sprintf(format, arr...)
			log.Println("DEBUG - " + msg)
		}
	}
}

func (c *context) INFO(format string, arr ...interface{}) {
	if logUse {
		ctx.RootLogger().INFO(format, arr...)
	} else {
		if logLevel <= logger.InfoLevel {
			if len(arr) > 0 {
				if _, ok := arr[len(arr)-1].(error); ok {
					format += " %v"
				}
			}
			msg := fmt.Sprintf(format, arr...)
			log.Println("INFO  - " + msg)
		}
	}
}

func (c *context) WARN(format string, arr ...interface{}) {
	if logUse {
		ctx.RootLogger().WARN(format, arr...)
	} else {
		if logLevel <= logger.WarnLevel {
			if len(arr) > 0 {
				if _, ok := arr[len(arr)-1].(error); ok {
					format += " %v"
				}
			}
			msg := fmt.Sprintf(format, arr...)
			log.Println("WARN  - " + msg)
		}
	}
}

func (c *context) ERROR(format string, arr ...interface{}) {
	if logUse {
		ctx.RootLogger().ERROR(format, arr...)
	} else {
		if logLevel < logger.ErrorLevel {
			if len(arr) > 0 {
				if _, ok := arr[len(arr)-1].(error); ok {
					format += " %v"
				}
			}
			msg := fmt.Sprintf(format, arr...)
			log.Println("ERROR - " + msg)
		}
	}
}

func (c *context) AutoConfig(auto ...autoconfig.AutoConfig) {
	c.auto = append(c.auto, auto...)
}

func (c *context) Use(middleware ...middleware.OrderedMiddleware) {
	c.middleware = append(c.middleware, middleware...)
}

func (c *context) Route(router ...Router) {
	c.routers = append(c.routers, router...)
}

func (c *context) Set(key string, value interface{}) {
	c.ctx[key] = value
}

func (c *context) Get(key string) (interface{}, bool) {
	value, ok := c.ctx[key]
	return value, ok
}

func (c *context) DataSource() (*sql.DB, bool) {
	if v, ok := c.Get(autoconfig.MySQLDatasourceKey); ok {
		dbs := v.(map[string]*sql.DB)
		for _, db := range dbs {
			return db, true
		}
	}
	return nil, false
}

func (c *context) DataSourceWithName(name string) (*sql.DB, bool) {
	if v, ok := c.Get(autoconfig.MySQLDatasourceKey); ok {
		dbs := v.(map[string]*sql.DB)
		if db, ok := dbs[name]; ok {
			return db, true
		}
	}
	return nil, false
}

func (c *context) Redis() (*redis.Client, bool) {
	if v, ok := c.Get(autoconfig.RedisKey); ok {
		if c, ok := v.(*redis.Client); ok {
			return c, true
		}
	}
	return nil, false
}

func (c *context) RedisCluster() (*redis.ClusterClient, bool) {
	if v, ok := c.Get(autoconfig.RedisKey); ok {
		if c, ok := v.(*redis.ClusterClient); ok {
			return c, true
		}
	}
	return nil, false
}

func (c *context) Zookeeper() (*zk.Conn, bool) {
	if v, ok := c.Get(autoconfig.ZookeeperKey); ok {
		if c, ok := v.(*zk.Conn); ok {
			return c, true
		}
	}
	return nil, false
}

func (c *context) EnvGet(key string) (interface{}, bool) {
	return config.Get(key)
}

func (c *context) EnvGetInt(key string) (int, bool) {
	return config.GetInt(key)
}

func (c *context) EnvGetString(key string) (string, bool) {
	return config.GetString(key)
}

func (c *context) EnvGetBool(key string) (bool, bool) {
	return config.GetBool(key)
}

func (c *context) EnvGetOr(key string, defaultValue interface{}) interface{} {
	return config.GetOr(key, defaultValue)
}

func (c *context) EnvGetIntOr(key string, defaultValue int) int {
	return config.GetIntOr(key, defaultValue)
}

func (c *context) EnvGetBoolOr(key string, defaultValue bool) bool {
	return config.GetBoolOr(key, defaultValue)
}

func (c *context) EnvGetStringOr(key string, defaultValue string) string {
	return config.GetStringOr(key, defaultValue)
}

func (c *context) Server() *gin.Engine {
	return c.server
}

func (c *context) Run() {
	s := autoconfig.AutoConfigSlice(ctx.auto)
	sort.Sort(s)

	defer func() {
		for i := len(s) - 1; i >= 0; i-- {
			if s[i].Condition() {
				ctx.INFO("%s is stoping", s[i].Name())
				err := s[i].OnStop()
				if err != nil {
					ctx.ERROR("", err)
				}
			}
		}
	}()

	for _, a := range s {
		if a.Condition() {
			ctx.INFO("%s is starting", a.Name())
			err := a.OnStart()
			if err != nil {
				ctx.ERROR("", err)
				panic(err)
			} else {
				c.Set(a.Name(), a.Get())
			}
		}
	}

	err := xmain()
	if err != nil {
		panic(err)
	}
}

func xmain() error {
	mode := ctx.EnvGetStringOr("server.mode", "release")
	gin.SetMode(mode)
	server := ctx.server
	server.SetTrustedProxies(nil)
	m := middleware.OrderedMiddlewareSlice(ctx.middleware)
	sort.Sort(m)
	for _, middleware := range m {
		if middleware.Condition() {
			server.Use(middleware.Function())
		}
	}
	for _, router := range ctx.routers {
		rs := router.Router()
		group := server.Group("")
		if mr, ok := router.(MiddlewareRouter); ok {
			if ms := mr.Middleware(); ms != nil {
				group.Use(ms...)
			}
		}
		for name, function := range rs {
			tokens := strings.Split(name, " ")
			group.Handle(strings.ToUpper(tokens[0]), tokens[1], function)
		}
	}
	ip := ctx.EnvGetStringOr("server.ip", "0.0.0.0")
	port := ctx.EnvGetStringOr("server.port", "8080")
	go func() {
		err := server.Run(fmt.Sprintf("%s:%s", ip, port))
		if err != nil {
			panic(err)
		}
	}()
	ctx.INFO("Listening on: %s:%s", ip, port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx.INFO("Server stoping...")
	return nil
}
