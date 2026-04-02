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
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/common"
	"github.com/stella-go/siu/config"
)

const (
	MCPDisableKey       = "mcp.disable"
	MCPPathKey          = "mcp.path"
	MCPInstructionsKey  = "mcp.instructions"
	MCPServerNameKey    = "mcp.server-name"
	MCPServerVersionKey = "mcp.server-version"
	MCPMiddleOrder      = 40
)

// MCPConfig holds optional settings for the MCP server endpoint.
type MCPConfig struct {
	ServerName    string // server name reported in initialize (default: "siu-mcp-server")
	ServerVersion string // server version reported in initialize (default: "1.0.0")
	Instructions  string // free-form instructions returned in initialize result
}

// MCPToolDef binds an MCP tool definition to the underlying HTTP route.
type MCPToolDef struct {
	Definition  map[string]any
	Method      string
	Path        string
	ContentType string          // request body content type (e.g. "multipart/form-data")
	FileParams  map[string]bool // parameter names that are file uploads (format=binary)
}

var (
	globalMCPToolMap  = make(map[string]MCPToolDef)
	globalMCPToolList []map[string]any
)

func CollectMCPData(methods []string, fullPath string, def *RouteDef) {
	var effective map[string]any
	var contentType string
	var fileParams map[string]bool
	if def != nil {
		effective = expandMCPRouteDef(strings.ToUpper(methods[0]), fullPath, *def)
		contentType = def.ContentType
		if contentType != "" {
			fileParams = FileParamNames(ResolvedParams(*def))
		}
	}

	if effective != nil {
		t := MCPToolDef{
			Definition:  effective,
			Method:      strings.ToUpper(methods[0]),
			Path:        fullPath,
			ContentType: contentType,
			FileParams:  fileParams,
		}
		name, _ := t.Definition["name"].(string)
		if name != "" {
			globalMCPToolMap[name] = t
			globalMCPToolList = append(globalMCPToolList, t.Definition)
		}
	} else {
		for _, method := range methods {
			d := generateMCPMeta(strings.ToUpper(method), fullPath)
			t := MCPToolDef{
				Definition: d,
				Method:     strings.ToUpper(method),
				Path:       fullPath,
			}
			name, _ := t.Definition["name"].(string)
			if name != "" {
				globalMCPToolMap[name] = t
				globalMCPToolList = append(globalMCPToolList, t.Definition)
			}
		}
	}
}

// --- MiddlewareMCP ---

type MiddlewareMCP struct {
	Server *gin.Engine        `@siu:"name='server',default='type'"`
	Conf   config.TypedConfig `@siu:"name='environment',default='type'"`

	basePath string
	cfg      *MCPConfig
}

func (p *MiddlewareMCP) Init() {
	serverPrefix := p.Conf.GetStringOr("server.prefix", "")
	p.basePath = path.Join(serverPrefix, p.Conf.GetStringOr(MCPPathKey, "/mcp"))

	instructions := p.Conf.GetStringOr(MCPInstructionsKey, "")
	serverName := p.Conf.GetStringOr(MCPServerNameKey, "siu-mcp-server")
	serverVersion := p.Conf.GetStringOr(MCPServerVersionKey, "1.0.0")
	p.cfg = &MCPConfig{
		ServerName:    serverName,
		ServerVersion: serverVersion,
		Instructions:  instructions,
	}
}

func (p *MiddlewareMCP) Condition() bool {
	if v := p.Conf.GetBoolOr(MCPDisableKey, true); v {
		return false
	}
	common.INFO("MCP endpoint enabled at %s", p.basePath)
	return true
}

func (p *MiddlewareMCP) Function() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(globalMCPToolMap) == 0 {
			c.Next()
			return
		}
		if c.Request.Method == http.MethodPost && c.Request.URL.Path == p.basePath {
			p.handleRequest(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (p *MiddlewareMCP) Order() int {
	return MCPMiddleOrder
}

// --- JSON-RPC handling ---

type jsonRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func (p *MiddlewareMCP) handleRequest(c *gin.Context) {
	var req jsonRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: map[string]any{
				"code":    -32700,
				"message": "Parse error",
			},
		})
		return
	}
	switch req.Method {
	case "initialize":
		initResult := map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    p.cfg.ServerName,
				"version": p.cfg.ServerVersion,
			},
		}
		if p.cfg.Instructions != "" {
			initResult["instructions"] = p.cfg.Instructions
		}
		c.JSON(http.StatusOK, jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  initResult,
		})
	case "notifications/initialized":
		c.Status(http.StatusAccepted)
	case "tools/list":
		c.JSON(http.StatusOK, jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": globalMCPToolList,
			},
		})
	case "tools/call":
		p.handleToolCall(c, req)
	default:
		c.JSON(http.StatusOK, jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: map[string]any{
				"code":    -32601,
				"message": "Method not found",
			},
		})
	}
}

