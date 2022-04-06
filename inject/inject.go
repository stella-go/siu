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

package inject

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/stella-go/siu/common"
)

const _LOG_HEADER = "[SIU]"

var (
	named = &sync.Map{}
	typed = &sync.Map{}
)

type Initializable interface {
	// run after properties set
	Init()
}

var initializableType = reflect.TypeOf((*Initializable)(nil)).Elem()

type ValueResolver interface {
	Resolve(string) (interface{}, bool)
}

func RegisterTyped(refType reflect.Type, obj interface{}) error {
	if _, ok := typed.Load(refType); ok {
		common.ERROR("Typed object %s is already registered", refType)
		return fmt.Errorf("Typed object %s is already registered", refType)
	}
	typed.Store(refType, reflect.ValueOf(obj))
	common.DEBUG("Typed object %s registered", refType)
	return nil
}

func RegisterNamed(name string, obj interface{}) error {
	if _, ok := named.Load(name); ok {
		common.ERROR("Named object \"%s\" is already registered", name)
		return fmt.Errorf("Named object \"%s\" is already registered", name)
	}
	named.Store(name, reflect.ValueOf(obj))
	common.DEBUG("Named object \"%s\" registered", name)
	return nil
}

func GetTyped(refType reflect.Type) (interface{}, bool) {
	return typed.Load(refType)
}

func GetNamed(name string) (interface{}, bool) {
	return named.Load(name)
}

func Inject(r ValueResolver, obj interface{}) error {
	prefType := reflect.TypeOf(obj)
	prefValue := reflect.ValueOf(obj)
	if prefType.Kind() != reflect.Ptr {
		return fmt.Errorf("The object to be injected must be a pointer")
	}
	refType := prefType.Elem()
	refValue := prefValue.Elem()
	if refType.Kind() != reflect.Struct {
		return nil
	}
	common.DEBUG("Process object of type %s", prefType)
	for i := 0; i < refType.NumField(); i++ {
		fieldType := refType.Field(i)
		fieldValue := refValue.Field(i)
		err := setValue(r, fieldType, fieldValue)
		if err != nil {
			common.ERROR("Inject field %s.%s with error: %v", refType, fieldType.Name, err)
			return err
		}
	}
	if prefType.Implements(initializableType) {
		method := prefValue.MethodByName("Init")
		method.Call(nil)
		common.DEBUG("Execute the initialization method of %s", prefType)
	} else if refType.Implements(initializableType) {
		method := refValue.MethodByName("Init")
		method.Call(nil)
		common.DEBUG("Execute the initialization method of %s", refType)
	}
	return nil
}

func setValue(r ValueResolver, field reflect.StructField, val reflect.Value) error {
	tagMap, err := extractTag(field.Tag)
	if err != nil {
		return err
	}
	if _, ok := tagMap["tag"]; !ok {
		return nil
	}
	if !val.CanSet() {
		return fmt.Errorf("field %s can not be set", field.Name)
	}
	if isValueType(field.Type) {
		valueTag, ok := tagMap["value"]
		if ok && strings.HasPrefix(valueTag, "${") && strings.HasSuffix(valueTag, "}") {
			value, err := resolveValue(r, valueTag)
			if err != nil {
				return err
			}
			convertValue := reflect.ValueOf(value).Convert(field.Type)
			val.Set(convertValue)
		} else {
			convertValue := reflect.ValueOf(valueTag).Convert(field.Type)
			val.Set(convertValue)
		}
	} else {
		switch field.Type.Kind() {
		case reflect.Interface:
			value, zero, err := resolveInterface(tagMap, r, field.Type)
			if err != nil {
				return err
			}
			if zero {
				return nil
			}
			val.Set(value)
		case reflect.Struct:
			value, _, err := resolveStruct(tagMap, r, field.Type)
			if err != nil {
				return err
			}
			val.Set(value)
		case reflect.Ptr:
			value, zero, err := resolvePtr(tagMap, r, field.Type)
			if err != nil {
				return err
			}
			if zero {
				return nil
			}
			val.Set(value)
		default:
			value, err := create(r, field.Type)
			if err != nil {
				return err
			}
			val.Set(value)
		}
	}
	return nil
}

