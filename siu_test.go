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

package siu_test

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-zookeeper/zk"
	"github.com/stella-go/siu"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/meta"
	"github.com/stella-go/siu/t/n"
)

// ===========================================================================
// Test fixtures: IoC / middleware / router stubs used by TestRun
// ===========================================================================

type testMiddleware struct {
	Conn *zk.Conn `@siu:""`
}

func (p *testMiddleware) Init() {
	fmt.Printf("siu test init\n")
	siu.Cron("*/5 * * * * ?", func() {
		fmt.Println("cron", time.Now())
	})
	siu.Cron("*/5 * * * * ?", func() {
		panic(fmt.Errorf("cron panic"))
	})
}

func (p *testMiddleware) Condition() bool { return true }
func (p *testMiddleware) Function() gin.HandlerFunc {
	return func(c *gin.Context) {}
}
func (p *testMiddleware) Order() int { return 0 }

type testCipher struct{}

func (*testCipher) Decrypt(enc string) (string, error) {
	return strings.ReplaceAll(enc, "x", ""), nil
}

var testEnv = &config.DecryptEnvironment{Cipher: &testCipher{}}

type testRegister1 struct{}

func (*testRegister1) Named() map[string]any {
	return map[string]any{"environment": testEnv}
}
func (*testRegister1) Typed() map[reflect.Type]any {
	return map[reflect.Type]any{
		reflect.TypeOf((*config.TypedConfig)(nil)).Elem(): testEnv,
	}
}
func (*testRegister1) Order() int { return -1 }

type testRegister2 struct {
	Server *gin.Engine `@siu:""`
}

type testAfterServer struct {
	Server *gin.Engine `@siu:""`
}

func (p *testAfterServer) Init() {
	fmt.Println("Initialized L")
	p.Server.GET("/after", func(ctx *gin.Context) {})
}

var afterServer = &testAfterServer{}

func (*testRegister2) Named() map[string]any {
	return map[string]any{"after": afterServer}
}
func (*testRegister2) Typed() map[reflect.Type]any { return map[reflect.Type]any{} }
func (*testRegister2) Order() int                  { return 1 }

type testRouter struct{}

func (*testRouter) Router() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GET /hi": func(ctx *gin.Context) {
			time.Sleep(5 * time.Millisecond)
			ctx.String(200, "hello")
		},
	}
}

// ===========================================================================
// Integration test: full server lifecycle
// ===========================================================================

func TestRun(t *testing.T) {
	go func() {
		listener, _ := net.Listen("tcp", "127.0.0.1:2181")
		listener.Accept()
	}()
	go func() {
		time.Sleep(15 * time.Second)
		http.Get("http://localhost:8080/hi")
		http.Get("http://localhost:8080/abc")
		siu.INFO("__sLINE__ %s", "hello")
		syscall.Kill(os.Getpid(), 15)
	}()
	os.Setenv("STELLA_SERVER_MODE", "debug")
	os.Setenv("STELLA_LOGGER_SIU", "true")
	os.Setenv("STELLA_LOGGER_LEVEL", "debug")
	os.Setenv("STELLA_LOGGER_TAG", "[SIU]")
	os.Setenv("STELLA_LOGGER_PATTERN", "%d{2006-01-02T15:04:05} %c %p [%g] - %l{7} %m")
	os.Setenv("STELLA_LOGGER_SYSLOG", "127.0.0.1:514")
	os.Setenv("STELLA_ZOOKEEPER", "zookeeperxxx")
	os.Setenv("STELLA_ZOOKEEPER_SERVERS", "127.x0x.0.1:x21x81")
	os.Setenv("STELLA_MIDDLEWARE_CORS_DISABLE", "true")
	siu.Register(&testRegister2{}, &testRegister1{})
	siu.Use(&testMiddleware{})
	siu.Route(&testRouter{})
	siu.Run()
}

// ===========================================================================
// Decorator tests: siu.Meta
// ===========================================================================

