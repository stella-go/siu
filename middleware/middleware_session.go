// Copyright 2010-2023 the original author or authors.

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
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/interfaces"
)

type expiration struct {
	value string
	exp   int64
}

const (
	SessionDisableKey  = "middleware.session.disable"
	SessionMiddleOrder = 50
	SessionCookieKey   = "siuid"
	SessionContextKey  = "session"
)

type MiddlewareSession struct {
	Conf   config.TypedConfig `@siu:"name='environment',default='type'"`
	Logger interfaces.Logger  `@siu:"name='logger',default='type'"`
	Redis  redis.Cmdable      `@siu:"name='redis',default='zero'"`

	store *sync.Map
}

func (p *MiddlewareSession) Init() {
	if p.Redis == nil {
		p.store = &sync.Map{}
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			for range ticker.C {
				keys := make([]string, 0)
				p.store.Range(func(key, value any) bool {
					v, _ := value.(*expiration)
					if v.exp < time.Now().Unix() {
						keys = append(keys, key.(string))
					}
					return true
				})
				for _, key := range keys {
					p.store.Delete(key)
				}
			}
		}()
	}
}

func (p *MiddlewareSession) Condition() bool {
	if v, ok := p.Conf.GetBool(SessionDisableKey); ok && v {
		return false
	}
	return true
}

func (p *MiddlewareSession) Function() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid, _ := c.Cookie(SessionCookieKey)
		if sid == "" {
			sid = uuid.NewString()
		}
		if session, ok := p.Get(sid); ok {
			c.Set(SessionContextKey, session)
		}
		c.SetCookie(SessionCookieKey, sid, 86400, "", "", false, true)

		c.Next()

		if session := c.GetString(SessionContextKey); session != "" {
			p.Set(sid, session)
		}
	}
}

func (p *MiddlewareSession) Order() int {
	return SessionMiddleOrder
}

func (p *MiddlewareSession) Get(key string) (string, bool) {
	if p.Redis != nil {
		cmd := p.Redis.Get(context.Background(), key)
		value := cmd.Val()
		if value != "" {
			return value, true
		} else {
			return "", false
		}
	} else {
		if value, ok := p.store.Load(key); ok {
			value, _ := value.(*expiration)
			if value.exp >= time.Now().Unix() {
				return value.value, true
			} else {
				return "", false
			}
		} else {
			return "", false
		}
	}
}

func (p *MiddlewareSession) Set(key string, value string) {
	if p.Redis != nil {
		p.Redis.Set(context.Background(), "session#"+key, value, 86400*time.Second)
	} else {
		p.store.Store(key, &expiration{
			exp:   time.Now().Unix() + int64(86400),
			value: value,
		})
	}
}
