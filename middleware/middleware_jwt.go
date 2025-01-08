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

package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/t"
)

const (
	JwtKey              = "middleware.jwt"
	JwtDisableKey       = "middleware.jwt.disable"
	JwtCookiedomainKey  = "middleware.jwt.cookie-domain"
	JwtExpiresecondsKey = "middleware.jwt.expire-seconds"
	JwtSecretKey        = "middleware.jwt.secret"
	JwtExcludesKey      = "middleware.jwt.excludes"

	JwtMiddleOrder = 50
)

const (
	JwtCookieKey         = "Authorization"
	JwtTokenContextKey   = "jwt"
	JwtSubjectContextKey = "subject"
)

type Subject struct {
	Id     int64                  `json:"id"`
	Name   string                 `json:"name"`
	Roles  []string               `json:"roles"`
	Others map[string]interface{} `json:"others"`
}

type MiddlewareJwt struct {
	Conf          config.TypedConfig `@siu:"name='environment',default='type'"`
	cookieDomain  string
	expireSeconds int
	secret        string
	excludes      []string
}

func (p *MiddlewareJwt) Init() {
	p.cookieDomain = p.Conf.GetStringOr(JwtCookiedomainKey, "")
	p.expireSeconds = p.Conf.GetIntOr(JwtExpiresecondsKey, 3600)
	p.secret = p.Conf.GetStringOr(JwtSecretKey, uuid.NewString())
	excludes := p.Conf.GetOr(JwtExcludesKey, []string{"/login", "/admin/login", "/api/login"})
	if e, ok := excludes.([]string); ok {
		p.excludes = e
	} else {
		if e, ok := excludes.(string); ok {
			splits := strings.Split(e, ",")
			for _, s := range splits {
				p.excludes = append(p.excludes, strings.TrimSpace(s))
			}
		}
	}
}

func (p *MiddlewareJwt) Condition() bool {
	_, ok1 := p.Conf.Get(JwtKey)
	v, ok2 := p.Conf.GetBool(JwtDisableKey)

	if ok2 && v {
		return false
	}
	return ok1
}

func (p *MiddlewareJwt) Function() gin.HandlerFunc {
	return func(c *gin.Context) {
		if p.isIncludes(c) {
			token, _ := c.Cookie(JwtCookieKey)
			if token == "" {
				token = c.GetHeader(JwtCookieKey)
			}
			if token == "" {
				c.JSON(200, t.FailWith(401, "Unauthorized"))
				c.Abort()
				return
			}
			subject, err := JwtVerify(token, p.secret)
			if err != nil {
				c.JSON(200, t.FailWith(401, "Unauthorized"))
				c.Abort()
				return
			}
			c.Set(JwtSubjectContextKey, subject)
		}
		c.Next()
	}
}

func (p *MiddlewareJwt) Order() int {
	return JwtMiddleOrder
}

func (p *MiddlewareJwt) isIncludes(c *gin.Context) bool {
	fullPath := c.FullPath()
	for _, p := range p.excludes {
		if p == fullPath {
			return false
		}
		if strings.HasSuffix(p, "/**") {
			if strings.HasPrefix(fullPath, p[:len(p)-3]) {
				return false
			}
		}
	}
	return true
}

func (p *MiddlewareJwt) SetCookie(c *gin.Context) {
	if token := c.GetString(JwtTokenContextKey); token != "" {
		c.SetCookie(JwtCookieKey, token, p.expireSeconds, "/", p.cookieDomain, false, true)
	} else {
		if value, ok := c.Get(JwtSubjectContextKey); ok {
			if subject, ok := value.(*Subject); ok {
				token, err := JwtSign(subject, p.secret, time.Duration(p.expireSeconds)*time.Second)
				if err != nil {
					c.JSON(200, t.FailWith(500, "system error"))
				}
				c.SetCookie(JwtCookieKey, token, p.expireSeconds, "/", p.cookieDomain, false, true)
			}
		}
	}
}
func (p *MiddlewareJwt) SetTokenCookie(c *gin.Context, token string) {
	c.SetCookie(JwtCookieKey, token, p.expireSeconds, "/", p.cookieDomain, false, true)
}
func (p *MiddlewareJwt) SetSubjectCookie(c *gin.Context, subject *Subject) error {
	token, err := JwtSign(subject, p.secret, time.Duration(p.expireSeconds)*time.Second)
	if err != nil {
		return err
	}
	c.SetCookie(JwtCookieKey, token, p.expireSeconds, "/", p.cookieDomain, false, true)
	return nil
}
func (p *MiddlewareJwt) ClearCookie(c *gin.Context) {
	c.SetCookie(JwtCookieKey, "", 0, "/", p.cookieDomain, false, true)
}

type Claims struct {
	*Subject
	jwt.RegisteredClaims
}

func JwtSign(subject *Subject, secret string, expire time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		Subject: subject,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
		},
	})
	return token.SignedString([]byte(secret))
}

func JwtVerify(tokenString string, secret string) (*Subject, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("the signature method is not supported: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok {
		return claims.Subject, nil
	} else {
		return nil, fmt.Errorf("the claims type is not supported")
	}
}

func PermissionsHaveRole(handler gin.HandlerFunc, needRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(needRoles) == 0 {
			handler(c)
			return
		} else {
			for _, r := range needRoles {
				if r == "any" {
					handler(c)
					return
				}
			}
		}
		if value, ok := c.Get(JwtSubjectContextKey); ok {
			if subject, ok := value.(*Subject); ok {
				roles := subject.Roles
				for _, r1 := range roles {
					for _, r2 := range needRoles {
						if r1 == r2 {
							handler(c)
							return
						}
					}
				}
			}
		}
		c.JSON(200, t.FailWith(403, "permission denied"))
	}
}