func (p *MiddlewareMCP) handleToolCall(c *gin.Context, req jsonRPCRequest) {
	toolName, _ := req.Params["name"].(string)
	tool, ok := globalMCPToolMap[toolName]
	if !ok {
		c.JSON(http.StatusOK, jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: map[string]any{
				"code":    -32602,
				"message": "Tool not found: " + toolName,
			},
		})
		return
	}

	arguments, _ := req.Params["arguments"].(map[string]any)
	result, statusCode, err := p.callTool(c, tool, arguments)
	if err != nil {
		c.JSON(http.StatusOK, jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: map[string]any{
				"code":    -32603,
				"message": err.Error(),
			},
		})
		return
	}
	isError := statusCode >= 400
	c.JSON(http.StatusOK, jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": result,
				},
			},
			"isError": isError,
		},
	})
}

var passHeaders = []string{"Authorization", "Cookie"}

func (p *MiddlewareMCP) callTool(c *gin.Context, tool MCPToolDef, arguments map[string]any) (string, int, error) {
	path := tool.Path
	method := tool.Method

	remaining := make(map[string]any)
	for k, v := range arguments {
		remaining[k] = v
	}

	for k, v := range remaining {
		placeholder := ":" + k
		if strings.Contains(path, placeholder) {
			path = strings.Replace(path, placeholder, fmt.Sprintf("%v", v), 1)
			delete(remaining, k)
		}
	}

	var body io.Reader
	var contentType string

	if len(remaining) > 0 {
		if method == "GET" || method == "HEAD" {
			values := url.Values{}
			for k, v := range remaining {
				values.Set(k, fmt.Sprintf("%v", v))
			}
			path += "?" + values.Encode()
		} else if tool.ContentType == "multipart/form-data" {
			b, ct, err := buildMultipartBody(remaining, tool.FileParams)
			if err != nil {
				return "", 0, err
			}
			body = b
			contentType = ct
		} else if tool.ContentType == "application/x-www-form-urlencoded" {
			values := url.Values{}
			for k, v := range remaining {
				values.Set(k, fmt.Sprintf("%v", v))
			}
			body = strings.NewReader(values.Encode())
			contentType = "application/x-www-form-urlencoded"
		} else {
			bodyBytes, err := json.Marshal(remaining)
			if err != nil {
				return "", 0, err
			}
			body = bytes.NewReader(bodyBytes)
			contentType = "application/json"
		}
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return "", 0, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for _, h := range passHeaders {
		if v := c.GetHeader(h); v != "" {
			req.Header.Set(h, v)
		}
	}

	w := httptest.NewRecorder()
	p.Server.ServeHTTP(w, req)
	return w.Body.String(), w.Code, nil
}

func buildMultipartBody(args map[string]any, fileParams map[string]bool) (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for k, v := range args {
		if fileParams[k] {
			raw, ok := v.(string)
			if !ok {
				return nil, "", fmt.Errorf("file param %q must be a base64 string", k)
			}
			decoded, err := base64.StdEncoding.DecodeString(raw)
			if err != nil {
				return nil, "", fmt.Errorf("file param %q: invalid base64: %w", k, err)
			}
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, k, k))
			h.Set("Content-Type", "application/octet-stream")
			part, err := writer.CreatePart(h)
			if err != nil {
				return nil, "", err
			}
			if _, err := part.Write(decoded); err != nil {
				return nil, "", err
			}
		} else {
			if err := writer.WriteField(k, fmt.Sprintf("%v", v)); err != nil {
				return nil, "", err
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return &buf, writer.FormDataContentType(), nil
}

// generateMCPMeta auto-generates MCP tool metadata from method and path.
func generateMCPMeta(method string, path string) map[string]any {
	name := strings.ToLower(method)
	cleanPath := strings.ReplaceAll(path, ":", "")
	parts := strings.Split(cleanPath, "/")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			name += "_" + part
		}
	}

	properties := make(map[string]any)
	required := make([]string, 0)
	pathParts := strings.Split(path, "/")
	for _, part := range pathParts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			properties[paramName] = map[string]any{
				"type":        "string",
				"description": paramName,
			}
			required = append(required, paramName)
		}
	}

	inputSchema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": true,
	}
	if len(required) > 0 {
		inputSchema["required"] = required
	}

	return map[string]any{
		"name":        name,
		"description": method + " " + path,
		"inputSchema": inputSchema,
	}
}

// expandMCPRouteDef converts a RouteDef into MCP tool metadata.
func expandMCPRouteDef(method, path string, def RouteDef) map[string]any {
	summary := def.Summary
	if summary == "" {
		summary = method + " " + path
	}

	pathParamSet := make(map[string]bool)
	for _, p := range ExtractPathParams(path) {
		pathParamSet[p] = true
	}

	mcpName := def.Name
	if mcpName == "" {
		generated := generateMCPMeta(method, path)
		mcpName, _ = generated["name"].(string)
	}

	mcpProps := make(map[string]any)
	var mcpRequired []string
	for name, p := range ResolvedParams(def) {
		mcpProps[name] = ParamSchema(p)
		if p.Required || pathParamSet[name] {
			mcpRequired = append(mcpRequired, name)
		}
	}

	inputSchema := map[string]any{
		"type":                 "object",
		"properties":           mcpProps,
		"additionalProperties": true,
	}
	if len(mcpRequired) > 0 {
		inputSchema["required"] = mcpRequired
	}

	return map[string]any{
		"name":        mcpName,
		"description": summary,
		"inputSchema": inputSchema,
	}
}
