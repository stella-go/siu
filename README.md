# Create a Gin project with siu

siu `/sjuː/ meaning very very fast` is an secondary packaging of the [Gin](https://github.com/gin-gonic/gin) Web Framework to quickly build enterprise-level Web applications.

Siu quickly configures startup components through configuration files, such as rolling log, integrating component such as mysql, redis, zookeeper, CORS configuration, and support for history routing web applications. And also keeps open for extensions.

**Incompatibility Update**

Since v1.1.0, The Inversion of Control (IoC) feature was introduced. The struct field tag was used to inject dependencies and attributes, and the functions of siu for obtaining dependencies and attributes manually was removed.

Since v1.4.0, The following breaking changes were introduced:
- **Cipher Interface**: `Encrypt`, `PublickeyEncrypt`, `Sign` now return `(string, error)` instead of `string`. Callers must handle the error.
- **CORS Rename**: `MiddlewareCROS` renamed to `MiddlewareCORS`. Configuration keys changed from `middleware.cros.*` to `middleware.cors.*`. Environment variables changed from `STELLA_MIDDLEWARE_CROS_*` to `STELLA_MIDDLEWARE_CORS_*`.
- **JWT Secret**: No longer auto-generates a random secret. If `middleware.jwt.secret` is not configured, siu will derive the secret from the `application` configuration key. If neither is set, it panics.
- **Cron**: `siu.Cron()` now returns `error` instead of void.
- **TaggedLogger**: `TagedLogger` renamed to `TaggedLogger`.
- **Type Aliases**: All `interface{}` replaced with `any`.
- **Constant Fixes**: `OssDdisableSSLKey` renamed to `OssDisableSSLKey`.

## Installation
```bash
go get -u github.com/stella-go/siu
```

## Quick Start
```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu"
)

type HelloRouter struct{}

func (p *HelloRouter) Router() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GET /hello": p.Hello,
	}
}

func (p *HelloRouter) Hello(c *gin.Context) {
	c.String(200, "Hello.")
}

func main() {
	siu.Route(&HelloRouter{})
	siu.Run()
}
```
Yeah, you have completed the development of your first web application, try to open `http://127.0.0.1:8080/hello` in your browser.

## Routing Interface
```go
type Router interface {
	Router() map[string]gin.HandlerFunc
}
```
Implement this interface and then use `siu.Route(&HelloRouter{})` to register routers.
The return value of the function is a map, the key of the map needs to meet the format `"GET /hello"`, and the value of the map is the handler function.
```go
type MiddlewareRouter interface {
	Router
	Middleware() []gin.HandlerFunc
}
```
If the registered route is an implementation of `MiddlewareRouter`, the middlewares will be applied to the routing group.

## Configuration File
siu will load the configuration files in the following order:
1. Environment variable STELLA_CONFIG_FILES
2. application.yml
3. config/application.yml

**NOTICE**: If the same configuration item exists in different configuration files, the configuration loaded first will take effect.

Obtaining a configuration item:
```go
type Service struct {
	Content string             `@siu:"value='${my.system.content:defaultValue}'"`
}
```
or inject instances of the environment configuration, the interface `config.TypedConfig`
```go
type Service struct {
	Conf    config.TypedConfig `@siu:"name='environment',default='type'"`
}

func (p *Service) Handle() {
	fmt.Println(p.Conf.GetStringOr("my.system.content", "defaultValue"))
}
```

### Server Related Configuration
```yml
server:
  mode: release
  ip: 127.0.0.1
  port: 8080
  prefix: "/"
```
- **server.mode** Gin server mode. Optional value `release` or `debug`. Default value `release`.
- **server.ip** Gin server bind ip. Default value `0.0.0.0`.
- **server.port** Gin server port. Default value `8080`.
- **server.prefix** Gin routers prefix. Default value `/`.

### Logger Related Configuration
```yml
logger:
  siu: true
  level: info
  daily: true
  path: ./logs
  file: log.txt
  maxFiles: 31
  maxFileSize: 200
```
- **logger.siu** Whether to use the logging implementation of siu, set to false to use golang built-in log. Optional value `true` or `false`. Default value `true`.
- **logger.level** Log Level. Optional value `debug`, `info`, `warn` or `error`. Default value `info`.
- **logger.daily** Whether to enable daily log rotating. Optional value `true` or `false`. Default value `true`.
- **logger.path** Log Path Dir. Default value `.`.
- **logger.fileName** Log file name. Default value `stdout`, does not print logs to a file, but rather to the console as a standard output stream.
- **logger.maxFiles** Maximum number of files to be retained. Default value `30`.
- **logger.maxFileSize** Maximum file size. Default value `200`.

Obtaining a Logger instance:
```go
type Service struct {
	Logger  siu.Logger         `@siu:"name='logger',default='type'"`
}
```

### MySQL Related Configuration
```yml
mysql:
  user: root
  passwd: root
  addr: 127.0.0.1:3306
  dbName: test
  collation: utf8mb4_bin
  timeout: 100000
  readTimeout: 50000
  writeTimeout: 50000
```
- **mysql.user** mysql username.
- **mysql.passwd** mysql password.
- **mysql.addr** mysql server ip:port.
- **mysql.dbName** the name of the database to link to.
- **mysql.collation** character set. Default value `utf8mb4_bin`.
- **mysql.timeout** Connection timeout in milliseconds. Default value `60000`.
- **mysql.readTimeout** Read timeout in milliseconds. Default value `30000`.
- **mysql.writeTimeout** Write timeout in milliseconds. Default value `30000`.
Obtaining a MySQL instance:
```go
type Service struct {
	DB      *sql.DB            `@siu:""`
}
```

To use multiple data sources, configure as follows.
```yml
mysql:
  db1:
    user: root
    passwd: root
    addr: 127.0.0.1:3306
    dbName: test1
  db2:
    user: root
    passwd: root
    addr: 127.0.0.1:3306
    dbName: test2
```
Obtaining a MySQL instance:
```go
type Service struct {
	DB1     *sql.DB            `@siu:"name='mysql.db1'"`
	DB2     *sql.DB            `@siu:"name='mysql.db2'"`
}
```

### Gorm Related Configuration
```yml
gorm:
  user: root
  passwd: root
  addr: 127.0.0.1:3306
  dbName: test
  collation: utf8mb4_bin
  timeout: 100000
  readTimeout: 50000
  writeTimeout: 50000
```
- configurations are the same as MySQL

Obtaining a Gorm instance:
```go
type Service struct {
	DB      *gorm.DB            `@siu:""`
}
```

### Redis Related Configuration
```yml
redis:
  addr: 127.0.0.1:6379
  password: 
  db: 0
  poolSize: 4
  maxIdle: 1
  dialTimeout: 5000
  readTimeout: 5000
  writeTimeout: 5000
```
- **redis.addr** redis server ip:port, if it's a cluster ip1:port1,ip2:port2,ip3:port3.
- **redis.password** redis password.
- **redis.db** redis database serial number.
- **redis.poolSize** size of the redis connection pool. Default value `4`.
- **redis.minIdle** minimum idle number. Default value `1`.
- **redis.dialTimeout** connection timeout in milliseconds. Default value `5000`.
- **redis.readTimeout** read timeout in milliseconds. Default value `5000`.
- **redis.writeTimeout** write timeout in milliseconds. Default value `5000`.

Obtaining a Redis instance:
```go
type Service struct {
	Redis   redis.Cmdable      `@siu:""`
}
// Cmdable is the common interface of RedisClient and RedisClusterClient
```

### Zookeeper Related Configuration
```yml
zookeeper:
  servers: 127.0.0.1:2181,127.0.0.1:2182,127.0.0.1:2183
  sessionTimeoutKey: 60000
```
- **zookeeper.servers** zookeeper servers ip:port, if it's a cluster ip1:port1,ip2:port2,ip3:port3.
- **zookeeper.sessionTimeoutKey** session timeout in milliseconds. Default value `60000`.

Obtaining a Zookeeper instance:
```go
type Service struct {
	Zk      *zk.Conn           `@siu:""`
}
```

### OSS Related Configuration
```yml
oss:
  endpoint: 127.0.0.1:9000
  ak: <some ak>
  sk: <some sk>
  region: default
  disable-ssl: false
  force-path-style: true
```
- **oss.endpoint** oss server endpoint.
- **oss.ak** oss access key.
- **oss.sk** oss access secret.
- **oss.region** oss region. Default value `default`.
- **oss.disable-ssl** oss disable ssl access. Default value `false`.
- **oss.force-path-style** oss force use path stype. Default value `true`.

Obtaining a OSS instance:
```go
type Service struct {
	Oss      *s3.S3           `@siu:""`
}
```

### Cipher Related Configuration
```yml
cipher:
  key: <some aes hex value>
  hmac-key: <some hex value>
  public-key: |
    -----BEGIN RSA PUBLIC KEY-----
    MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCzm464TXngLXtZ1xdoGKoVodSM
    Sjd6q/Hwr/jRId9...
    -----END RSA PUBLIC KEY-----
  private-key: |
    -----BEGIN RSA PRIVATE KEY-----
    MIICXAIBAAKBgQCzm464TXngLXtZ1xdoGKoVodSMSjd6q/Hwr/jRId9WlM+VPglg
    snajexPi0GHDPL4...
    -----END RSA PRIVATE KEY----
```

Obtaining a Cipher instance:
```go
type Service struct {
	Cipher      interfaces.Cipher           `@siu:""`
}

func (p *Service) Handle() {
	encrypted, err := p.Cipher.Encrypt("plaintext")
	if err != nil {
		// handle error
	}
	fmt.Println(encrypted)
}
```

## Middleware Related Configuration
```yml
middleware:
  rewrite:
    disable: false  # Set whether to disable path rewrite, default true
    match: "^/something(/|$)(.*)" # Set match regexp
    rewrite: "/$2" # Set replace repl
  access.disable: false # Set whether to disable access logging, default false
  cors:
    disable: false # Set whether to disable CORS, default false
    wildcard: false # Set whether to enable wildcards, default true
    expose: "*" # Set "Access-Control-Expose-Headers", separated by commas, default "*"
  error-log.disable: false # Set whether to disable error logging, default false
  resource:
    disable: false # Set whether to disable resources serve, default false
    prefix: "/resources" # Set resources path prefix, default "/resources"
    index-not-found: false # Set whether to index when router not found, default false
    compress: true # Set whether to compress static resources, default true
  session:
    disable: false # Set whether to disable session middleware, default true
    timeout: 3600 # session idle timeout in seconds. Default value `86400`.
  jwt:
    disable: false # Set whether to disable jwt authorization, default false.
    cookie-domain: # Set domain the cookie will be set, default "".
    expire-seconds: 3600 # Set the jwt Token expire times.
    secret: <some value> # Set jwt secret. Required when JWT is enabled.
    excludes:  # Set jwt authorization exclude paths, default /login, /admin/login, /api/login.
      - "/login"
      - "/admin/login"
      - "/api/login"
```

**NOTICE**: Since v1.4.0, `middleware.jwt.secret` is required when JWT is enabled. If not set, siu will try to derive the secret from the `application` configuration key via SHA-256 hash. If neither is configured, the application will panic on startup.

## Swagger API Documentation
siu has built-in support for Swagger (OpenAPI 3.0) API documentation with a Swagger UI page. Both are disabled by default and can be enabled via configuration.

### Configuration
```yml
swagger:
  disable: false  # Set to false to enable Swagger, default true
  path: /swagger  # Swagger UI base path, default "/swagger"
  title: "My API"  # API document title, default "API Documentation"
  description: "API description"  # API description, default ""
  version: "1.0.0"  # API version, default "1.0.0"
```

When enabled, siu exposes:
- `GET {swagger.path}/` — Swagger UI page (loads JS/CSS from CDN)
- `GET {swagger.path}/doc.json` — OpenAPI 3.0 JSON spec

### Zero-Code Integration
When Swagger is enabled, siu **automatically generates** basic API documentation for all registered routes. No code changes are required — just set `swagger.disable: false` and restart the application.

Auto-generated metadata includes:
- Summary derived from method and path (e.g. `GET /users/{id}`)
- Path parameters automatically detected from `:param` patterns
- Request body schema for POST/PUT/PATCH methods (`additionalProperties: true`)
- 200 OK response

### Custom Metadata with `siu.Meta()`
To provide richer API descriptions, wrap handlers with `siu.Meta()` using typed structs:
```go
func (p *UserRouter) Router() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"POST /users": siu.Meta(p.CreateUser, siu.RouteDef{
			Name:    "create_user",       // MCP tool name (optional, auto-derived)
			Summary: "Create a new user", // → swagger.summary + mcp.description
			Params: map[string]siu.ParamDef{ // → swagger parameters/requestBody + mcp inputSchema
				"name":  {Type: "string", Description: "User name", Required: true},
				"email": {Type: "string", Description: "Email address"},
			},
		}),
		"GET /users/:id": siu.Meta(p.GetUser, siu.RouteDef{
			Summary: "Get user by ID",
			Params: map[string]siu.ParamDef{
				"id": {Type: "string", Description: "User ID", Required: true},
			},
		}),
		"DELETE /users/:id": p.DeleteUser, // auto-generated metadata
	}
}
```

### Struct Tag Metadata with `@meta`
Instead of manually building `Params`, you can pass request/response struct instances via the `Request` and `Response` fields. siu will automatically parse struct tags to generate parameter definitions.

The `@meta` tag uses `"k=v,k=v"` format. Supported keys:
| Key | Description | Example |
|---|---|---|
| `desc` | Field description | `desc=User name` |
| `required` | Mark as required (flag or `=true`) | `required` or `required=true` |
| `ignore` | Exclude field from output (flag or `=true`) | `ignore` or `ignore=true` |

Field names are derived from the `json` tag (falls back to the Go field name). Field types are automatically mapped from Go types (`string` → `"string"`, `int` → `"integer"`, `float64` → `"number"`, `bool` → `"boolean"`, struct → `"object"`, slice → `"array"`).

```go
type CreateUserRequest struct {
	Name  string `json:"name"  @meta:"desc=User name,required"`
	Email string `json:"email" @meta:"desc=Email address"`
	Age   int    `json:"age"   @meta:"desc=User age"`
}

type UserResponse struct {
	ID    string `json:"id"    @meta:"desc=User ID"`
	Name  string `json:"name"  @meta:"desc=User name"`
	Email string `json:"email" @meta:"desc=Email address"`
}

func (p *UserRouter) Router() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"POST /users": siu.Meta(p.CreateUser, siu.RouteDef{
			Summary:  "Create a new user",
			Request:  CreateUserRequest{},  // auto-parse request body schema
			Response: UserResponse{},       // auto-parse response schema (Swagger only)
		}),
	}
}
```

Features:
- **Nested structs**: Struct fields are recursively parsed as `"object"` with `properties`
- **Array/Slice fields**: Automatically mapped to `"array"` with element `items` schema
- **Embedded structs**: Anonymous struct fields are flattened into the parent
- **`json:"-"`**: Fields with `json:"-"` are skipped
- **`time.Time`**: Automatically mapped to `"string"` type
- **Manual override**: `Params` entries take precedence over auto-parsed `Request` fields when both are provided
Routes without decorators will use auto-generated metadata. Routes with decorators will use the provided metadata instead.

`siu.Meta()` automatically generates both Swagger and MCP formats from a single definition:
- `Summary` → `swagger.summary` + `mcp.description`
- `Params` → `swagger.parameters` / `swagger.requestBody` + `mcp.inputSchema`
- Path params (e.g. `:id`) are auto-detected as `in: path` for Swagger and `required` for MCP
- Non-path params become `in: query` for GET/DELETE/HEAD, or `requestBody` properties for POST/PUT/PATCH

## MCP (Model Context Protocol)
siu has built-in support for MCP Streamable HTTP, enabling AI agents (such as Claude, Cursor, etc.) to discover and invoke your API as tools. Disabled by default.

### Configuration
```yml
mcp:
  disable: false  # Set to false to enable MCP, default true
  path: /mcp      # MCP endpoint path, default "/mcp"
```

When enabled, siu exposes a `POST {mcp.path}` endpoint implementing the MCP JSON-RPC 2.0 protocol:
- `initialize` — Returns server capabilities
- `tools/list` — Returns all registered tools
- `tools/call` — Invokes a tool by name with arguments

### Zero-Code Integration
When MCP is enabled, siu **automatically generates** tool definitions for all registered routes. No code changes are required — just set `mcp.disable: false` and restart the application.

Auto-generated tool definitions include:
- Tool name derived from method and path (e.g. `GET /users/:id` → `get_users_id`)
- Path parameters as required fields in inputSchema
- Open schema (`additionalProperties: true`) for additional parameters

### Custom Metadata with `siu.Meta()` (same decorator)
`siu.Meta()` works for both Swagger and MCP simultaneously — see the example above. There is no need for separate MCP-specific decorators.

**Decorator function summary:**
| Function | Description |
|---|---|
| `siu.Meta(handler, siu.RouteDef{...})` | Attach typed metadata, generates both Swagger and MCP |

| RouteDef Field | Description |
|---|---|
| `Name` | MCP tool name (optional, auto-derived from method+path) |
| `Summary` | → `swagger.summary` + `mcp.description` |
| `Params` | Manual parameter definitions → `swagger.parameters`/`requestBody` + `mcp.inputSchema` |
| `Request` | Request struct instance; fields are auto-parsed via `@meta` tag |
| `Response` | Response struct instance; auto-parsed for Swagger response schema |

The decorator returns the original handler unchanged.

### Authentication Pass-Through
When MCP is used alongside authentication middleware (JWT, Session, etc.), siu automatically forwards `Authorization` and `Cookie` headers from the external MCP request to the internal tool call. This ensures that auth middleware applies transparently — the AI client simply includes the same credentials it would use for a direct API call.

No additional configuration is needed. If the AI client sends a valid `Authorization: Bearer <token>` or `Cookie` header on the MCP request, the internal tool dispatch will carry the same headers and pass through any middleware checks.

## Cron Scheduling
Register cron jobs using `siu.Cron()`. The cron expression follows the standard 6-field format (second minute hour day month weekday).
```go
func main() {
	err := siu.Cron("*/5 * * * * ?", func() {
		fmt.Println("executed every 5 seconds")
	})
	if err != nil {
		log.Fatal(err)
	}
	siu.Run()
}
```

## Utility Functions

### fn.ToSnakeCase
Convert CamelCase strings to snake_case.
```go
import "github.com/stella-go/siu/fn"

fn.ToSnakeCase("MyFunction")  // "my_function"
fn.ToSnakeCase("HTTPServer")  // "h_t_t_p_server"
```

### fn.IfElse
Generic ternary helper.
```go
import "github.com/stella-go/siu/fn"

result := fn.IfElse(true, "yes", "no")  // "yes"
```

## Custom Injection
Implement the InjectRegister interface and use `siu.Register()` to register.

## Custom Component
Implement the AutoFactory interface and use `siu.AutoFactory()` to register.

## Custom Middleware
Implement the OrderedMiddleware interface and use `siu.Use()` to register.

## Instructions on Dependency Injection
All struct pointers registered in siu will perform dependency injection. All fields of struct are scanned by siu, and fields with the `@siu` tag are processed. After all fields are injected, if the struct/pointer implements `Initializable` interface, its `Init` function is executed.

### Tag
- Inject a configuration item
  ```go
  type Service struct {
    // This will look for the `my.system.content` configuration item and cause panic if it is not found.
    Content1 string             `@siu:"value='${my.system.content}'"`
    // This will look for the `my.system.content` configuration item and inject "defaultValue" if it is not found
    Content2 string             `@siu:"value='${my.system.content:defaultValue}'"`
    // This will look for the `my.system.content` configuration item and inject "defaultValue" if it is not found
    Content3 string             `@siu:"value='${my.system.content}',default='defaultValue'"`
  }
  ```

  - Injecting a dependency
  ```go
  type Foo interface{}
  type Service struct {
    // This will look for the object of type `Foo` and cause panic if it is not found
    Foo      Foo               `@siu:""`
    // This will look for the object of name "foo" and cause panic if it is not found
    Fop      Foo               `@siu:"name='foo'"`
    // This will look for the object of name "foo" and set to `nil` if it is not found
    Foq      Foo               `@siu:"name='foo',default='zero'"`
    // This will look for the object of name "foo", and then look for the object type `Foo` and cause panic if it is not found in either
    For      Foo               `@siu:"name='foo',default='type'"`
  }
  ```

  ```go
  type Bar struct{}

  func (*Bar) Init() {
    // This function is executed each time siu inject creates an instance of type Bar/*Bar
    fmt.Println("Bar")
  }

  type Service struct {
    // This will create a object of type `Bar`
    Bar      Bar               `@siu:""`
    // This will create a private object of type `*Bar`
    Bas      *Bar              `@siu:"type='private'"`
    // This will look for the object of type `*Bar` and create if it is not found
    Bat      *Bar              `@siu:""`
    // This will look for the object of name "bar" and cause panic if it is not found
    Bau      *Bar              `@siu:"name='bar'"`
    // This will look for the object of name "bar" and set to `nil` if it is not found
    Bav      *Bar              `@siu:"name='bar',default='zero'"`
    // This will look for the object of name "bar", and then look for the object type `*Bar`
    // If neither is found, an object of type `*Bar` will be created and its name and type will be stored for use in the next search
    Baw      *Bar              `@siu:"name='bar',default='type'"`
  }
  ```
