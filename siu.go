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

	"github.com/go-redis/redis/v8"
	"github.com/go-zookeeper/zk"
	"github.com/stella-go/logger"
	"github.com/stella-go/siu/autoconfig"
	"github.com/stella-go/siu/middleware"
)

// export method

func RootLogger() *logger.Logger {
	return ctx.RootLogger()
}

func NewLogger(name string) *logger.Logger {
	return ctx.NewLogger(name)
}

func AutoConfig(auto ...autoconfig.AutoConfig) {
	ctx.AutoConfig(auto...)
}

func Use(middleware ...middleware.OrderedMiddleware) {
	ctx.Use(middleware...)
}

func Rotate(rotater ...Rotater) {
	ctx.Rotate(rotater...)
}

func Get(key string) (interface{}, bool) {
	return ctx.Get(key)
}

func DataSource() (*sql.DB, bool) {
	return ctx.DataSource()
}

func DataSourceWithName(name string) (*sql.DB, bool) {
	return ctx.DataSourceWithName(name)
}

func Redis() (*redis.Client, bool) {
	return ctx.Redis()
}

func RedisCluster() (*redis.ClusterClient, bool) {
	return ctx.RedisCluster()
}

func Zookeeper() (*zk.Conn, bool) {
	return ctx.Zookeeper()
}

func EnvGet(key string) (interface{}, bool) {
	return ctx.EnvGet(key)
}

func EnvGetInt(key string) (int, bool) {
	return ctx.EnvGetInt(key)
}

func EnvGetString(key string) (string, bool) {
	return ctx.EnvGetString(key)
}

func EnvGetBool(key string) (bool, bool) {
	return ctx.EnvGetBool(key)
}

func EnvGetOr(key string, defaultValue interface{}) interface{} {
	return ctx.EnvGetOr(key, defaultValue)
}

func EnvGetIntOr(key string, defaultValue int) int {
	return ctx.EnvGetIntOr(key, defaultValue)
}

func EnvGetBoolOr(key string, defaultValue bool) bool {
	return ctx.EnvGetBoolOr(key, defaultValue)
}

func EnvGetStringOr(key string, defaultValue string) string {
	return ctx.EnvGetStringOr(key, defaultValue)
}

func Run() {
	ctx.Run()
}
