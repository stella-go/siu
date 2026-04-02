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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/config"
	"github.com/stella-go/siu/t/n"
)

// ===========================================================================
// Test helper types (redefined here since meta_test is a separate compilation)
// ===========================================================================

type createUserRequest struct {
	Name  string `json:"name" @meta:"desc=User name,required"`
	Email string `json:"email" @meta:"desc=Email address"`
	Age   int    `json:"age" @meta:"desc=User age,required=true"`
}

type listResponse struct {
	Total int      `json:"total" @meta:"desc=Total count"`
	Names []string `json:"names" @meta:"desc=Name list"`
}

type byteFieldRequest struct {
	Name string `json:"name" @meta:"desc=Name,required"`
	Data []byte `json:"data" @meta:"desc=Binary data"`
}

type nullableRequest struct {
	Name    n.String  `json:"name" @meta:"desc=User name,required"`
	Age     n.Int     `json:"age" @meta:"desc=User age"`
	Score   n.Float64 `json:"score" @meta:"desc=Score"`
	Active  n.Bool    `json:"active" @meta:"desc=Is active"`
	Created n.Time    `json:"created" @meta:"desc=Created time"`
}

// ===========================================================================
// Test helpers
// ===========================================================================

func resetGlobals() {
	globalSwaggerPaths = make(map[string]map[string]any)
	globalMCPToolMap = make(map[string]MCPToolDef)
	globalMCPToolList = nil
}

// registerSwagger is a test helper that builds swagger spec and registers routes.
func registerSwagger(server *gin.Engine, basePath string, paths map[string]map[string]any, env config.TypedConfig) {
	m := &MiddlewareSwagger{
		Conf:     env,
		basePath: basePath,
	}
	m.title = "API Documentation"
	m.description = ""
	m.version = "1.0.0"
	if env != nil {
		m.title = env.GetStringOr("swagger.title", "API Documentation")
		m.description = env.GetStringOr("swagger.description", "")
		m.version = env.GetStringOr("swagger.version", "1.0.0")
	}
	globalSwaggerPaths = paths
	m.buildSpec()

	swaggerGroup := server.Group(basePath)
	swaggerGroup.GET("/doc.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", m.specJSON)
	})
	swaggerGroup.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(m.html))
	})
}

// registerMCP is a test helper that registers MCP JSON-RPC endpoint.
func registerMCP(server *gin.Engine, basePath string, tools []MCPToolDef, cfg *MCPConfig) {
	m := &MiddlewareMCP{
		Server:   server,
		basePath: basePath,
	}
	if cfg == nil {
		cfg = &MCPConfig{}
	}
	if cfg.ServerName == "" {
		cfg.ServerName = "siu-mcp-server"
	}
	if cfg.ServerVersion == "" {
		cfg.ServerVersion = "1.0.0"
	}
	m.cfg = cfg

	toolMap := make(map[string]MCPToolDef)
	var toolList []map[string]any
	for _, t := range tools {
		name, _ := t.Definition["name"].(string)
		if name != "" {
			toolMap[name] = t
			toolList = append(toolList, t.Definition)
		}
	}
	globalMCPToolMap = toolMap
	globalMCPToolList = toolList

	server.POST(basePath, func(c *gin.Context) {
		m.handleRequest(c)
	})
}

// mcpCall sends a JSON-RPC request to the MCP endpoint.
func mcpCall(engine *gin.Engine, method string, id int, params map[string]any) map[string]any {
	return mcpCallWithHeaders(engine, method, id, params, nil)
}

