package siu_test

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-zookeeper/zk"
	"github.com/stella-go/siu"
	"github.com/stella-go/siu/config"
)

type S struct {
	Conn *zk.Conn `@siu:""`
}

func (p *S) Init() {
	fmt.Printf("siu test\n")
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

func TestRun(t *testing.T) {
	go func() {
		listener, _ := net.Listen("tcp", "127.0.0.1:2181")
		listener.Accept()
	}()
	os.Setenv("STELLA_LOGGER_LEVEL", "debug")
	os.Setenv("STELLA_ZOOKEEPER", "zookeeperxxx")
	os.Setenv("STELLA_ZOOKEEPER_SERVERS", "127.x0x.0.1:x21x81")
	os.Setenv("STELLA_MIDDLEWARE_CROS_DISABLE", "true")
	siu.Register(&R{})
	siu.Use(&S{})
	siu.Run()
}
