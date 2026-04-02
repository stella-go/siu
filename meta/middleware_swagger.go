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
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/common"
	"github.com/stella-go/siu/config"
)

var swaggerUITemplate string = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Swagger UI</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "{{SPEC_URL}}",
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset,
                ],
                layout: "BaseLayout",
                persistAuthorization: true,
                requestInterceptor: function(req) {
                    req.credentials = 'same-origin';
                    return req;
                },
            });
        };
    </script>
</body>
</html>
`

const (
	SwaggerDisableKey  = "swagger.disable"
	SwaggerPathKey     = "swagger.path"
	SwaggerMiddleOrder = 40
)

var globalSwaggerPaths = make(map[string]map[string]any)

func CollectSwaggerData(methods []string, fullPath string, def *RouteDef) {
	var effective map[string]any
	if def != nil {
		effective = expandSwaggerRouteDef(strings.ToUpper(methods[0]), fullPath, *def)
	}

	if _, ok := globalSwaggerPaths[fullPath]; !ok {
		globalSwaggerPaths[fullPath] = make(map[string]any)
	}
	if effective != nil {
		for _, method := range methods {
			globalSwaggerPaths[fullPath][strings.ToLower(method)] = effective
		}
	} else {
		for _, method := range methods {
			globalSwaggerPaths[fullPath][strings.ToLower(method)] = generateSwaggerMeta(strings.ToUpper(method), fullPath)
		}
	}
}

// --- MiddlewareSwagger ---

type MiddlewareSwagger struct {
	Server *gin.Engine        `@siu:"name='server',default='type'"`
	Conf   config.TypedConfig `@siu:"name='environment',default='type'"`

	basePath    string
	title       string
	description string
	version     string
	specJSON    []byte
	html        string
	once        sync.Once
}

func (p *MiddlewareSwagger) Init() {
	serverPrefix := p.Conf.GetStringOr("server.prefix", "")
	p.basePath = path.Join(serverPrefix, p.Conf.GetStringOr(SwaggerPathKey, "/swagger"))
	p.title = p.Conf.GetStringOr("swagger.title", "API Documentation")
	p.description = p.Conf.GetStringOr("swagger.description", "")
	p.version = p.Conf.GetStringOr("swagger.version", "1.0.0")
}

func (p *MiddlewareSwagger) buildSpec() {

	openAPIPaths := make(map[string]map[string]any)
	for path, methods := range globalSwaggerPaths {
		openAPIPath := ginPathToOpenAPI(path)
		openAPIPaths[openAPIPath] = methods
	}

	spec := map[string]any{
		"openapi": "3.0.0",
		"info": map[string]any{
			"title":       p.title,
			"description": p.description,
			"version":     p.version,
		},
		"paths": openAPIPaths,
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"Authorization": map[string]any{
					"type": "apiKey",
					"in":   "header",
					"name": "Authorization",
				},
			},
		},
		"security": []map[string]any{
			{"Authorization": []string{}},
		},
	}

	p.specJSON, _ = json.MarshalIndent(spec, "", "  ")
	specURL := p.basePath + "/doc.json"
	p.html = strings.Replace(swaggerUITemplate, "{{SPEC_URL}}", specURL, 1)
}

func (p *MiddlewareSwagger) Condition() bool {
	if v := p.Conf.GetBoolOr(SwaggerDisableKey, true); v {
		return false
	}
	common.INFO("Swagger UI enabled at %s/", p.basePath)
	return true
}

func (p *MiddlewareSwagger) Function() gin.HandlerFunc {
	return func(c *gin.Context) {
		p.once.Do(func() {
			if len(globalSwaggerPaths) > 0 {
				p.buildSpec()
			}
		})
		if p.specJSON == nil {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if c.Request.Method == http.MethodGet {
			if path == p.basePath+"/doc.json" {
				c.Data(http.StatusOK, "application/json; charset=utf-8", p.specJSON)
				c.Abort()
				return
			}
			if path == p.basePath+"/" || path == p.basePath {
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(p.html))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func (p *MiddlewareSwagger) Order() int {
	return SwaggerMiddleOrder
}

func ginPathToOpenAPI(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

// generateSwaggerMeta auto-generates Swagger operation metadata from method and path.
func generateSwaggerMeta(method string, path string) map[string]any {
	openAPIPath := ginPathToOpenAPI(path)
	meta := map[string]any{
		"summary": method + " " + openAPIPath,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK"},
		},
	}

	params := ExtractPathParams(path)
	if len(params) > 0 {
		paramList := make([]map[string]any, 0, len(params))
		for _, p := range params {
			paramList = append(paramList, map[string]any{
				"name":     p,
				"in":       "path",
				"required": true,
				"schema":   map[string]any{"type": "string"},
			})
		}
		meta["parameters"] = paramList
	}

	if method == "POST" || method == "PUT" || method == "PATCH" {
		meta["requestBody"] = map[string]any{
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
		}
	}
	return meta
}

// expandSwaggerRouteDef converts a RouteDef into Swagger operation metadata.
func expandSwaggerRouteDef(method, path string, def RouteDef) map[string]any {
	summary := def.Summary
	if summary == "" {
		summary = method + " " + path
	}

	pathParamSet := make(map[string]bool)
	for _, p := range ExtractPathParams(path) {
		pathParamSet[p] = true
	}

	swagger := map[string]any{
		"summary": summary,
	}

	respSchema := StructToSchema(def.Response)
	if respSchema != nil {
		swagger["responses"] = map[string]any{
			"200": map[string]any{
				"description": "OK",
				"content": map[string]any{
					"application/json": map[string]any{"schema": respSchema},
				},
			},
		}
	} else {
		swagger["responses"] = map[string]any{"200": map[string]any{"description": "OK"}}
	}

	var swaggerParams []map[string]any
	bodyProps := make(map[string]any)
	var bodyRequired []string

	for name, p := range ResolvedParams(def) {
		if pathParamSet[name] {
			sp := map[string]any{
				"name":     name,
				"in":       "path",
				"required": true,
				"schema":   ParamSchema(p),
			}
			if p.Description != "" {
				sp["description"] = p.Description
			}
			swaggerParams = append(swaggerParams, sp)
		} else if method == "GET" || method == "HEAD" {
			sp := map[string]any{
				"name":   name,
				"in":     "query",
				"schema": ParamSchema(p),
			}
			if p.Description != "" {
				sp["description"] = p.Description
			}
			if p.Required {
				sp["required"] = true
			}
			swaggerParams = append(swaggerParams, sp)
		} else {
			bodyProps[name] = ParamSchema(p)
			if p.Required {
				bodyRequired = append(bodyRequired, name)
			}
		}
	}

	if len(swaggerParams) > 0 {
		swagger["parameters"] = swaggerParams
	}
	if len(bodyProps) > 0 {
		schema := map[string]any{
			"type":       "object",
			"properties": bodyProps,
		}
		if len(bodyRequired) > 0 {
			schema["required"] = bodyRequired
		}
		mediaType := "application/json"
		if def.ContentType != "" {
			mediaType = def.ContentType
		}
		swagger["requestBody"] = map[string]any{
			"content": map[string]any{
				mediaType: map[string]any{"schema": schema},
			},
		}
	}

	return swagger
}
