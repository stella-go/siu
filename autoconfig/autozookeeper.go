// Copyright 2010-2024 the original author or authors.

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
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/stella-go/siu/config"
)

const (
	ZookeeperKey               = "zookeeper"
	ZookeeperDisableKey        = ZookeeperKey + ".disable"
	ZookeeperServersKey        = ZookeeperKey + ".servers"
	ZookeepersessionTimeoutKey = ZookeeperKey + ".sessionTimeout"
	ZookeeperOrder             = 10
)

type AutoZookeeper struct {
	Conf config.TypedConfig `@siu:"name='environment',default='type'"`
	conn *zk.Conn
}

func (p *AutoZookeeper) Condition() bool {
	_, ok1 := p.Conf.Get(ZookeeperKey)
	v, ok2 := p.Conf.GetBool(ZookeeperDisableKey)

	if ok2 && v {
		return false
	}
	return ok1
}

func (p *AutoZookeeper) OnStart() error {
	conn, err := createZookeeper(p.Conf, ZookeeperKey)
	if err != nil {
		return err
	}
	if conn != nil {
		p.conn = conn
	}
	return nil
}

func (p *AutoZookeeper) OnStop() error {
	if p.conn != nil {
		p.conn.Close()
	}
	return nil
}

func (*AutoZookeeper) Order() int {
	return ZookeeperOrder
}

func (*AutoZookeeper) Name() string {
	return ZookeeperKey
}

func (p *AutoZookeeper) Named() map[string]interface{} {
	return map[string]interface{}{
		ZookeeperKey: p.conn,
	}
}

func (p *AutoZookeeper) Typed() map[reflect.Type]interface{} {
	refType := reflect.TypeOf((*zk.Conn)(nil))
	return map[reflect.Type]interface{}{
		refType: p.conn,
	}
}

func createZookeeper(Conf config.TypedConfig, prefix string) (*zk.Conn, error) {
	serversStr, ok := Conf.GetString(ZookeeperServersKey)
	if !ok {
		return nil, fmt.Errorf("zookeeper servers can not be empty")
	}
	servers := strings.Split(serversStr, ",")
	timeout := Conf.GetIntOr(ZookeepersessionTimeoutKey, 60000)

	conn, event, err := zk.Connect(servers, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		return nil, err
	}
outer:
	for e := range event {
		switch e.State {
		case zk.StateConnecting:
			continue outer
		case zk.StateConnected:
			break outer
		}
	}
	return conn, nil
}
