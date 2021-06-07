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
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stella-go/siu/config"
)

const (
	MySQLDatasourceKey   = "mysql"
	MySQLDatasourceOrder = 20
)

type AutoMysql struct {
	dbs map[string]*sql.DB
}

func (*AutoMysql) Condition() bool {
	_, ok := config.Get(MySQLDatasourceKey)
	return ok
}

func (p *AutoMysql) OnStart() error {
	p.dbs = make(map[string]*sql.DB)

	datasources, _ := config.Get(MySQLDatasourceKey)

	db, err := createDB(MySQLDatasourceKey)
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
		db, err := createDB(MySQLDatasourceKey + "." + datasource)
		if err != nil {
			return err
		}
		if db != nil {
			p.dbs[datasource] = db
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

func (p *AutoMysql) Get() interface{} {
	return p.dbs
}

func createDB(prefix string) (*sql.DB, error) {
	user, ok1 := config.GetString(prefix + ".user")
	passwd, ok2 := config.GetString(prefix + ".passwd")
	addr, ok3 := config.GetString(prefix + ".addr")
	dbName, ok4 := config.GetString(prefix + ".dbName")

	if !ok1 || !ok2 || !ok3 || !ok4 {
		if prefix == MySQLDatasourceKey {
			return nil, nil
		}
		return nil, fmt.Errorf("mysql datasource user/passwd/addr/dbName can not be empty")
	}

	collation := config.GetStringOr(prefix+".collation", "utf8mb4_bin")
	timeout := config.GetIntOr(prefix+".timeout", 60000)
	readTimeout := config.GetIntOr(prefix+".readTimeout", 30000)
	writeTimeout := config.GetIntOr(prefix+".writeTimeout", 30000)

	cparams, ok := config.Get(prefix + ".params")
	params := make(map[string]string)
	if ok {
		switch cparams := cparams.(type) {
		case map[string]string:
			params = cparams
		case map[string]interface{}:
			for k, v := range cparams {
				params[k] = fmt.Sprintf("%s", v)
			}
		default:
		}
	}
	maxOpenConns := config.GetIntOr(prefix+".maxOpenConns", 5)
	maxIdleConns := config.GetIntOr(prefix+".maxIdleConns", 1)

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