// mcpCallWithHeaders sends a JSON-RPC request with custom HTTP headers.
func mcpCallWithHeaders(engine *gin.Engine, method string, id int, params map[string]any, headers map[string]string) map[string]any {
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		payload["params"] = params
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	engine.ServeHTTP(w, req)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

// ===========================================================================
// Auto-generation tests: generateSwaggerMeta / generateMCPMeta
// ===========================================================================

func TestGenerateSwaggerMeta(t *testing.T) {
	t.Run("GET with path params", func(t *testing.T) {
		meta := generateSwaggerMeta("GET", "/users/:id")
		if meta["summary"] != "GET /users/{id}" {
			t.Errorf("Expected summary 'GET /users/{id}', got %v", meta["summary"])
		}
		params, ok := meta["parameters"].([]map[string]any)
		if !ok || len(params) != 1 || params[0]["name"] != "id" {
			t.Error("Expected path param 'id'")
		}
		if meta["requestBody"] != nil {
			t.Error("GET should not have requestBody")
		}
	})

	t.Run("POST without path params", func(t *testing.T) {
		meta := generateSwaggerMeta("POST", "/users")
		if meta["requestBody"] == nil {
			t.Error("POST should have requestBody")
		}
		params, _ := meta["parameters"].([]map[string]any)
		if len(params) != 0 {
			t.Error("Expected no path params for /users")
		}
	})
}

func TestGenerateMCPMeta(t *testing.T) {
	t.Run("tool name and inputSchema from path", func(t *testing.T) {
		meta := generateMCPMeta("GET", "/users/:id")
		name, _ := meta["name"].(string)
		if name != "get_users_id" {
			t.Errorf("Expected tool name 'get_users_id', got '%s'", name)
		}
		schema, _ := meta["inputSchema"].(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		if _, ok := props["id"]; !ok {
			t.Error("Expected property 'id' in inputSchema")
		}
		required, _ := schema["required"].([]string)
		if len(required) != 1 || required[0] != "id" {
			t.Error("Expected 'id' in required")
		}
	})
}

// ===========================================================================
// Swagger endpoint tests
// ===========================================================================

func TestSwaggerEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	paths := map[string]map[string]any{
		"/users": {
			"get": map[string]any{
				"summary":   "Get users",
				"responses": map[string]any{"200": map[string]any{"description": "OK"}},
			},
		},
		"/users/:id": {
			"get": map[string]any{"summary": "Get user by ID"},
		},
	}
	registerSwagger(engine, "/swagger", paths, &config.ConfigurationEnvironment{})

	t.Run("doc.json returns OpenAPI 3.0 spec", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/swagger/doc.json", nil)
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("Expected 200, got %d", w.Code)
		}
		var spec map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
			t.Fatalf("Failed to parse spec: %v", err)
		}
		if spec["openapi"] != "3.0.0" {
			t.Error("Expected openapi 3.0.0")
		}
		specPaths, _ := spec["paths"].(map[string]any)
		if _, ok := specPaths["/users/{id}"]; !ok {
			t.Error("Expected path /users/{id} with OpenAPI param format")
		}

		// Verify securitySchemes and global security
		components, _ := spec["components"].(map[string]any)
		if components == nil {
			t.Fatal("Expected components in spec")
		}
		schemes, _ := components["securitySchemes"].(map[string]any)
		authScheme, _ := schemes["Authorization"].(map[string]any)
		if authScheme["type"] != "apiKey" || authScheme["in"] != "header" || authScheme["name"] != "Authorization" {
			t.Errorf("Unexpected securityScheme: %v", authScheme)
		}
		security, _ := spec["security"].([]any)
		if len(security) != 1 {
			t.Errorf("Expected 1 global security entry, got %d", len(security))
		}
	})

	t.Run("UI page contains swagger-ui and spec URL", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/swagger/", nil)
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("Expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "swagger-ui") {
			t.Error("Expected swagger-ui content")
		}
		if !strings.Contains(body, "/swagger/doc.json") {
			t.Error("Expected spec URL in UI page")
		}
	})
}

func TestSwaggerEndpointAutoGenerated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	paths := map[string]map[string]any{
		"/items/:id": {
			"get":    generateSwaggerMeta("GET", "/items/:id"),
			"delete": generateSwaggerMeta("DELETE", "/items/:id"),
		},
		"/items": {
			"post": generateSwaggerMeta("POST", "/items"),
		},
	}
	registerSwagger(engine, "/swagger", paths, &config.ConfigurationEnvironment{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/swagger/doc.json", nil)
	engine.ServeHTTP(w, req)

	var spec map[string]any
	json.Unmarshal(w.Body.Bytes(), &spec)
	specPaths, _ := spec["paths"].(map[string]any)

	t.Run("path param converted to OpenAPI format", func(t *testing.T) {
		itemPath, ok := specPaths["/items/{id}"]
		if !ok {
			t.Fatal("Expected /items/{id} in spec paths")
		}
		getMeta, _ := itemPath.(map[string]any)["get"].(map[string]any)
		if getMeta["summary"] != "GET /items/{id}" {
			t.Errorf("Expected auto-generated summary, got %v", getMeta["summary"])
		}
	})

	t.Run("POST has auto-generated requestBody", func(t *testing.T) {
		postPath, _ := specPaths["/items"].(map[string]any)["post"].(map[string]any)
		if postPath["requestBody"] == nil {
			t.Error("Expected auto-generated requestBody for POST")
		}
	})
}

// ===========================================================================
// MCP endpoint tests
// ===========================================================================

func TestMCPEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/users", func(c *gin.Context) {
		c.JSON(200, map[string]any{"users": []string{"alice", "bob"}})
	})

	tools := []MCPToolDef{
		{
			Definition: map[string]any{
				"name":        "get_users",
				"description": "Get user list",
				"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
			},
			Method: "GET",
			Path:   "/users",
		},
	}
	registerMCP(engine, "/mcp", tools, nil)

	t.Run("initialize returns protocol version", func(t *testing.T) {
		resp := mcpCall(engine, "initialize", 1, nil)
		result, _ := resp["result"].(map[string]any)
		if result["protocolVersion"] != "2025-03-26" {
			t.Error("Expected protocol version 2025-03-26")
		}
	})

	t.Run("tools/list returns registered tools", func(t *testing.T) {
		resp := mcpCall(engine, "tools/list", 2, nil)
		result, _ := resp["result"].(map[string]any)
		toolsList, _ := result["tools"].([]any)
		if len(toolsList) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(toolsList))
		}
	})

	t.Run("tools/call invokes handler and returns result", func(t *testing.T) {
		resp := mcpCall(engine, "tools/call", 3, map[string]any{
			"name": "get_users", "arguments": map[string]any{},
		})
		result, _ := resp["result"].(map[string]any)
		content, _ := result["content"].([]any)
		if len(content) != 1 {
			t.Fatalf("Expected 1 content item, got %d", len(content))
		}
		text, _ := content[0].(map[string]any)["text"].(string)
		if !strings.Contains(text, "alice") {
			t.Errorf("Expected response containing 'alice', got: %s", text)
		}
	})

	t.Run("tools/call with unknown tool returns error", func(t *testing.T) {
		resp := mcpCall(engine, "tools/call", 4, map[string]any{
			"name": "not_exist", "arguments": map[string]any{},
		})
		if resp["error"] == nil {
			t.Error("Expected error for non-existent tool")
		}
	})

	t.Run("unknown method returns error", func(t *testing.T) {
		resp := mcpCall(engine, "unknown/method", 5, nil)
		if resp["error"] == nil {
			t.Error("Expected error for unknown method")
		}
	})
}

