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
	"fmt"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/stella-go/siu/config"
)

const (
	ZookeeperKey   = "zookeeper"
	ZookeeperOrder = 10
)

type AutoZookeeper struct {
	conn *zk.Conn
}

func (*AutoZookeeper) Condition() bool {
	_, ok := config.Get(ZookeeperKey)
	return ok
}

func (p *AutoZookeeper) OnStart() error {
	conn, err := createZookeeper(ZookeeperKey)
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

func (p *AutoZookeeper) Get() interface{} {
	return p.conn
}

func createZookeeper(prefix string) (*zk.Conn, error) {
	serversStr, ok := config.GetString(prefix + ".servers")
	if !ok {
		return nil, fmt.Errorf("zookeeper servers can not be empty")
	}
	servers := strings.Split(serversStr, ",")
	timeout := config.GetIntOr(prefix+".sessionTimeoutKey", 60000)

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
