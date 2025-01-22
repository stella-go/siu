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

package autoconfig

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stella-go/siu/config"
)

const (
	RedisKey             = "redis"
	RedisDisableKey      = RedisKey + ".disable"
	RedisAddrKey         = RedisKey + ".addr"
	RedisPasswordKey     = RedisKey + ".password"
	RedisDBKey           = RedisKey + ".db"
	RedisPoolSizeKey     = RedisKey + ".poolSize"
	RedisMaxIdleKey      = RedisKey + ".maxIdle"
	RedisDialTimeoutKey  = RedisKey + ".dialTimeout"
	RedisReadTimeoutKey  = RedisKey + ".readTimeout"
	RedisWriteTimeoutKey = RedisKey + ".writeTimeout"
	RedisDatasourceOrder = 30
)

type AutoRedis struct {
	Conf          config.TypedConfig `@siu:"name='environment',default='type'"`
	client        *redis.Client
	clusterClient *redis.ClusterClient
}

func (p *AutoRedis) Condition() bool {
	_, ok1 := p.Conf.Get(RedisKey)
	v, ok2 := p.Conf.GetBool(RedisDisableKey)

	if ok2 && v {
		return false
	}
	return ok1
}

func (p *AutoRedis) OnStart() error {
	addrStr, ok := p.Conf.GetString(RedisKey + ".addr")
	if !ok {
		return nil
	}
	addrs := strings.Split(addrStr, ",")
	if len(addrs) > 1 {
		clusterClient, err := createClusterRedis(p.Conf, RedisKey)
		if err != nil {
			return err
		}
		if clusterClient != nil {
			p.clusterClient = clusterClient
		}
	} else {
		client, err := createRedis(p.Conf, RedisKey)
		if err != nil {
			return err
		}
		if client != nil {
			p.client = client
		}
	}
	return nil
}

func (p *AutoRedis) OnStop() error {
	if p.client != nil {
		p.client.Close()
	}
	if p.clusterClient != nil {
		p.clusterClient.Close()
	}
	return nil
}

func (*AutoRedis) Order() int {
	return RedisDatasourceOrder
}

func (*AutoRedis) Name() string {
	return RedisKey
}

func (p *AutoRedis) Named() map[string]interface{} {
	if p.client != nil {
		return map[string]interface{}{
			RedisKey: p.client,
		}
	}
	if p.clusterClient != nil {
		return map[string]interface{}{
			RedisKey: p.clusterClient,
		}
	}
	return nil
}

func (p *AutoRedis) Typed() map[reflect.Type]interface{} {
	refType := reflect.TypeOf((*redis.Cmdable)(nil)).Elem()
	if p.client != nil {
		return map[reflect.Type]interface{}{
			refType: p.client,
		}
	}
	if p.clusterClient != nil {
		return map[reflect.Type]interface{}{
			refType: p.clusterClient,
		}
	}
	return nil
}

func createRedis(conf config.TypedConfig, _ /*prefix*/ string) (*redis.Client, error) {
	addr, ok := conf.GetString(RedisAddrKey)
	if !ok {
		return nil, fmt.Errorf("reids address can not be empty")
	}
	password := conf.GetStringOr(RedisPasswordKey, "")
	db := conf.GetIntOr(RedisDBKey, 0)
	poolSize := conf.GetIntOr(RedisPoolSizeKey, 4)
	minIdle := conf.GetIntOr(RedisMaxIdleKey, 1)
	dialTimeout := conf.GetIntOr(RedisDialTimeoutKey, 5000)
	readTimeout := conf.GetIntOr(RedisReadTimeoutKey, 5000)
	writeTimeout := conf.GetIntOr(RedisWriteTimeoutKey, 5000)
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdle,
		DialTimeout:  time.Duration(dialTimeout) * time.Millisecond,
		ReadTimeout:  time.Duration(readTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(writeTimeout) * time.Millisecond,
	})
	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	} else {
		return client, nil
	}
}

func createClusterRedis(conf config.TypedConfig, _ /*prefix*/ string) (*redis.ClusterClient, error) {
	addrStr, ok := conf.GetString(RedisAddrKey)
	if !ok {
		return nil, fmt.Errorf("reids address can not be empty")
	}
	addrs := strings.Split(addrStr, ",")
	password := conf.GetStringOr(RedisPasswordKey, "")
	poolSize := conf.GetIntOr(RedisPoolSizeKey, 4)
	minIdle := conf.GetIntOr(RedisMaxIdleKey, 1)
	dialTimeout := conf.GetIntOr(RedisDialTimeoutKey, 5000)
	readTimeout := conf.GetIntOr(RedisReadTimeoutKey, 5000)
	writeTimeout := conf.GetIntOr(RedisWriteTimeoutKey, 5000)

	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		PoolSize:     poolSize,
		MinIdleConns: minIdle,
		DialTimeout:  time.Duration(dialTimeout) * time.Millisecond,
		ReadTimeout:  time.Duration(readTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(writeTimeout) * time.Millisecond,
	})
	ctx := context.Background()
	if _, err := clusterClient.Ping(ctx).Result(); err != nil {
		return nil, err
	} else {
		return clusterClient, nil
	}
}
