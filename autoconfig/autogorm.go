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

package autoconfig

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/interfaces"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	GormDatasourceKey        = "gorm"
	GormDatasourceDisableKey = GormDatasourceKey + ".disable"
	GormDatasourceOrder      = 20
)

type AutoGorm struct {
	Conf   config.TypedConfig `@siu:"name='environment',default='type'"`
	Logger interfaces.Logger  `@siu:"name='logger',default='type'"`
	dbs    map[string]*gorm.DB
}

func (p *AutoGorm) Condition() bool {
	_, ok1 := p.Conf.Get(GormDatasourceKey)
	v, ok2 := p.Conf.GetBool(GormDatasourceDisableKey)

	if ok2 && v {
		return false
	}
	return ok1
}

func (p *AutoGorm) OnStart() error {
	p.dbs = make(map[string]*gorm.DB)

	datasources, _ := p.Conf.Get(GormDatasourceKey)

	db, err := createGormDB(p.Logger, p.Conf, GormDatasourceKey)
	if err != nil {
		return err
	}
	if db != nil {
		p.dbs[GormDatasourceKey] = db
	}

	datasourcesMap, ok := datasources.(map[string]interface{})
	if !ok {
		return nil
	}
	for datasource := range datasourcesMap {
		name := GormDatasourceKey + "." + datasource
		db, err := createGormDB(p.Logger, p.Conf, name)
		if err != nil {
			return err
		}
		if db != nil {
			p.dbs[name] = db
		}
	}
	return nil
}

func (p *AutoGorm) OnStop() error {
	for _, db := range p.dbs {
		if db != nil {
			if ins, err := db.DB(); err == nil {
				ins.Close()
			}
		}
	}
	return nil
}

func (*AutoGorm) Order() int {
	return GormDatasourceOrder
}

func (*AutoGorm) Name() string {
	return GormDatasourceKey
}

func (p *AutoGorm) Named() map[string]interface{} {
	n := make(map[string]interface{})
	for k, v := range p.dbs {
		n[k] = v
	}
	return n
}

func (p *AutoGorm) Typed() map[reflect.Type]interface{} {
	if len(p.dbs) == 1 {
		for _, v := range p.dbs {
			refType := reflect.TypeOf((*gorm.DB)(nil))
			return map[reflect.Type]interface{}{
				refType: v,
			}
		}
	}
	return nil
}

func createGormDB(logger interfaces.Logger, conf config.TypedConfig, prefix string) (*gorm.DB, error) {
	user, ok1 := conf.GetString(prefix + ".user")
	passwd, ok2 := conf.GetString(prefix + ".passwd")
	addr, ok3 := conf.GetString(prefix + ".addr")
	dbName, ok4 := conf.GetString(prefix + ".dbName")

	if !ok1 || !ok2 || !ok3 || !ok4 {
		if ok1 || ok2 || ok3 || ok4 {
			return nil, fmt.Errorf("gorm datasource user/passwd/addr/dbName can not be empty")
		}
		return nil, nil
	}

	collation := conf.GetStringOr(prefix+".collation", "utf8mb4_bin")
	timeout := conf.GetIntOr(prefix+".timeout", 60000)
	readTimeout := conf.GetIntOr(prefix+".readTimeout", 30000)
	writeTimeout := conf.GetIntOr(prefix+".writeTimeout", 30000)
	loc := time.Local
	if loca, ok := conf.GetString(prefix + ".loc"); ok {
		if locl, err := time.LoadLocation(loca); err == nil {
			loc = locl
		}
	}

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

	cfg := driver.Config{
		User:                 user,
		Passwd:               passwd,
		Net:                  "tcp",
		Addr:                 addr,
		DBName:               dbName,
		ParseTime:            true,
		Loc:                  loc,
		Collation:            collation,
		Timeout:              time.Duration(timeout) * time.Millisecond,
		ReadTimeout:          time.Duration(readTimeout) * time.Millisecond,
		WriteTimeout:         time.Duration(writeTimeout) * time.Millisecond,
		Params:               params,
		AllowNativePasswords: true,
	}

	option := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}

	debug := conf.GetStringOr("logger.level", "info")
	if strings.ToLower(debug) == "debug" {
		option.Logger = &Logger{inner: logger}
	}

	if db, err := gorm.Open(mysql.Open(cfg.FormatDSN()), option); err != nil {
		return nil, err
	} else {
		if ins, err := db.DB(); err != nil {
			return nil, err
		} else {
			if err := ins.Ping(); err != nil {
				return nil, err
			} else {
				ins.SetMaxOpenConns(maxOpenConns)
				ins.SetMaxIdleConns(maxIdleConns)
				ins.SetConnMaxLifetime(15 * time.Minute)
				return db, nil
			}
		}
	}
}

type Logger struct {
	inner interfaces.Logger
}

func (p *Logger) LogMode(level logger.LogLevel) logger.Interface {
	return p
}
func (p *Logger) Info(ctx context.Context, format string, args ...interface{}) {
	p.inner.INFO("%s", "[GORM] "+fmt.Sprintf(format, args...))
}
func (p *Logger) Warn(ctx context.Context, format string, args ...interface{}) {
	p.inner.WARN("%s", "[GORM] "+fmt.Sprintf(format, args...))
}
func (p *Logger) Error(ctx context.Context, format string, args ...interface{}) {
	p.inner.ERROR("%s", "[GORM] "+fmt.Sprintf(format, args...))
}
func (p *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, _ := fc()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			p.inner.DEBUG("%s", "[GORM] "+fmt.Sprintf("SQL: %s, ERROR: %v", sql, err))
		}
	} else {
		p.inner.DEBUG("%s", "[GORM] "+fmt.Sprintf("SQL: %s", sql))
	}
}
