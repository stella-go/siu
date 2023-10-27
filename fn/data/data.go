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

package data

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/stella-go/siu/t/n"
)

const (
	tag_free         = "@free"
	primary          = "primary"
	autoincrment     = "auto-incrment"
	table            = "table"
	column           = "column"
	currenttimestamp = "current-timestamp"
	round            = "round"
	ignore           = "ignore"

	s_true = "true"
)

func Create[T any](db *sql.DB, s *T) (int64, error) {
	if s == nil {
		return 0, fmt.Errorf("pointer is nil")
	}
	rt := reflect.TypeOf(s)
	rv := reflect.ValueOf(s).Elem()
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	table := toSnakeCase(rt.Name())
	columns := make([]string, 0)
	holders := make([]string, 0)
	args := make([]interface{}, 0)
	SQL := "insert into `%s` (%s) values (%s)"
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag, err := extractTag(f.Tag.Get(tag_free))
		if err != nil {
			return 0, err
		}
		if value, ok := tag[ignore]; ok && value == s_true {
			continue
		}
		if value, ok := tag[table]; ok {
			table = value
		}
		if value, ok := tag[autoincrment]; ok && value == s_true {
			continue
		}
		if value, ok := tag[currenttimestamp]; ok && value == s_true {
			continue
		}
		fv := rv.Field(i)
		if fv.IsNil() {
			continue
		}
		if value, ok := tag[column]; ok {
			columns = append(columns, "`"+value+"`")
		} else {
			columns = append(columns, "`"+toSnakeCase(f.Name)+"`")
		}
		v := roundIfTime(fv.Interface(), tag[round])
		args = append(args, v)
		holders = append(holders, "?")
	}
	SQL = fmt.Sprintf(SQL, table, strings.Join(columns, ", "), strings.Join(holders, ", "))
	ret, err := db.Exec(SQL, args...)
	if err != nil {
		return 0, err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

func Update[T any](db *sql.DB, s *T) (int64, error) {
	if s == nil {
		return 0, fmt.Errorf("pointer is nil")
	}
	rt := reflect.TypeOf(s)
	rv := reflect.ValueOf(s)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	table := toSnakeCase(rt.Name())
	set := make([]string, 0)
	where := make([]string, 0)
	args := make([]interface{}, 0)
	whereArgs := make([]interface{}, 0)
	SQL := "update `%s` set %s where %s"
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag, err := extractTag(f.Tag.Get(tag_free))
		if err != nil {
			return 0, err
		}
		if value, ok := tag[ignore]; ok && value == s_true {
			continue
		}
		if value, ok := tag[table]; ok {
			table = value
		}
		fv := rv.Field(i)
		if fv.IsNil() {
			continue
		}
		var column string
		if value, ok := tag["column"]; ok {
			column = value
		} else {
			column = toSnakeCase(f.Name)
		}
		v := roundIfTime(fv.Interface(), tag[round])
		if value, ok := tag[primary]; ok && value == s_true {
			where = append(where, fmt.Sprintf("`%s` = ?", column))
			whereArgs = append(whereArgs, v)
			continue
		}
		set = append(set, fmt.Sprintf("`%s` = ?", column))
		args = append(args, v)
	}
	if len(where) == 0 {
		return 0, fmt.Errorf("primary not found, where condition empty")
	}
	SQL = fmt.Sprintf(SQL, table, strings.Join(set, ", "), strings.Join(where, ", "))
	ret, err := db.Exec(SQL, append(args, whereArgs...)...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

func Query[T any](db *sql.DB, s *T) (*T, error) {
	if s == nil {
		return nil, fmt.Errorf("pointer is nil")
	}
	rt := reflect.TypeOf(s)
	rv := reflect.ValueOf(s)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	table := toSnakeCase(rt.Name())
	where := make([]string, 0)
	whereArgs := make([]interface{}, 0)
	SQL := "select * from `%s` %s limit 1"
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag, err := extractTag(f.Tag.Get(tag_free))
		if err != nil {
			return nil, err
		}
		if value, ok := tag[ignore]; ok && value == s_true {
			continue
		}
		if value, ok := tag[table]; ok {
			table = value
		}
		fv := rv.Field(i)
		if fv.IsNil() {
			continue
		}
		var column string
		if value, ok := tag["column"]; ok {
			column = value
		} else {
			column = toSnakeCase(f.Name)
		}
		v := roundIfTime(fv.Interface(), tag[round])
		where = append(where, fmt.Sprintf("`%s` = ?", column))
		whereArgs = append(whereArgs, v)
	}
	sWhere := strings.Join(where, " and ")
	if sWhere != "" {
		sWhere = "where " + sWhere
	}

	ret, scan, err := newScan[T](rt)
	if err != nil {
		return nil, err
	}
	SQL = fmt.Sprintf(SQL, table, sWhere)
	err = db.QueryRow(SQL, whereArgs...).Scan(scan...)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}
	return ret, nil
}

