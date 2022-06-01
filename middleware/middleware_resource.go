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

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/config"
)

const (
	ResourceMiddlePrefixKey     = "middleware.resource.prefix"
	ResourceMiddleDisableKey    = "middleware.resource.disable"
	ResourceMiddleDefaultPrefix = "/resources"
	ResourceMiddleOrder         = 20
)

type MiddlewareResource struct {
	Conf config.TypedConfig `@siu:"name='environment',default='type'"`
}

func (p *MiddlewareResource) Condition() bool {
	if v, ok := p.Conf.GetBool(ResourceMiddleDisableKey); ok && v {
		return false
	}
	return true
}

func (p *MiddlewareResource) Function() gin.HandlerFunc {
	resourcesPrefix := p.Conf.GetStringOr(ResourceMiddlePrefixKey, ResourceMiddleDefaultPrefix)
	return Serve(resourcesPrefix, LocalFile("resources", true))
}

func (p *MiddlewareResource) Order() int {
	return ResourceMiddleOrder
}

type ServeFileSystem interface {
	http.FileSystem
	Exists(prefix string, path string) bool
}

type LocalFileSystem struct {
	http.FileSystem
	root    string
	indexes bool
}

func LocalFile(root string, indexes bool) *LocalFileSystem {
	return &LocalFileSystem{
		FileSystem: gin.Dir(root, indexes),
		root:       root,
		indexes:    indexes,
	}
}

func (l *LocalFileSystem) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		return true
	}
	return false
}
func (l *LocalFileSystem) Open(name string) (http.File, error) {
	f, err := l.FileSystem.Open(name)
	if err != nil {
		return l.FileSystem.Open("/index.html")
	}
	return f, err
}

func Serve(resourcesPrefix string, fs ServeFileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fs)
	if resourcesPrefix != "" {
		fileserver = http.StripPrefix(resourcesPrefix, fileserver)
	}
	return func(c *gin.Context) {
		uri := c.Request.URL.Path
		if uri == "/" || uri == "/index.html" {
			c.Redirect(http.StatusFound, resourcesPrefix)
			c.Abort()
			return
		}
		if fs.Exists(resourcesPrefix, c.Request.URL.Path) {
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}
