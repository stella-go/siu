// Copyright 2010-2026 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package meta

import (
	"reflect"
	"strings"
	"time"
)

// ParamDef defines a single parameter for route metadata.
// When Type is "object", use Properties to describe nested fields.
// When Type is "array", use Items to describe the element type.
type ParamDef struct {
	Type        string // "string", "integer", "boolean", "number", "object", "array"
	Format      string // optional: "byte" (base64 []byte), "binary" (file upload)
	Description string
	Required    bool
	Properties  map[string]ParamDef // nested fields when Type is "object"
	Items       *ParamDef           // element type when Type is "array"
}

// RouteDef provides typed metadata for unified Swagger + MCP generation.
type RouteDef struct {
	Name        string              // MCP tool name (optional, auto-derived from method+path)
	Summary     string              // → swagger.summary + mcp.description
	Params      map[string]ParamDef // → swagger parameters/requestBody + mcp inputSchema
	Request     any                 // request struct instance; fields are parsed via @meta tag
	Response    any                 // response struct instance; fields are parsed via @meta tag
	ContentType string              // request body content type (default: "application/json"; set "multipart/form-data" for file uploads)
}

// ParamSchema converts a ParamDef into a JSON-Schema-compatible map used by
// both OpenAPI (Swagger) and MCP inputSchema.
func ParamSchema(p ParamDef) map[string]any {
	schema := map[string]any{"type": p.Type}
	if p.Format != "" {
		schema["format"] = p.Format
	}
	if p.Description != "" {
		schema["description"] = p.Description
	}
	if p.Type == "object" && len(p.Properties) > 0 {
		props := make(map[string]any, len(p.Properties))
		var req []string
		for name, child := range p.Properties {
			props[name] = ParamSchema(child)
			if child.Required {
				req = append(req, name)
			}
		}
		schema["properties"] = props
		if len(req) > 0 {
			schema["required"] = req
		}
	}
	if p.Type == "array" && p.Items != nil {
		schema["items"] = ParamSchema(*p.Items)
	}
	return schema
}

// parseMetaTag parses a @meta struct tag value in "k=v,k=v" format.
// Supported keys: desc (description), required (boolean flag), ignore (boolean flag).
func parseMetaTag(tag string) (string, bool, bool) {
	if tag == "" {
		return "", false, false
	}
	var desc string
	var required bool
	var ignore bool
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			switch strings.TrimSpace(kv[0]) {
			case "desc":
				desc = strings.TrimSpace(kv[1])
			case "required":
				required = strings.TrimSpace(kv[1]) == "true"
			case "ignore":
				ignore = strings.TrimSpace(kv[1]) == "true"
			}
		} else if part == "required" {
			required = true
		} else if part == "ignore" {
			ignore = true
		}
	}
	return desc, required, ignore
}

// goTypeToSchemaType maps a Go reflect.Type to a JSON Schema type string.
func goTypeToSchemaType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// time.Time → string
	if t == reflect.TypeOf(time.Time{}) {
		return "string"
	}
	// []byte → string (base64)
	if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return "string"
	}
	// n.Nullxxx types → map to the underlying Val field type
	if valType, ok := nullableValType(t); ok {
		return goTypeToSchemaType(valType)
	}
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Struct:
		return "object"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map:
		return "object"
	default:
		return "string"
	}
}

// goTypeToSchemaFormat returns the JSON Schema format for special Go types.
// Returns "byte" for []byte (base64-encoded).
func goTypeToSchemaFormat(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return "byte"
	}
	return ""
}

// nullableValType checks whether t is a nullable type from the t/n package.
// These types are structs with a single exported field "Val".
// Returns the Val field's type and true if matched.
func nullableValType(t reflect.Type) (reflect.Type, bool) {
	if t.Kind() != reflect.Struct {
		return nil, false
	}
	if t.PkgPath() != "github.com/stella-go/siu/t/n" {
		return nil, false
	}
	if valField, ok := t.FieldByName("Val"); ok {
		return valField.Type, true
	}
	return nil, false
}

// isNullableType returns true if t is a nullable type from the t/n package.
func isNullableType(t reflect.Type) bool {
	_, ok := nullableValType(t)
	return ok
}

// ParseStructToParams parses a struct's fields into a map of ParamDef using
// json tags for field names and @meta tags for descriptions and constraints.
func ParseStructToParams(v any) map[string]ParamDef {
	if v == nil {
		return nil
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	return parseStructType(t)
}

func parseStructType(t reflect.Type) map[string]ParamDef {
	params := make(map[string]ParamDef)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Handle embedded (anonymous) structs: flatten their fields.
		// But skip nullable types — they should not be flattened.
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if field.Anonymous && ft.Kind() == reflect.Struct && ft != reflect.TypeOf(time.Time{}) && !isNullableType(ft) {
			for k, v := range parseStructType(ft) {
				params[k] = v
			}
			continue
		}

		// Determine field name from json tag, fallback to field name.
		name := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] == "-" {
				continue
			}
			if parts[0] != "" {
				name = parts[0]
			}
		}

		desc, required, ignore := parseMetaTag(field.Tag.Get("@meta"))
		if ignore {
			continue
		}
		schemaType := goTypeToSchemaType(ft)
		schemaFormat := goTypeToSchemaFormat(ft)

		pd := ParamDef{
			Type:        schemaType,
			Format:      schemaFormat,
			Description: desc,
			Required:    required,
		}

		if schemaType == "object" && ft.Kind() == reflect.Struct {
			pd.Properties = parseStructType(ft)
		}
		if schemaType == "array" {
			elemType := ft.Elem()
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}
			items := &ParamDef{Type: goTypeToSchemaType(elemType)}
			if elemType.Kind() == reflect.Struct && elemType != reflect.TypeOf(time.Time{}) && !isNullableType(elemType) {
				items.Properties = parseStructType(elemType)
			}
			pd.Items = items
		}

		params[name] = pd
	}
	return params
}

// StructToSchema converts a struct instance to a JSON-Schema-compatible map.
func StructToSchema(v any) map[string]any {
	params := ParseStructToParams(v)
	if params == nil {
		return nil
	}
	return ParamSchema(ParamDef{Type: "object", Properties: params})
}

// mergeParams merges auto-parsed params with manually specified params.
// Manual params take precedence over auto-parsed ones.
func mergeParams(auto, manual map[string]ParamDef) map[string]ParamDef {
	if len(auto) == 0 {
		return manual
	}
	if len(manual) == 0 {
		return auto
	}
	merged := make(map[string]ParamDef, len(auto)+len(manual))
	for k, v := range auto {
		merged[k] = v
	}
	for k, v := range manual {
		merged[k] = v
	}
	return merged
}

// ResolvedParams returns the effective parameters for a RouteDef by merging
// auto-parsed Request struct fields with manually specified Params.
func ResolvedParams(def RouteDef) map[string]ParamDef {
	auto := ParseStructToParams(def.Request)
	return mergeParams(auto, def.Params)
}

// FileParamNames returns the set of parameter names whose Format is "binary".
func FileParamNames(params map[string]ParamDef) map[string]bool {
	result := make(map[string]bool)
	for name, p := range params {
		if p.Format == "binary" {
			result[name] = true
		}
	}
	return result
}

// ExtractPathParams returns parameter names from a Gin-style route path.
func ExtractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, part[1:])
		}
	}
	return params
}