func TestDecorators(t *testing.T) {
	dummyHandler := func(c *gin.Context) { c.String(200, "ok") }

	t.Run("Meta returns same handler", func(t *testing.T) {
		h := siu.Meta(dummyHandler, siu.RouteDef{
			Summary: "test",
			Params:  map[string]siu.ParamDef{"q": {Type: "string"}},
		})
		if reflect.ValueOf(h).Pointer() != reflect.ValueOf(dummyHandler).Pointer() {
			t.Error("Meta should return the same handler")
		}
	})
}

// NOTE: Swagger/MCP internal tests were moved to meta/meta_test.go

type CreateUserRequest struct {
	Name  string `json:"name" @meta:"desc=User name,required"`
	Email string `json:"email" @meta:"desc=Email address"`
	Age   int    `json:"age" @meta:"desc=User age,required=true"`
}

type Address struct {
	City   string `json:"city" @meta:"desc=City name"`
	Street string `json:"street" @meta:"desc=Street address"`
}

type OrderRequest struct {
	ItemID  string  `json:"item_id" @meta:"desc=Item identifier,required"`
	Qty     int     `json:"qty" @meta:"desc=Quantity"`
	Address Address `json:"address" @meta:"desc=Shipping address"`
}

type ListResponse struct {
	Total int      `json:"total" @meta:"desc=Total count"`
	Names []string `json:"names" @meta:"desc=Name list"`
}

type EmbeddedBase struct {
	ID string `json:"id" @meta:"desc=Resource ID,required"`
}

type ExtendedRequest struct {
	EmbeddedBase
	Title string `json:"title" @meta:"desc=Title"`
}

type NullableRequest struct {
	Name    n.String  `json:"name" @meta:"desc=User name,required"`
	Age     n.Int     `json:"age" @meta:"desc=User age"`
	Score   n.Float64 `json:"score" @meta:"desc=Score"`
	Active  n.Bool    `json:"active" @meta:"desc=Is active"`
	Created n.Time    `json:"created" @meta:"desc=Created time"`
}

type NullableArrayRequest struct {
	Tags []n.String `json:"tags" @meta:"desc=Tag list"`
}

