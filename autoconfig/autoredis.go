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

package autoconfig

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stella-go/siu/config"
)

const (
	RedisKey             = "redis"
	RedisDatasourceOrder = 30
)

type AutoRedis struct {
	client        *redis.Client
	clusterClient *redis.ClusterClient
}

func (*AutoRedis) Condition() bool {
	_, ok := config.Get(RedisKey)
	return ok
}

func (p *AutoRedis) OnStart() error {
	addrStr, ok := config.GetString(RedisKey + ".addr")
	if !ok {
		return nil
	}
	addrs := strings.Split(addrStr, ",")
	if len(addrs) > 1 {
		clusterClient, err := createClusterRedis(RedisKey)
		if err != nil {
			return err
		}
		if clusterClient != nil {
			p.clusterClient = clusterClient
		}
	} else {
		client, err := createRedis(RedisKey)
		if err != nil {
			return err
		}
		if client != nil {
			p.client = client
		}
	}
	return nil
}

func createRedis(prefix string) (*redis.Client, error) {
	addr, ok := config.GetString(prefix + ".addr")
	if !ok {
		return nil, fmt.Errorf("reids address can not be empty")
	}
	password := config.GetStringOr(prefix+".password", "")
	db := config.GetIntOr(prefix+".db", 0)
	poolSize := config.GetIntOr(prefix+".poolSize", 4)
	minIdle := config.GetIntOr(prefix+".maxIdle", 1)
	dialTimeout := config.GetIntOr(prefix+".dialTimeout", 5000)
	readTimeout := config.GetIntOr(prefix+".readTimeout", 5000)
	writeTimeout := config.GetIntOr(prefix+".writeTimeout", 5000)
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

func createClusterRedis(prefix string) (*redis.ClusterClient, error) {
	addrStr, ok := config.GetString(prefix + ".addr")
	if !ok {
		return nil, fmt.Errorf("reids address can not be empty")
	}
	addrs := strings.Split(addrStr, ",")
	password := config.GetStringOr(prefix+".password", "")
	poolSize := config.GetIntOr(prefix+".poolSize", 4)
	minIdle := config.GetIntOr(prefix+".maxIdle", 1)
	dialTimeout := config.GetIntOr(prefix+".dialTimeout", 5000)
	readTimeout := config.GetIntOr(prefix+".readTimeout", 5000)
	writeTimeout := config.GetIntOr(prefix+".writeTimeout", 5000)

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

func (p *AutoRedis) Get() interface{} {
	if p.client != nil {
		return p.client
	}
	if p.clusterClient != nil {
		return p.clusterClient
	}
	return nil
}
