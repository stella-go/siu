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

package inject

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/stella-go/siu/common"
)

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
		return fmt.Errorf("typed object %s is already registered", refType)
	}
	typed.Store(refType, reflect.ValueOf(obj))
	common.DEBUG("Typed object %s registered", refType)
	return nil
}

func RegisterNamed(name string, obj interface{}) error {
	if _, ok := named.Load(name); ok {
		common.ERROR("Named object \"%s\" is already registered", name)
		return fmt.Errorf("named object \"%s\" is already registered", name)
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
	defer func() {
		if err := recover(); err != nil {
			common.ERROR("panic:", err)
			panic(err)
		}
	}()
	return inject(r, obj)
}

func inject(r ValueResolver, obj interface{}) error {
	prefType := reflect.TypeOf(obj)
	prefValue := reflect.ValueOf(obj)
	if prefType.Kind() != reflect.Ptr {
		return fmt.Errorf("the object to be injected must be a pointer")
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
			common.ERROR("Inject field %s.%s with error:", refType, fieldType.Name, err)
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
	tag, ok := field.Tag.Lookup("@siu")
	if !ok {
		return nil
	}
	if !val.CanSet() {
		return fmt.Errorf("field %s can not be set", field.Name)
	}
	common.DEBUG("Process field %s", field.Name)

	tagMap, err := extractTag(tag)
	if err != nil {
		return err
	}
	if isValueType(field.Type) {
		value, zero, err := resolveValue(tagMap, r, field.Type)
		if err != nil {
			return err
		}
		if zero {
			return nil
		}
		val.Set(value)
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
		case reflect.Ptr:
			value, zero, err := resolvePtr(tagMap, r, field.Type)
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
	} else {
		if defaultValue, ok := tagMap["default"]; ok && defaultValue == "zero" {
			common.DEBUG("Not found interface %s with type %s, default is zero", typ, typ)
			return reflect.Value{}, true, nil
		}
	}
	common.ERROR("Not found interface %s", typ)
	return reflect.Value{}, true, fmt.Errorf("interface type %s not found", typ)
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
	} else {
		if defaultValue, ok := tagMap["default"]; ok && defaultValue == "zero" {
			common.DEBUG("Not found object %s with type %s, default is zero", typ, typ)
			return reflect.Value{}, true, nil
		}
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

func resolveStruct(tagMap map[string]string, r ValueResolver, typ reflect.Type) (reflect.Value, bool, error) {
	value, err := create(r, typ)
	if err != nil {
		return reflect.Value{}, true, nil
	}
	common.DEBUG("Create object %s", typ)
	return value, false, nil
}

func create(r ValueResolver, typ reflect.Type) (reflect.Value, error) {
	var value reflect.Value
	switch typ.Kind() {
	case reflect.Struct:
		pvalue := reflect.New(typ)
		err := inject(r, pvalue.Interface())
		if err != nil {
			return value, err
		}
		value = pvalue.Elem()
		return value, nil
	case reflect.Ptr:
		value = reflect.New(typ.Elem())
		err := inject(r, value.Interface())
		if err != nil {
			return reflect.Value{}, err
		}
		return value, nil
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		v := reflect.New(typ)
		return v.Elem(), nil
	default:
		value := reflect.Zero(typ)
		err := inject(r, value.Interface())
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

func resolveValue(tagMap map[string]string, r ValueResolver, typ reflect.Type) (reflect.Value, bool, error) {
	if valueTag, ok := tagMap["value"]; ok {
		if strings.HasPrefix(valueTag, "${") && strings.HasSuffix(valueTag, "}") {
			placeholder := valueTag[2 : len(valueTag)-1]
			if placeholder == "" {
				return resolveValueWithDefault(tagMap, r, typ)
			}
			if index := strings.Index(placeholder, ":"); index != -1 {
				key, defaultValue := placeholder[0:index], placeholder[index+1:]
				if value, ok := r.Resolve(key); ok {
					convertValue := reflect.ValueOf(value).Convert(typ)
					common.DEBUG("Found type %s value with key \"%s\": \"%v\"", typ, valueTag, value)
					return convertValue, convertValue.IsZero(), nil
				} else {
					convertValue, err := convertString(defaultValue, typ)
					if err != nil {
						return reflect.Value{}, true, err
					}
					common.DEBUG("Not found type %s value with key \"%s\", use default value: \"%v\"", typ, valueTag, defaultValue)
					return convertValue, convertValue.IsZero(), nil
				}
			} else {
				if value, ok := r.Resolve(placeholder); ok {
					convertValue := reflect.ValueOf(value).Convert(typ)
					common.DEBUG("Found type %s value with key \"%s\": \"%v\"", typ, valueTag, value)
					return convertValue, convertValue.IsZero(), nil
				} else {
					return resolveValueWithDefault(tagMap, r, typ)
				}
			}
		} else {
			convertValue, err := convertString(valueTag, typ)
			if err != nil {
				return reflect.Value{}, true, err
			}
			common.DEBUG("type %s value with key \"%s\", the key is not a placeholder, use the value: \"%v\"", typ, valueTag, valueTag)
			return convertValue, convertValue.IsZero(), nil
		}
	} else {
		return resolveValueWithDefault(tagMap, r, typ)
	}
}

func resolveValueWithDefault(tagMap map[string]string, r ValueResolver, typ reflect.Type) (reflect.Value, bool, error) {
	if defaultValue, ok := tagMap["default"]; ok {
		convertValue, err := convertString(defaultValue, typ)
		if err != nil {
			return reflect.Value{}, true, err
		}
		common.DEBUG("Not found type %s value with key \"%s\", use default value: \"%v\"", typ, tagMap["value"], defaultValue)
		return convertValue, convertValue.IsZero(), nil
	}
	return reflect.Value{}, true, fmt.Errorf("type %s value with key \"%s\" not found, default is not set", typ, tagMap["value"])
}

func convertString(value string, typ reflect.Type) (reflect.Value, error) {
	switch typ.Kind() {
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(boolValue), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(intValue).Convert(typ), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uintValue).Convert(typ), nil
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(floatValue).Convert(typ), nil
	case reflect.Complex64, reflect.Complex128:
		complexValue, err := strconv.ParseComplex(value, 128)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(complexValue).Convert(typ), nil
	}
	return reflect.ValueOf(value).Convert(typ), nil
}

func extractTag(tag string) (map[string]string, error) {
	// tag example `@siu:"name='abc',value='${a.b.c}',default='type',type='private'"`
	r := make(map[string]string)

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