func TestMCPEndpointAutoGenerated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/items", func(c *gin.Context) {
		c.JSON(200, map[string]any{"items": []string{"a", "b"}})
	})

	def := generateMCPMeta("GET", "/items")
	tools := []MCPToolDef{
		{Definition: def, Method: "GET", Path: "/items"},
	}
	registerMCP(engine, "/mcp", tools, nil)

	t.Run("tools/list returns auto-generated tool name", func(t *testing.T) {
		resp := mcpCall(engine, "tools/list", 1, nil)
		result, _ := resp["result"].(map[string]any)
		toolsList, _ := result["tools"].([]any)
		if len(toolsList) != 1 {
			t.Fatalf("Expected 1 tool, got %d", len(toolsList))
		}
		tool := toolsList[0].(map[string]any)
		if tool["name"] != "get_items" {
			t.Errorf("Expected auto-generated name 'get_items', got '%v'", tool["name"])
		}
	})

	t.Run("tools/call works with auto-generated tool", func(t *testing.T) {
		resp := mcpCall(engine, "tools/call", 2, map[string]any{
			"name": "get_items", "arguments": map[string]any{},
		})
		result, _ := resp["result"].(map[string]any)
		content, _ := result["content"].([]any)
		text, _ := content[0].(map[string]any)["text"].(string)
		if !strings.Contains(text, "a") {
			t.Errorf("Expected auto-generated tool to return data, got: %s", text)
		}
	})
}

func TestMCPAuthHeaderPassThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/echo-auth", func(c *gin.Context) {
		c.JSON(200, map[string]any{
			"authorization": c.GetHeader("Authorization"),
			"cookie":        c.GetHeader("Cookie"),
		})
	})

	tools := []MCPToolDef{
		{
			Definition: map[string]any{
				"name":        "echo_auth",
				"description": "Echo auth headers",
				"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
			},
			Method: "GET",
			Path:   "/echo-auth",
		},
	}
	registerMCP(engine, "/mcp", tools, nil)

	t.Run("Authorization header is forwarded to internal request", func(t *testing.T) {
		resp := mcpCallWithHeaders(engine, "tools/call", 1, map[string]any{
			"name": "echo_auth", "arguments": map[string]any{},
		}, map[string]string{"Authorization": "Bearer test-jwt-token"})

		result, _ := resp["result"].(map[string]any)
		content, _ := result["content"].([]any)
		if len(content) == 0 {
			t.Fatal("Expected content in response")
		}
		text, _ := content[0].(map[string]any)["text"].(string)
		if !strings.Contains(text, "Bearer test-jwt-token") {
			t.Errorf("Expected Authorization header forwarded, got: %s", text)
		}
	})

	t.Run("Cookie header is forwarded to internal request", func(t *testing.T) {
		resp := mcpCallWithHeaders(engine, "tools/call", 2, map[string]any{
			"name": "echo_auth", "arguments": map[string]any{},
		}, map[string]string{"Cookie": "session_id=abc123"})

		result, _ := resp["result"].(map[string]any)
		content, _ := result["content"].([]any)
		text, _ := content[0].(map[string]any)["text"].(string)
		if !strings.Contains(text, "session_id=abc123") {
			t.Errorf("Expected Cookie header forwarded, got: %s", text)
		}
	})

	t.Run("both headers forwarded together", func(t *testing.T) {
		resp := mcpCallWithHeaders(engine, "tools/call", 3, map[string]any{
			"name": "echo_auth", "arguments": map[string]any{},
		}, map[string]string{
			"Authorization": "Bearer multi-token",
			"Cookie":        "sid=xyz",
		})

		result, _ := resp["result"].(map[string]any)
		content, _ := result["content"].([]any)
		text, _ := content[0].(map[string]any)["text"].(string)
		if !strings.Contains(text, "Bearer multi-token") {
			t.Errorf("Expected Authorization forwarded, got: %s", text)
		}
		if !strings.Contains(text, "sid=xyz") {
			t.Errorf("Expected Cookie forwarded, got: %s", text)
		}
	})

	t.Run("no headers when none provided", func(t *testing.T) {
		resp := mcpCall(engine, "tools/call", 4, map[string]any{
			"name": "echo_auth", "arguments": map[string]any{},
		})

		result, _ := resp["result"].(map[string]any)
		content, _ := result["content"].([]any)
		text, _ := content[0].(map[string]any)["text"].(string)
		var echoed map[string]any
		json.Unmarshal([]byte(text), &echoed)
		if echoed["authorization"] != "" {
			t.Errorf("Expected empty Authorization, got: %v", echoed["authorization"])
		}
		if echoed["cookie"] != "" {
			t.Errorf("Expected empty Cookie, got: %v", echoed["cookie"])
		}
	})
}

