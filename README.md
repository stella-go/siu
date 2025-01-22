# Create a Gin project with siu

siu `/sjuː/ meaning very very fast` is an secondary packaging of the [Gin](https://github.com/gin-gonic/gin) Web Framework to quickly build enterprise-level Web applications.

Siu quickly configures startup components through configuration files, such as rolling log, integrating component such as mysql, redis, zookeeper, CROS configuration, and support for history routing web applications. And also keeps open for extensions.

**Incompatibility Update**

Since v1.1.0, The Inversion of Control (IoC) feature was introduced. The struct field tag was used to inject dependencies and attributes, and the functions of siu for obtaining dependencies and attributes manually was removed.

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
  daliy: true
  path: ./logs
  file: log.txt
  maxFiles: 31
  maxFileSize: 200
```
- **logger.siu** Whether to use the logging implementation of siu, set to false to use golang built-in log. Optional value `true` or `false`. Default value `true`.
- **logger.level** Log Level. Optional value `debug`, `info`, `warn` or `error`. Default value `info`.
- **logger.daliy** Whether to enable daily log rotating. Optional value `true` or `false`. Default value `true`.
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
```

## Middleware Related Configuration
```yml
middleware:
  rewrite:
    disable: false  # Set whether to disable path rewrite, default true
    match: "^/something(/|$)(.*)" # Set match regexp
    rewrite: "/$2" # Set replace repl
  access.disable: false # Set whether to disable access logging, default false
  cros:
    disable: false # Set whether to disable CROS, default false
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
    cookie-domain: # # Set domain the cookie will be set, default "".
    expire-seconds: 3600 # Set the jwt Token expire times.
    secret: <some value> # Set jwt secret, default random value.
    excludes:  # Set jwt authorization exclude paths, default /login, /admin/login, /api/login.
      - "/login"
      - "/admin/login"
      - "/api/login"


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
