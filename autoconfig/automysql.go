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

package autoconfig

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stella-go/siu/config"
)

const (
	MySQLDatasourceKey        = "mysql"
	MySQLDatasourceDisableKey = MySQLDatasourceKey + ".disable"
	MySQLDatasourceOrder      = 20
)

type AutoMysql struct {
	Conf config.TypedConfig `@siu:"name='environment',default='type'"`
	dbs  map[string]*sql.DB
}

func (p *AutoMysql) Condition() bool {
	_, ok1 := p.Conf.Get(MySQLDatasourceKey)
	v, ok2 := p.Conf.GetBool(MySQLDatasourceDisableKey)

	if ok2 && v {
		return false
	}
	return ok1
}

func (p *AutoMysql) OnStart() error {
	p.dbs = make(map[string]*sql.DB)

	datasources, _ := p.Conf.Get(MySQLDatasourceKey)

	db, err := createDB(p.Conf, MySQLDatasourceKey)
	if err != nil {
		return err
	}
	if db != nil {
		p.dbs[MySQLDatasourceKey] = db
	}

	datasourcesMap, ok := datasources.(map[string]interface{})
	if !ok {
		return nil
	}
	for datasource := range datasourcesMap {
		name := MySQLDatasourceKey + "." + datasource
		db, err := createDB(p.Conf, name)
		if err != nil {
			return err
		}
		if db != nil {
			p.dbs[name] = db
		}
	}
	return nil
}

func (p *AutoMysql) OnStop() error {
	for _, db := range p.dbs {
		if db != nil {
			db.Close()
		}
	}
	return nil
}

func (*AutoMysql) Order() int {
	return MySQLDatasourceOrder
}

func (*AutoMysql) Name() string {
	return MySQLDatasourceKey
}

func (p *AutoMysql) Named() map[string]interface{} {
	n := make(map[string]interface{})
	for k, v := range p.dbs {
		n[k] = v
	}
	return n
}

func (p *AutoMysql) Typed() map[reflect.Type]interface{} {
	if len(p.dbs) == 1 {
		for _, v := range p.dbs {
			refType := reflect.TypeOf((*sql.DB)(nil))
			return map[reflect.Type]interface{}{
				refType: v,
			}
		}
	}
	return nil
}

func createDB(conf config.TypedConfig, prefix string) (*sql.DB, error) {
	user, ok1 := conf.GetString(prefix + ".user")
	passwd, ok2 := conf.GetString(prefix + ".passwd")
	addr, ok3 := conf.GetString(prefix + ".addr")
	dbName, ok4 := conf.GetString(prefix + ".dbName")

	if !ok1 || !ok2 || !ok3 || !ok4 {
		if prefix == MySQLDatasourceKey {
			return nil, nil
		}
		return nil, fmt.Errorf("mysql datasource user/passwd/addr/dbName can not be empty")
	}

	collation := conf.GetStringOr(prefix+".collation", "utf8mb4_bin")
	timeout := conf.GetIntOr(prefix+".timeout", 60000)
	readTimeout := conf.GetIntOr(prefix+".readTimeout", 30000)
	writeTimeout := conf.GetIntOr(prefix+".writeTimeout", 30000)

	cparams, ok := conf.Get(prefix + ".params")
	params := make(map[string]string)
	if ok {
		switch cparams := cparams.(type) {
		case map[string]string:
			params = cparams
		case map[string]interface{}:
			for k, v := range cparams {
				params[k] = fmt.Sprintf("%s", v)
			}
		case map[interface{}]interface{}:
			for k, v := range cparams {
				key := fmt.Sprintf("%v", k)
				value := fmt.Sprintf("%v", v)
				params[key] = value
			}
		default:
		}
	}
	maxOpenConns := conf.GetIntOr(prefix+".maxOpenConns", 5)
	maxIdleConns := conf.GetIntOr(prefix+".maxIdleConns", 1)

	cfg := mysql.Config{
		User:                 user,
		Passwd:               passwd,
		Net:                  "tcp",
		Addr:                 addr,
		DBName:               dbName,
		ParseTime:            true,
		Loc:                  time.Now().Location(),
		Collation:            collation,
		Timeout:              time.Duration(timeout) * time.Millisecond,
		ReadTimeout:          time.Duration(readTimeout) * time.Millisecond,
		WriteTimeout:         time.Duration(writeTimeout) * time.Millisecond,
		Params:               params,
		AllowNativePasswords: true,
	}
	if db, err := sql.Open("mysql", cfg.FormatDSN()); err != nil {
		return nil, err
	} else if err = db.Ping(); err != nil {
		return nil, err
	} else {
		db.SetMaxOpenConns(maxOpenConns)
		db.SetMaxIdleConns(maxIdleConns)
		db.SetConnMaxLifetime(15 * time.Minute)
		return db, nil
	}
}