func resolveInterface(tagMap map[string]string, r ValueResolver, typ reflect.Type) (reflect.Value, bool, error) {
	if name, ok := tagMap["name"]; ok {
		if v, ok := named.Load(name); ok {
			common.DEBUG("Found interface %s with name \"%s\"", typ, name)
			return v.(reflect.Value), false, nil
		} else {
			if defaultValue, ok := tagMap["default"]; ok && defaultValue == "zero" {
				common.DEBUG("Not found interface %s with name \"%s\", default is zero", typ, name)
				return reflect.Value{}, true, nil
			}
		}
	}
	if v, ok := typed.Load(typ); ok {
		common.DEBUG("Found interface %s with type \"%s\"", typ, typ)
		return v.(reflect.Value), false, nil
	}
	common.ERROR("Not found interface %s", typ)
	return reflect.Value{}, true, fmt.Errorf("interface type %s not found", typ.Name())
}

func resolveStruct(tagMap map[string]string, r ValueResolver, typ reflect.Type) (reflect.Value, bool, error) {
	value, err := create(r, typ)
	if err != nil {
		return reflect.Value{}, true, nil
	}
	common.DEBUG("Create object %s", typ)
	return value, false, nil
}

func resolvePtr(tagMap map[string]string, r ValueResolver, typ reflect.Type) (reflect.Value, bool, error) {
	if t, ok := tagMap["type"]; ok && t == "private" {
		value, err := create(r, typ)
		if err != nil {
			return reflect.Value{}, true, nil
		}
		common.DEBUG("Create private object %s", typ)
		return value, false, nil
	}
	if name, ok := tagMap["name"]; ok {
		if v, ok := named.Load(name); ok {
			common.DEBUG("Found object %s with name \"%s\"", typ, name)
			return v.(reflect.Value), false, nil
		} else {
			if defaultValue, ok := tagMap["default"]; ok && defaultValue == "zero" {
				common.DEBUG("Not found object %s with name \"%s\", default is zero", typ, name)
				return reflect.Value{}, true, nil
			}
		}
	}
	if v, ok := typed.Load(typ); ok {
		common.DEBUG("Found object %s with type %s", typ, typ)
		return v.(reflect.Value), false, nil
	}
	value, err := create(r, typ)
	if err != nil {
		return reflect.Value{}, true, err
	}
	common.DEBUG("Create object %s", typ)
	if name, ok := tagMap["name"]; ok {
		named.Store(name, value)
	}
	typed.Store(typ, value)
	return value, false, nil
}

func create(r ValueResolver, typ reflect.Type) (reflect.Value, error) {
	var value reflect.Value
	switch typ.Kind() {
	case reflect.Struct:
		pvalue := reflect.New(typ)
		err := Inject(r, pvalue.Interface())
		if err != nil {
			return value, err
		}
		value = pvalue.Elem()
		return value, nil
	case reflect.Ptr:
		value = reflect.New(typ.Elem())
		err := Inject(r, value.Interface())
		if err != nil {
			return reflect.Value{}, err
		}
		return value, nil
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		v := reflect.New(typ)
		return v.Elem(), nil
	default:
		value := reflect.Zero(typ)
		err := Inject(r, value.Interface())
		if err != nil {
			return reflect.Value{}, err
		}
		return value, nil
	}
}

func isValueType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		return true
	default:
		return false
	}
}

func resolveValue(r ValueResolver, placeholder string) (interface{}, error) {
	key := placeholder[2 : len(placeholder)-1]
	if splits := strings.Split(key, ":"); len(splits) > 1 {
		key, defaultValue := splits[0], splits[1]
		if v, ok := r.Resolve(key); ok {
			return v, nil
		} else {
			return defaultValue, nil
		}
	}
	if v, ok := r.Resolve(key); ok {
		return v, nil
	} else {
		return nil, fmt.Errorf("%s not found", placeholder)
	}
}

func extractTag(structTag reflect.StructTag) (map[string]string, error) {
	// tag example `@siu-inject:"name='abc',value='${a.b.c}',default='type',type='private'"`
	r := make(map[string]string)
	tag, ok := structTag.Lookup("@siu")
	if !ok {
		return r, nil
	}
	syntaxErr := fmt.Errorf("@siu inject tag syntax error: %s", tag)
	r["tag"] = tag
	for tag != "" {
		// skip leading space and comma
		i := 0
		for i < len(tag) && (tag[i] == ' ' || tag[i] == ',') {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}
		// scan to equals mark.
		// a space or a quote is a syntax error
		i = 0
		for i < len(tag) && tag[i] != ' ' && tag[i] != '\'' && tag[i] != '=' {
			i++
		}
		if i+1 >= len(tag) || tag[i] != '=' || tag[i+1] != '\'' {
			return nil, syntaxErr
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// scan quoted string to find value
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

		tag = tag[i+1:]
	}
	return r, nil
}
