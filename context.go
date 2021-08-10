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
	"io"
	"io/ioutil"
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

type context struct {
	rootLogger *logger.Logger
	auto       []autoconfig.AutoConfig
	middleware []middleware.OrderedMiddleware
	ctx        map[string]interface{}

	routers []Router
}

var ctx *context = &context{
	auto:       make([]autoconfig.AutoConfig, 0),
	middleware: make([]middleware.OrderedMiddleware, 0),
	ctx:        make(map[string]interface{}),
	routers:    make([]Router, 0),
}

func init() {
	initLogger()
	banner()
	setDefault()
}

func initLogger() {
	_, ok := ctx.EnvGet(LoggerEnvKey)
	if !ok {
		ctx.rootLogger = logger.GetLogger("SIU")
		return
	}
	slevel := ctx.EnvGetStringOr(LoggerLevelEnvKey, "info")
	level := logger.Parse(slevel)

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
	rotateWriter, _ := logger.NewConfigRotateWriter(cfg)
	writer := io.MultiWriter(os.Stdout, rotateWriter)
	rootLogger := logger.NewRootLogger(level, &logger.DefaultFormatter{}, writer)
	ctx.rootLogger = rootLogger.GetLogger("SIU")
}

func banner() {
	if bannerFile, ok := ctx.EnvGetString("banner.file"); ok {
		bannerBts, err := ioutil.ReadFile(bannerFile)
		if err != nil {
			ctx.RootLogger().INFO(string(bannerBts))
			return
		}
	}
	ctx.RootLogger().INFO(fmt.Sprintf(defaultBanner, VERSION))
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

func (c *context) Run() {
	s := autoconfig.AutoConfigSlice(ctx.auto)
	sort.Sort(s)

	defer func() {
		for i := len(s) - 1; i >= 0; i-- {
			if s[i].Condition() {
				ctx.RootLogger().INFO("%s is stoping", s[i].Name())
				err := s[i].OnStop()
				if err != nil {
					ctx.RootLogger().ERROR("", err)
				}
			}
		}
	}()

	for _, a := range s {
		if a.Condition() {
			ctx.RootLogger().INFO("%s is starting", a.Name())
			err := a.OnStart()
			if err != nil {
				ctx.RootLogger().ERROR("", err)
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
	server := gin.New()
	m := middleware.OrderedMiddlewareSlice(ctx.middleware)
	sort.Sort(m)
	for _, middleware := range m {
		if middleware.Condition() {
			server.Use(middleware.Function())
		}
	}
	for _, route := range ctx.routers {
		rs := route.Router()
		for name, function := range rs {
			tokens := strings.Split(name, " ")
			server.Handle(strings.ToUpper(tokens[0]), tokens[1], function)
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
	ctx.RootLogger().INFO("Listening on: %s:%s", ip, port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx.RootLogger().INFO("Server stoping...")
	return nil
}
