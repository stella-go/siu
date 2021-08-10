# Create a Gin project with siu

siu `/sjuː/ meaning very very fast` is an secondary packaging of the Gin Web Framework to quickly build enterprise-level Web applications.

Siu quickly configures startup components through configuration files, such as rolling log, integrating component such as mysql, redis, zookeeper, CROS configuration, and support for history routing web applications. And also keeps open for extensions.

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

type HelloRoute struct{}

func (p *HelloRoute) Router() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GET /hello": p.Hello,
	}
}

func (p *HelloRoute) Hello(c *gin.Context) {
	c.String(200, "Hello.")
}

func main() {
	siu.Route(&HelloRoute{})
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
and then use `siu.Route(&HelloRoute{})` to register routers.
The return value of the method is a map, the key of the map needs to meet the format of <method><space><path>, and the value of the map is the handler function.

## Configuration File
siu will load the configuration files in the following order:
1. Environment variable STELLA_CONFIG_FILES
2. application.yml
3. config/application.yml
**NOTICE**: If the same configuration exists in different configuration files, the configuration loaded first will take effect.

Get a custom configuration, use these methods `siu.EnvGetX()` or `siu.EnvGetXOr()`.

### Server Related Configuration
```yml
server:
  mode: release
  ip: 127.0.0.1
  port: 8080
```
- **server.mode** Gin server mode. Optional value `release` or `debug`. Default value `release`.
- **server.ip** Gin server ip. Default value `0.0.0.0`.
- **server.port** Gin server port. Default value `8080`.

### Logger Related Configuration
```yml
logger:
  level: info
  daliy: true
  path: ./logs
  file: log.txt
  maxFiles: 31
  maxFileSize: 200
```
- **logger.level** Log Level. Optional value `debug`, `info`, `warn` or `error`. Default value `info`.
- **logger.daliy** Whether to enable daily log rotating. Optional value `true` or `false`. Default value `true`.
- **logger.path** Log Path Dir. Default value `.`.
- **logger.fileName** Log file name. Default value `stdout`, does not print logs to a file, but rather to the console as a standard output stream.
- **logger.maxFiles** Maximum number of files to be retained. Default value `30`.
- **logger.maxFileSize** Maximum file size. Default value `200`.

Obtaining a Logger instance:
```go
rootLogger := siu.RootLogger()
namedLogger := siu.NewLogger(name)
```

### MySQL Related Configuration
```yml
mysql:
  user: root
  passwd: root
  addr: 127.0.0.1:3306
  dbName: test
  collation: utf8
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
db,ok := siu.DataSource()
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
db1,ok := siu.DataSourceWithName("db1")
db2,ok := siu.DataSourceWithName("db2")
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
redis,ok := siu.Redis()
redisCluster,ok := siu.RedisCluster()
```

### Redis Related Configuration
```yml
zookeeper:
  servers: 127.0.0.1:2181,127.0.0.1:2182,127.0.0.1:2183
  sessionTimeoutKey: 60000
```
- **zookeeper.servers** zookeeper servers ip:port, if it's a cluster ip1:port1,ip2:port2,ip3:port3.
- **zookeeper.sessionTimeoutKey** session timeout in milliseconds. Default value `60000`.

## Custom Component
Implement the AutoConfig interface and use `siu.AutoConfig()` to register.

## Custom Middleware
Implement the OrderedMiddleware interface and use `siu.Use()` to register.