func TestParseStructToParams(t *testing.T) {
	t.Run("basic struct fields", func(t *testing.T) {
		params := meta.ParseStructToParams(CreateUserRequest{})
		if len(params) != 3 {
			t.Fatalf("Expected 3 params, got %d", len(params))
		}
		name := params["name"]
		if name.Type != "string" || name.Description != "User name" || !name.Required {
			t.Errorf("name param wrong: %+v", name)
		}
		email := params["email"]
		if email.Type != "string" || email.Description != "Email address" || email.Required {
			t.Errorf("email param wrong: %+v", email)
		}
		age := params["age"]
		if age.Type != "integer" || age.Description != "User age" || !age.Required {
			t.Errorf("age param wrong: %+v", age)
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		params := meta.ParseStructToParams(&CreateUserRequest{})
		if len(params) != 3 {
			t.Fatalf("Expected 3 params from pointer, got %d", len(params))
		}
	})

	t.Run("nested struct", func(t *testing.T) {
		params := meta.ParseStructToParams(OrderRequest{})
		addr := params["address"]
		if addr.Type != "object" || addr.Description != "Shipping address" {
			t.Errorf("address param wrong: %+v", addr)
		}
		if len(addr.Properties) != 2 {
			t.Errorf("Expected 2 address properties, got %d", len(addr.Properties))
		}
		if addr.Properties["city"].Description != "City name" {
			t.Errorf("city desc wrong: %s", addr.Properties["city"].Description)
		}
	})

	t.Run("array field", func(t *testing.T) {
		params := meta.ParseStructToParams(ListResponse{})
		names := params["names"]
		if names.Type != "array" || names.Items == nil || names.Items.Type != "string" {
			t.Errorf("names param wrong: %+v", names)
		}
	})

	t.Run("embedded struct flattened", func(t *testing.T) {
		params := meta.ParseStructToParams(ExtendedRequest{})
		if len(params) != 2 {
			t.Fatalf("Expected 2 params (flattened), got %d", len(params))
		}
		id := params["id"]
		if id.Type != "string" || id.Description != "Resource ID" || !id.Required {
			t.Errorf("id param wrong: %+v", id)
		}
		title := params["title"]
		if title.Description != "Title" {
			t.Errorf("title param wrong: %+v", title)
		}
	})

	t.Run("nil returns nil", func(t *testing.T) {
		if meta.ParseStructToParams(nil) != nil {
			t.Error("Expected nil for nil input")
		}
	})

	t.Run("non-struct returns nil", func(t *testing.T) {
		if meta.ParseStructToParams("hello") != nil {
			t.Error("Expected nil for non-struct input")
		}
	})

	t.Run("nullable types map to primitive types", func(t *testing.T) {
		params := meta.ParseStructToParams(NullableRequest{})
		if len(params) != 5 {
			t.Fatalf("Expected 5 params, got %d", len(params))
		}
		if params["name"].Type != "string" {
			t.Errorf("n.String should be 'string', got %q", params["name"].Type)
		}
		if !params["name"].Required {
			t.Error("name should be required")
		}
		if params["name"].Description != "User name" {
			t.Errorf("name desc = %q", params["name"].Description)
		}
		if params["age"].Type != "integer" {
			t.Errorf("n.Int should be 'integer', got %q", params["age"].Type)
		}
		if params["score"].Type != "number" {
			t.Errorf("n.Float64 should be 'number', got %q", params["score"].Type)
		}
		if params["active"].Type != "boolean" {
			t.Errorf("n.Bool should be 'boolean', got %q", params["active"].Type)
		}
		if params["created"].Type != "string" {
			t.Errorf("n.Time should be 'string', got %q", params["created"].Type)
		}
		// Nullable types should NOT have Properties (they are not objects)
		for _, name := range []string{"name", "age", "score", "active", "created"} {
			if params[name].Properties != nil {
				t.Errorf("%s should not have Properties", name)
			}
		}
	})

	t.Run("array of nullable types", func(t *testing.T) {
		params := meta.ParseStructToParams(NullableArrayRequest{})
		tags := params["tags"]
		if tags.Type != "array" {
			t.Fatalf("Expected array type, got %q", tags.Type)
		}
		if tags.Items == nil || tags.Items.Type != "string" {
			t.Errorf("[]n.String items should be 'string', got %+v", tags.Items)
		}
		if tags.Items.Properties != nil {
			t.Error("n.String array items should not have Properties")
		}
	})
}

func TestStructToSchema(t *testing.T) {
	schema := meta.StructToSchema(CreateUserRequest{})
	if schema["type"] != "object" {
		t.Errorf("Expected object type, got %v", schema["type"])
	}
	props, _ := schema["properties"].(map[string]any)
	if len(props) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(props))
	}
	nameProp, _ := props["name"].(map[string]any)
	if nameProp["description"] != "User name" {
		t.Errorf("name description wrong: %v", nameProp["description"])
	}
}

// ===========================================================================
// Format=binary ParamDef tests
// ===========================================================================

func TestBinaryParamDef(t *testing.T) {
	t.Run("FileParamNames identifies binary params", func(t *testing.T) {
		params := map[string]meta.ParamDef{
			"title": {Type: "string", Description: "File title", Required: true},
			"file":  {Type: "string", Format: "binary", Description: "Upload file", Required: true},
		}
		fileParams := meta.FileParamNames(params)
		if !fileParams["file"] {
			t.Error("Expected 'file' in FileParamNames")
		}
		if fileParams["title"] {
			t.Error("'title' should not be in FileParamNames")
		}
	})

	t.Run("ParamSchema includes format=binary", func(t *testing.T) {
		p := meta.ParamDef{Type: "string", Format: "binary", Description: "Upload file"}
		schema := meta.ParamSchema(p)
		if schema["format"] != "binary" {
			t.Errorf("Expected format=binary, got %v", schema["format"])
		}
	})
}
