// Copyright 2010-2022 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package siu_test

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-zookeeper/zk"
	"github.com/stella-go/siu"
	"github.com/stella-go/siu/config"
)

type S struct {
	Conn *zk.Conn `@siu:""`
}

func (p *S) Init() {
	fmt.Printf("siu test init\n")
}

func (p *S) Condition() bool {
	return true
}

func (p *S) Function() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}
func (p *S) Order() int {
	return 0
}

type C struct{}

func (*C) Decrypt(enc string) (string, error) {
	return strings.ReplaceAll(enc, "x", ""), nil
}

type R struct{}

func (*R) Named() map[string]interface{} {
	return map[string]interface{}{
		"environment": &config.EnciphermentEnvironment{Cipher: &C{}},
	}
}
func (*R) Typed() map[reflect.Type]interface{} {
	return nil
}

type Router struct{}

func (*Router) Router() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GET /hi": func(ctx *gin.Context) {
			time.Sleep(5 * time.Millisecond)
			ctx.String(200, "hello")
		},
	}
}

func TestRun(t *testing.T) {
	go func() {
		listener, _ := net.Listen("tcp", "127.0.0.1:2181")
		listener.Accept()
	}()
	go func() {
		time.Sleep(1 * time.Second)
		http.Get("http://localhost:8080/hi")
		http.Get("http://localhost:8080/abc")
		syscall.Kill(os.Getpid(), 15)
	}()
	os.Setenv("STELLA_LOGGER_SIU", "false")
	os.Setenv("STELLA_LOGGER_LEVEL", "debug")
	os.Setenv("STELLA_ZOOKEEPER", "zookeeperxxx")
	os.Setenv("STELLA_ZOOKEEPER_SERVERS", "127.x0x.0.1:x21x81")
	os.Setenv("STELLA_MIDDLEWARE_CROS_DISABLE", "true")
	siu.Register(&R{})
	siu.Use(&S{})
	siu.Route(&Router{})
	siu.Run()
}
