package g

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/stella-go/siu/t"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func Create[T any](db *gorm.DB, s *T) error {
	if s == nil {
		return fmt.Errorf("pointer is nil")
	}
	r := db.Model(s).Create(s)
	return r.Error
}

func Create2[T any](db *gorm.DB, s *T) (int64, error) {
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
	values := make(map[string]interface{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		setting := schema.ParseTagSetting(f.Tag.Get("gorm"), ";")
		if val, ok := setting["-"]; ok && (val == "-" || val == "all") {
			continue
		}
		column := ""
		if value, ok := setting["COLUMN"]; ok {
			column = value
		} else {
			column = db.NamingStrategy.ColumnName("", f.Name)
		}
		fv := rv.Field(i)
		if fv.IsNil() {
			continue
		}
		if t.IsNull(fv.Interface()) {
			values[column] = gorm.Expr("NULL")
		} else {
			values[column] = fv.Interface()
		}
	}
	copied := make(map[string]interface{})
	for k, v := range values {
		copied[k] = v
	}
	r := db.Model(s).Create(copied)
	if r.Error != nil {
		return 0, r.Error
	}
	var ILastInsertId interface{}
	for k, v := range copied {
		if _, ok := values[k]; !ok {
			ILastInsertId = v
			break
		}
	}
	var lastInsertId int64
	if ILastInsertId != nil {
		switch id := ILastInsertId.(type) {
		case int:
			lastInsertId = int64(id)
		case int32:
			lastInsertId = int64(id)
		case int64:
			lastInsertId = int64(id)
		case uint:
			lastInsertId = int64(id)
		case uint32:
			lastInsertId = int64(id)
		case uint64:
			lastInsertId = int64(id)
		}
	}
	return lastInsertId, r.Error
}

func Update[T any](db *gorm.DB, s *T) (int64, error) {
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
	updates := make(map[string]interface{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		setting := schema.ParseTagSetting(f.Tag.Get("gorm"), ";")
		if val, ok := setting["-"]; ok && (val == "-" || val == "all") {
			continue
		}
		column := ""
		if value, ok := setting["COLUMN"]; ok {
			column = value
		} else {
			column = db.NamingStrategy.ColumnName("", f.Name)
		}
		fv := rv.Field(i)
		if fv.IsNil() {
			continue
		}
		if t.IsNull(fv.Interface()) {
			updates[column] = gorm.Expr("NULL")
		} else {
			updates[column] = fv.Interface()
		}
	}
	r := db.Model(s).Updates(updates)
	return r.RowsAffected, r.Error
}
func Update2[T any](db *gorm.DB, s *T) (int64, error) {
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
	updates := make(map[string]interface{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		setting := schema.ParseTagSetting(f.Tag.Get("gorm"), ";")
		if val, ok := setting["-"]; ok && (val == "-" || val == "all") {
			continue
		}
		column := ""
		if value, ok := setting["COLUMN"]; ok {
			column = value
		} else {
			column = db.NamingStrategy.ColumnName("", f.Name)
		}
		fv := rv.Field(i)
		if fv.IsNil() {
			updates[column] = gorm.Expr("NULL")
		} else {
			updates[column] = fv.Interface()
		}
	}
	r := db.Model(s).Updates(updates)
	return r.RowsAffected, r.Error
}

func Query[T any](db *gorm.DB, s *T) (*T, error) {
	if s == nil {
		return nil, fmt.Errorf("pointer is nil")
	}
	var empty T
	ss := &empty
	r := db.Model(s).Where(s).Take(&ss)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, r.Error
		}
	}
	return ss, nil
}

func QueryMany[T any](db *gorm.DB, s *T, page int, size int) (int, []*T, error) {
	stmt := db.Model(s).Where(s)
	var count int64
	many := make([]*T, 0)
	r := stmt.Count(&count)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return 0, many, nil
		} else {
			return 0, nil, r.Error
		}
	}
	r = stmt.Offset((page - 1) * size).Limit(size).Find(&many)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return 0, many, nil
		} else {
			return 0, nil, r.Error
		}
	}
	return int(count), many, nil
}

func QueryExec[T any](db *gorm.DB) (*T, error) {
	var empty T
	ss := &empty
	r := db.Take(&ss)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, r.Error
		}
	}
	return ss, nil
}

func QueryExecMany[T any](db *gorm.DB) ([]*T, error) {
	stmt := db
	many := make([]*T, 0)
	r := stmt.Find(&many)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return many, nil
		} else {
			return nil, r.Error
		}
	}
	return many, nil
}

func Delete[T any](db *gorm.DB, s *T) (int64, error) {
	if s == nil {
		return 0, fmt.Errorf("pointer is nil")
	}
	r := db.Model(s).Delete(s, s)
	return r.RowsAffected, r.Error
}