// ===========================================================================
// expandSwaggerRouteDef / expandMCPRouteDef tests
// ===========================================================================

func TestExpandRouteDefSwagger(t *testing.T) {
	t.Run("GET with path and query params", func(t *testing.T) {
		def := RouteDef{
			Name:    "get_user",
			Summary: "Get user by ID",
			Params: map[string]ParamDef{
				"id":     {Type: "string", Description: "User ID", Required: true},
				"fields": {Type: "string", Description: "Fields to include"},
			},
		}
		swagger := expandSwaggerRouteDef("GET", "/users/:id", def)

		if swagger["summary"] != "Get user by ID" {
			t.Errorf("swagger summary = %v", swagger["summary"])
		}
		params, _ := swagger["parameters"].([]map[string]any)
		if len(params) != 2 {
			t.Fatalf("Expected 2 swagger params, got %d", len(params))
		}
		var pathParam, queryParam map[string]any
		for _, p := range params {
			if p["in"] == "path" {
				pathParam = p
			} else {
				queryParam = p
			}
		}
		if pathParam == nil || pathParam["name"] != "id" || pathParam["required"] != true {
			t.Errorf("path param wrong: %v", pathParam)
		}
		if queryParam == nil || queryParam["name"] != "fields" || queryParam["in"] != "query" {
			t.Errorf("query param wrong: %v", queryParam)
		}
		if swagger["requestBody"] != nil {
			t.Error("GET should not have requestBody")
		}
	})

	t.Run("POST with body params", func(t *testing.T) {
		def := RouteDef{
			Summary: "Create user",
			Params: map[string]ParamDef{
				"name":  {Type: "string", Description: "User name", Required: true},
				"email": {Type: "string", Description: "Email"},
			},
		}
		swagger := expandSwaggerRouteDef("POST", "/users", def)

		if swagger["parameters"] != nil {
			t.Error("POST body params should not become parameters")
		}
		rb, _ := swagger["requestBody"].(map[string]any)
		if rb == nil {
			t.Fatal("Expected requestBody for POST")
		}
		content, _ := rb["content"].(map[string]any)
		jsonContent, _ := content["application/json"].(map[string]any)
		schema, _ := jsonContent["schema"].(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		if len(props) != 2 {
			t.Errorf("Expected 2 body properties, got %d", len(props))
		}
		reqList, _ := schema["required"].([]string)
		if len(reqList) != 1 || reqList[0] != "name" {
			t.Errorf("Expected required=[name], got %v", reqList)
		}
	})
}

func TestExpandRouteDefMCP(t *testing.T) {
	t.Run("GET with path and query params", func(t *testing.T) {
		def := RouteDef{
			Name:    "get_user",
			Summary: "Get user by ID",
			Params: map[string]ParamDef{
				"id":     {Type: "string", Description: "User ID", Required: true},
				"fields": {Type: "string", Description: "Fields to include"},
			},
		}
		mcp := expandMCPRouteDef("GET", "/users/:id", def)

		if mcp["name"] != "get_user" {
			t.Errorf("mcp name = %v", mcp["name"])
		}
		if mcp["description"] != "Get user by ID" {
			t.Errorf("mcp description = %v", mcp["description"])
		}
		schema, _ := mcp["inputSchema"].(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		if len(props) != 2 {
			t.Errorf("Expected 2 mcp properties, got %d", len(props))
		}
	})

	t.Run("POST name auto-derived", func(t *testing.T) {
		def := RouteDef{
			Summary: "Create user",
			Params: map[string]ParamDef{
				"name":  {Type: "string", Description: "User name", Required: true},
				"email": {Type: "string", Description: "Email"},
			},
		}
		mcp := expandMCPRouteDef("POST", "/users", def)

		if mcp["name"] != "post_users" {
			t.Errorf("mcp name = %v, want post_users", mcp["name"])
		}
		if mcp["description"] != "Create user" {
			t.Errorf("mcp description = %v", mcp["description"])
		}
	})
}

func TestMetaIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.GET("/users/:id", func(c *gin.Context) {
		c.JSON(200, map[string]any{"id": c.Param("id"), "name": "alice"})
	})

	def := RouteDef{
		Name:    "get_user",
		Summary: "Get user by ID",
		Params: map[string]ParamDef{
			"id": {Type: "string", Description: "User ID", Required: true},
		},
	}
	mcpDef := expandMCPRouteDef("GET", "/users/:id", def)
	tools := []MCPToolDef{
		{Definition: mcpDef, Method: "GET", Path: "/users/:id"},
	}
	registerMCP(engine, "/mcp", tools, nil)

	t.Run("tools/list shows Meta-derived tool", func(t *testing.T) {
		resp := mcpCall(engine, "tools/list", 1, nil)
		result, _ := resp["result"].(map[string]any)
		toolsList, _ := result["tools"].([]any)
		if len(toolsList) != 1 {
			t.Fatalf("Expected 1 tool, got %d", len(toolsList))
		}
		tool := toolsList[0].(map[string]any)
		if tool["name"] != "get_user" {
			t.Errorf("tool name = %v, want get_user", tool["name"])
		}
		if tool["description"] != "Get user by ID" {
			t.Errorf("tool description = %v", tool["description"])
		}
	})

	t.Run("tools/call works with Meta-derived tool", func(t *testing.T) {
		resp := mcpCall(engine, "tools/call", 2, map[string]any{
			"name":      "get_user",
			"arguments": map[string]any{"id": "42"},
		})
		result, _ := resp["result"].(map[string]any)
		content, _ := result["content"].([]any)
		text, _ := content[0].(map[string]any)["text"].(string)
		if !strings.Contains(text, "alice") || !strings.Contains(text, "42") {
			t.Errorf("Expected response with alice and 42, got: %s", text)
		}
	})
}

// ===========================================================================
// ExpandRouteDefWithRequest tests
// ===========================================================================

func TestExpandRouteDefWithRequest(t *testing.T) {
	t.Run("Swagger POST with Request struct", func(t *testing.T) {
		def := RouteDef{
			Summary: "Create user",
			Request: createUserRequest{},
		}
		swagger := expandSwaggerRouteDef("POST", "/users", def)
		rb, _ := swagger["requestBody"].(map[string]any)
		if rb == nil {
			t.Fatal("Expected requestBody")
		}
		content, _ := rb["content"].(map[string]any)
		jsonContent, _ := content["application/json"].(map[string]any)
		schema, _ := jsonContent["schema"].(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		if len(props) != 3 {
			t.Errorf("Expected 3 body properties, got %d", len(props))
		}
		nameProp, _ := props["name"].(map[string]any)
		if nameProp["description"] != "User name" {
			t.Errorf("name description = %v", nameProp["description"])
		}
		reqList, _ := schema["required"].([]string)
		if len(reqList) != 2 {
			t.Errorf("Expected 2 required fields (name, age), got %v", reqList)
		}
	})

	t.Run("Swagger POST with Request + manual Params override", func(t *testing.T) {
		def := RouteDef{
			Summary: "Create user",
			Request: createUserRequest{},
			Params: map[string]ParamDef{
				"name": {Type: "string", Description: "Override name", Required: false},
			},
		}
		swagger := expandSwaggerRouteDef("POST", "/users", def)
		rb, _ := swagger["requestBody"].(map[string]any)
		content, _ := rb["content"].(map[string]any)
		jsonContent, _ := content["application/json"].(map[string]any)
		schema, _ := jsonContent["schema"].(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		nameProp, _ := props["name"].(map[string]any)
		if nameProp["description"] != "Override name" {
			t.Errorf("manual Params should override: got %v", nameProp["description"])
		}
	})

	t.Run("Swagger with Response schema", func(t *testing.T) {
		def := RouteDef{
			Summary:  "List users",
			Response: listResponse{},
		}
		swagger := expandSwaggerRouteDef("GET", "/users", def)
		responses, _ := swagger["responses"].(map[string]any)
		ok200, _ := responses["200"].(map[string]any)
		content, _ := ok200["content"].(map[string]any)
		if content == nil {
			t.Fatal("Expected response content with schema")
		}
		jsonContent, _ := content["application/json"].(map[string]any)
		schema, _ := jsonContent["schema"].(map[string]any)
		if schema["type"] != "object" {
			t.Errorf("Expected response schema type=object, got %v", schema["type"])
		}
	})

	t.Run("MCP with Request struct", func(t *testing.T) {
		def := RouteDef{
			Summary: "Create user",
			Request: createUserRequest{},
		}
		mcp := expandMCPRouteDef("POST", "/users", def)
		schema, _ := mcp["inputSchema"].(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		if len(props) != 3 {
			t.Errorf("Expected 3 mcp properties, got %d", len(props))
		}
		nameProp, _ := props["name"].(map[string]any)
		if nameProp["description"] != "User name" {
			t.Errorf("name description = %v", nameProp["description"])
		}
	})
}

// ===========================================================================
// []byte (base64) type mapping tests
// ===========================================================================

func TestByteSliceParamSchema(t *testing.T) {
	t.Run("[]byte maps to string with format byte", func(t *testing.T) {
		params := ParseStructToParams(byteFieldRequest{})
		dp := params["data"]
		if dp.Type != "string" {
			t.Errorf("Expected type 'string', got %q", dp.Type)
		}
		if dp.Format != "byte" {
			t.Errorf("Expected format 'byte', got %q", dp.Format)
		}
		if dp.Description != "Binary data" {
			t.Errorf("Expected description 'Binary data', got %q", dp.Description)
		}

		schema := ParamSchema(dp)
		if schema["format"] != "byte" {
			t.Errorf("Schema format should be 'byte', got %v", schema["format"])
		}
	})

	t.Run("[]byte in Swagger requestBody", func(t *testing.T) {
		def := RouteDef{
			Summary: "Upload data",
			Request: byteFieldRequest{},
		}
		swagger := expandSwaggerRouteDef("POST", "/upload", def)
		rb, _ := swagger["requestBody"].(map[string]any)
		content, _ := rb["content"].(map[string]any)
		jsonContent, _ := content["application/json"].(map[string]any)
		schema, _ := jsonContent["schema"].(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		dataProp, _ := props["data"].(map[string]any)
		if dataProp["type"] != "string" || dataProp["format"] != "byte" {
			t.Errorf("Expected type=string,format=byte for data, got %v", dataProp)
		}
	})

	t.Run("[]byte in MCP inputSchema", func(t *testing.T) {
		def := RouteDef{
			Summary: "Upload data",
			Request: byteFieldRequest{},
		}
		mcp := expandMCPRouteDef("POST", "/upload", def)
		schema, _ := mcp["inputSchema"].(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		dataProp, _ := props["data"].(map[string]any)
		if dataProp["type"] != "string" || dataProp["format"] != "byte" {
			t.Errorf("Expected type=string,format=byte, got %v", dataProp)
		}
	})
}

// ===========================================================================
// ContentType RouteDef + Swagger tests
// ===========================================================================

func TestContentTypeSwagger(t *testing.T) {
	t.Run("ContentType=multipart/form-data in Swagger", func(t *testing.T) {
		def := RouteDef{
			Summary:     "Upload file",
			ContentType: "multipart/form-data",
			Params: map[string]ParamDef{
				"title": {Type: "string", Description: "Title", Required: true},
				"file":  {Type: "string", Format: "binary", Description: "File", Required: true},
			},
		}
		swagger := expandSwaggerRouteDef("POST", "/upload", def)
		rb, _ := swagger["requestBody"].(map[string]any)
		content, _ := rb["content"].(map[string]any)
		if _, ok := content["multipart/form-data"]; !ok {
			t.Error("Expected multipart/form-data content type")
		}
		if _, ok := content["application/json"]; ok {
			t.Error("Should not have application/json content type for multipart")
		}
	})

	t.Run("Default ContentType generates application/json", func(t *testing.T) {
		def := RouteDef{
			Summary: "Create user",
			Request: createUserRequest{},
		}
		swagger := expandSwaggerRouteDef("POST", "/users", def)
		rb, _ := swagger["requestBody"].(map[string]any)
		content, _ := rb["content"].(map[string]any)
		if _, ok := content["application/json"]; !ok {
			t.Error("Expected application/json content type")
		}
	})

	t.Run("Custom ContentType in Swagger", func(t *testing.T) {
		def := RouteDef{
			Summary:     "Submit form",
			ContentType: "application/x-www-form-urlencoded",
			Params: map[string]ParamDef{
				"name": {Type: "string", Description: "Name"},
				"age":  {Type: "integer", Description: "Age"},
			},
		}
		swagger := expandSwaggerRouteDef("POST", "/form", def)
		rb, _ := swagger["requestBody"].(map[string]any)
		content, _ := rb["content"].(map[string]any)
		if _, ok := content["application/x-www-form-urlencoded"]; !ok {
			t.Error("Expected application/x-www-form-urlencoded content type")
		}
	})
}

// ===========================================================================
// MCP multipart tool call tests
// ===========================================================================

func TestMCPMultipartToolCall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.POST("/upload", func(c *gin.Context) {
		title := c.PostForm("title")
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(400, map[string]any{"error": err.Error()})
			return
		}
		defer file.Close()
		content := make([]byte, header.Size)
		file.Read(content)
		c.JSON(200, map[string]any{
			"title":    title,
			"filename": header.Filename,
			"size":     header.Size,
			"content":  string(content),
		})
	})

	def := RouteDef{
		Name:        "upload_file",
		Summary:     "Upload a file",
		ContentType: "multipart/form-data",
		Params: map[string]ParamDef{
			"title": {Type: "string", Description: "File title", Required: true},
			"file":  {Type: "string", Format: "binary", Description: "Upload file", Required: true},
		},
	}
	resolved := ResolvedParams(def)
	fileParams := FileParamNames(resolved)
	mcpDef := expandMCPRouteDef("POST", "/upload", def)
	tools := []MCPToolDef{
		{
			Definition:  mcpDef,
			Method:      "POST",
			Path:        "/upload",
			ContentType: "multipart/form-data",
			FileParams:  fileParams,
		},
	}
	registerMCP(engine, "/mcp", tools, nil)

	t.Run("multipart tool call uploads file via base64", func(t *testing.T) {
		fileContent := "hello world"
		b64 := base64.StdEncoding.EncodeToString([]byte(fileContent))
		resp := mcpCall(engine, "tools/call", 1, map[string]any{
			"name": "upload_file",
			"arguments": map[string]any{
				"title": "test-file",
				"file":  b64,
			},
		})
		result, _ := resp["result"].(map[string]any)
		content, _ := result["content"].([]any)
		if len(content) == 0 {
			t.Fatal("Expected content in response")
		}
		text, _ := content[0].(map[string]any)["text"].(string)
		if !strings.Contains(text, "test-file") {
			t.Errorf("Expected title in response, got: %s", text)
		}
		if !strings.Contains(text, "hello world") {
			t.Errorf("Expected file content in response, got: %s", text)
		}
		isError, _ := result["isError"].(bool)
		if isError {
			t.Errorf("Expected isError=false, response: %s", text)
		}
	})

	t.Run("multipart tool with form fields only", func(t *testing.T) {
		engine2 := gin.New()
		engine2.POST("/form", func(c *gin.Context) {
			name := c.PostForm("name")
			age := c.PostForm("age")
			c.JSON(200, map[string]any{"name": name, "age": age})
		})
		formTools := []MCPToolDef{
			{
				Definition: map[string]any{
					"name":        "submit_form",
					"description": "Submit form",
					"inputSchema": map[string]any{"type": "object"},
				},
				Method:      "POST",
				Path:        "/form",
				ContentType: "multipart/form-data",
			},
		}
		registerMCP(engine2, "/mcp", formTools, nil)
		resp := mcpCall(engine2, "tools/call", 1, map[string]any{
			"name": "submit_form",
			"arguments": map[string]any{
				"name": "alice",
				"age":  "30",
			},
		})
		result, _ := resp["result"].(map[string]any)
		content, _ := result["content"].([]any)
		text, _ := content[0].(map[string]any)["text"].(string)
		if !strings.Contains(text, "alice") || !strings.Contains(text, "30") {
			t.Errorf("Expected form fields in response, got: %s", text)
		}
	})
}

// ===========================================================================
// MCP form-urlencoded tool call tests
// ===========================================================================

func TestMCPFormURLEncodedToolCall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.POST("/form", func(c *gin.Context) {
		name := c.PostForm("name")
		age := c.PostForm("age")
		c.JSON(200, map[string]any{"name": name, "age": age})
	})

	tools := []MCPToolDef{
		{
			Definition: map[string]any{
				"name":        "submit_form",
				"description": "Submit form",
				"inputSchema": map[string]any{"type": "object"},
			},
			Method:      "POST",
			Path:        "/form",
			ContentType: "application/x-www-form-urlencoded",
		},
	}
	registerMCP(engine, "/mcp", tools, nil)

	t.Run("form-urlencoded tool call sends form fields", func(t *testing.T) {
		resp := mcpCall(engine, "tools/call", 1, map[string]any{
			"name": "submit_form",
			"arguments": map[string]any{
				"name": "bob",
				"age":  "25",
			},
		})
		result, _ := resp["result"].(map[string]any)
		content, _ := result["content"].([]any)
		text, _ := content[0].(map[string]any)["text"].(string)
		if !strings.Contains(text, "bob") || !strings.Contains(text, "25") {
			t.Errorf("Expected form fields in response, got: %s", text)
		}
	})
}

// ===========================================================================
// MCPToolDef construction with ContentType tests
// ===========================================================================

func TestMCPToolDefContentType(t *testing.T) {
	def := RouteDef{
		Name:        "upload",
		Summary:     "Upload",
		ContentType: "multipart/form-data",
		Params: map[string]ParamDef{
			"file": {Type: "string", Format: "binary", Description: "File", Required: true},
			"name": {Type: "string", Description: "Name"},
		},
	}
	tool := MCPToolDef{
		Definition:  expandMCPRouteDef("POST", "/upload", def),
		Method:      "POST",
		Path:        "/upload",
		ContentType: def.ContentType,
		FileParams:  FileParamNames(ResolvedParams(def)),
	}
	if tool.ContentType != "multipart/form-data" {
		t.Errorf("Expected ContentType=multipart/form-data, got %q", tool.ContentType)
	}
	if !tool.FileParams["file"] {
		t.Error("Expected 'file' in FileParams")
	}
	if tool.FileParams["name"] {
		t.Error("'name' should not be in FileParams")
	}
}

// ===========================================================================
// CollectSwaggerData / CollectMCPData tests
// ===========================================================================

func TestCollectSwaggerData(t *testing.T) {
	resetGlobals()

	def := &RouteDef{
		Summary: "Get user",
		Params: map[string]ParamDef{
			"id": {Type: "string", Description: "User ID", Required: true},
		},
	}
	CollectSwaggerData([]string{"GET"}, "/users/:id", def)

	if _, ok := globalSwaggerPaths["/users/:id"]; !ok {
		t.Fatal("Expected /users/:id in globalSwaggerPaths")
	}
	getMeta, _ := globalSwaggerPaths["/users/:id"]["get"].(map[string]any)
	if getMeta["summary"] != "Get user" {
		t.Errorf("Expected summary 'Get user', got %v", getMeta["summary"])
	}
}

func TestCollectSwaggerDataAutoGenerate(t *testing.T) {
	resetGlobals()

	CollectSwaggerData([]string{"POST"}, "/items", nil)

	if _, ok := globalSwaggerPaths["/items"]; !ok {
		t.Fatal("Expected /items in globalSwaggerPaths")
	}
	postMeta, _ := globalSwaggerPaths["/items"]["post"].(map[string]any)
	if postMeta["requestBody"] == nil {
		t.Error("Expected auto-generated requestBody for POST")
	}
}

func TestCollectMCPData(t *testing.T) {
	resetGlobals()

	def := &RouteDef{
		Name:    "get_user",
		Summary: "Get user by ID",
		Params: map[string]ParamDef{
			"id": {Type: "string", Description: "User ID", Required: true},
		},
	}
	CollectMCPData([]string{"GET"}, "/users/:id", def)

	if len(globalMCPToolMap) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(globalMCPToolMap))
	}
	tool, ok := globalMCPToolMap["get_user"]
	if !ok {
		t.Fatal("Expected tool 'get_user'")
	}
	if tool.Method != "GET" {
		t.Errorf("Expected method GET, got %s", tool.Method)
	}
}

func TestCollectMCPDataAutoGenerate(t *testing.T) {
	resetGlobals()

	CollectMCPData([]string{"GET", "DELETE"}, "/items/:id", nil)

	if len(globalMCPToolMap) != 2 {
		t.Fatalf("Expected 2 tools, got %d", len(globalMCPToolMap))
	}
	if _, ok := globalMCPToolMap["get_items_id"]; !ok {
		t.Error("Expected tool 'get_items_id'")
	}
	if _, ok := globalMCPToolMap["delete_items_id"]; !ok {
		t.Error("Expected tool 'delete_items_id'")
	}
}

// ===========================================================================
// Middleware lifecycle tests (Condition + Function)
// ===========================================================================

func TestMiddlewareSwaggerCondition(t *testing.T) {
	t.Run("disabled by default", func(t *testing.T) {
		m := &MiddlewareSwagger{
			Conf: &config.ConfigurationEnvironment{},
		}
		m.Init()
		if m.Condition() {
			t.Error("Expected Condition()=false when swagger.disable is not set (default true)")
		}
	})
}

func TestMiddlewareMCPCondition(t *testing.T) {
	t.Run("disabled by default", func(t *testing.T) {
		m := &MiddlewareMCP{
			Conf: &config.ConfigurationEnvironment{},
		}
		m.Init()
		if m.Condition() {
			t.Error("Expected Condition()=false when mcp.disable is not set (default true)")
		}
	})
}

func TestMiddlewareSwaggerFunction(t *testing.T) {
	resetGlobals()
	gin.SetMode(gin.TestMode)

	globalSwaggerPaths["/test"] = map[string]any{
		"get": map[string]any{"summary": "Test"},
	}

	m := &MiddlewareSwagger{
		Conf:     &config.ConfigurationEnvironment{},
		basePath: "/swagger",
		title:    "Test",
		version:  "1.0",
		once:     sync.Once{},
	}

	handler := m.Function()

	engine := gin.New()
	engine.Use(handler)
	engine.GET("/other", func(c *gin.Context) {
		c.String(200, "other")
	})

	// Swagger doc should be served
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/swagger/doc.json", nil)
	engine.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200 for doc.json, got %d", w.Code)
	}

	// Non-swagger paths should pass through
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/other", nil)
	engine.ServeHTTP(w2, req2)
	if w2.Code != 200 || w2.Body.String() != "other" {
		t.Errorf("Expected passthrough, got %d %s", w2.Code, w2.Body.String())
	}
}
