// Copyright 2010-2021 the original author or authors.

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
	ResourceMiddleDisableKey = "middleware.resource.disable"
	ResourceMiddleOrder      = 20
)

type MiddlewareResource struct{}

func (p *MiddlewareResource) Condition() bool {
	if v, ok := config.GetBool(ResourceMiddleDisableKey); ok && v {
		return false
	}
	return true
}

func (p *MiddlewareResource) Function() gin.HandlerFunc {
	return Serve("/resources", LocalFile("resources", true))
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

func Serve(urlPrefix string, fs ServeFileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		if fs.Exists(urlPrefix, c.Request.URL.Path) {
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}