func QueryMany[T any](db *sql.DB, s *T, page int, size int) (int, []*T, error) {
	rt := reflect.TypeOf(s)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	SQL1 := "select count(*) from `%s` %s"
	SQL2 := "select * from `%s` %s limit ?, ?"

	table := toSnakeCase(rt.Name())
	where := make([]string, 0)
	whereArgs := make([]interface{}, 0)
	if s != nil {
		rv := reflect.ValueOf(s)
		if rv.Kind() == reflect.Pointer {
			rv = rv.Elem()
		}
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			tag, err := extractTag(f.Tag.Get(tag_free))
			if err != nil {
				return 0, nil, err
			}
			if value, ok := tag[ignore]; ok && value == s_true {
				continue
			}
			if value, ok := tag[table]; ok {
				table = value
			}
			fv := rv.Field(i)
			if fv.IsNil() {
				continue
			}
			var column string
			if value, ok := tag["column"]; ok {
				column = value
			} else {
				column = toSnakeCase(f.Name)
			}
			v := roundIfTime(fv.Interface(), tag[round])
			where = append(where, fmt.Sprintf("`%s` = ?", column))
			whereArgs = append(whereArgs, v)
		}
	}
	sWhere := strings.Join(where, " and ")
	if sWhere != "" {
		sWhere = "where " + sWhere
	}

	SQL1 = fmt.Sprintf(SQL1, table, sWhere)
	SQL2 = fmt.Sprintf(SQL2, table, sWhere)
	count := 0
	err := db.QueryRow(SQL1, whereArgs...).Scan(&count)
	if err != nil {
		return 0, nil, err
	}

	whereArgs = append(whereArgs, (page-1)*size, size)
	rows, err := db.Query(SQL2, whereArgs...)
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, nil, err
		}
	}
	defer rows.Close()

	results := make([]*T, 0)
	for rows.Next() {
		ret, scan, err := newScan[T](rt)
		if err != nil {
			return 0, nil, err
		}
		err = rows.Scan(scan...)
		if err != nil {
			return 0, nil, err
		}
		results = append(results, ret)
	}
	return count, results, nil
}

func newScan[T any](rt reflect.Type) (*T, []interface{}, error) {
	rv := reflect.New(rt)
	s := rv.Interface().(*T)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	scan := make([]interface{}, 0)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag, err := extractTag(f.Tag.Get(tag_free))
		if err != nil {
			return nil, nil, err
		}
		if value, ok := tag[ignore]; ok && value == s_true {
			continue
		}
		fv := rv.Field(i)
		scan = append(scan, fv.Addr().Interface())
	}
	return s, scan, nil
}

func Delete[T any](db *sql.DB, s *T) (int64, error) {
	if s == nil {
		return 0, fmt.Errorf("pointer is nil")
	}
	rt := reflect.TypeOf(s)
	rv := reflect.ValueOf(s)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	table := toSnakeCase(rt.Name())
	where := make([]string, 0)
	whereArgs := make([]interface{}, 0)
	SQL := "delete from `%s` where %s"
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag, err := extractTag(f.Tag.Get(tag_free))
		if err != nil {
			return 0, err
		}
		if value, ok := tag[ignore]; ok && value == s_true {
			continue
		}
		if value, ok := tag[table]; ok {
			table = value
		}
		if value, ok := tag[primary]; ok && value == s_true {
			var column string
			if value, ok := tag["column"]; ok {
				column = value
			} else {
				column = toSnakeCase(f.Name)
			}
			fv := rv.Field(i)
			if fv.IsNil() {
				return 0, fmt.Errorf("primary %s is empty", column)
			}
			v := roundIfTime(fv.Interface(), tag[round])
			where = append(where, fmt.Sprintf("`%s` = ?", column))
			whereArgs = append(whereArgs, v)
		}
	}
	if len(where) == 0 {
		return 0, fmt.Errorf("primary not found, where condition empty")
	}
	SQL = fmt.Sprintf(SQL, table, strings.Join(where, ", "))
	ret, err := db.Exec(SQL, whereArgs...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

func roundIfTime(v interface{}, round string) interface{} {
	var r time.Duration
	switch round {
	case "s", s_true:
		r = time.Second
	case "ms", "milli":
		r = time.Millisecond
	case "μs", "us", "micro":
		r = time.Microsecond
	}
	switch v := v.(type) {
	case time.Time:
		return v.Round(r)
	case *n.Time:
		v.Val = v.Val.Round(r)
		return v
	default:
		if tv, ok := v.(time.Time); ok {
			return tv.Round(r)
		} else {
			return v
		}
	}
}
func toSnakeCase(s string) string {
	re := regexp.MustCompile(`[A-Z]`)
	snake := re.ReplaceAllStringFunc(s, func(s string) string {
		return "_" + strings.ToLower(s[:1])
	})
	return strings.Trim(snake, "_")
}

func extractTag(tag string) (map[string]string, error) {
	// tag example `@free:"table='a_table',column='a_column',primary,auto-incrment,current-timestamp,round='s',ignore"`
	r := make(map[string]string)

	syntaxErr := fmt.Errorf("%s tag syntax error: %s", tag_free, tag)
	r["tag"] = tag
	for tag != "" {
		i := 0
		for i < len(tag) && (tag[i] == ' ' || tag[i] == ',') {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}
		i = 0
		for i < len(tag) && tag[i] != ' ' && tag[i] != '\'' && tag[i] != '=' && tag[i] != ',' {
			i++
		}
		if i >= len(tag) {
			name := string(tag[:i])
			value := s_true
			r[name] = value
			break
		}
		switch tag[i] {
		case ',':
			name := string(tag[:i])
			value := s_true
			r[name] = value
		case '=':
			name := string(tag[:i])
			tag = tag[i+1:]
			i = 1
			for i < len(tag) && tag[i] != '\'' {
				if tag[i] == '\\' {
					i++
				}
				i++
			}
			if i >= len(tag) {
				return nil, syntaxErr
			}
			value := strings.ReplaceAll(tag[1:i], "\\", "")
			r[name] = value
		default:
			return nil, syntaxErr
		}
		tag = tag[i+1:]
	}
	return r, nil
}
