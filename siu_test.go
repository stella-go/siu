package siu_test

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-zookeeper/zk"
	"github.com/stella-go/siu"
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

func TestRun(t *testing.T) {
	go func() {
		listener, _ := net.Listen("tcp", "127.0.0.1:2181")
		listener.Accept()
	}()
	os.Setenv("STELLA_LOGGER_LEVEL", "debug")
	os.Setenv("STELLA_ZOOKEEPER", "zookeeper")
	os.Setenv("STELLA_ZOOKEEPER_SERVERS", "127.0.0.1:2181")
	siu.Use(&S{})
	siu.Run()
}
