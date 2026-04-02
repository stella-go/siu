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

package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stella-go/siu/config"
)

const (
	ServerPrefix                   = "server.prefix"
	ResourceMiddlePrefixKey        = "middleware.resource.prefix"
	ResourceMiddleExcludeKey       = "middleware.resource.exclude"
	ResourceMiddleIndexNotFoundKey = "middleware.resource.index-not-found"
	ResourceMiddleDisableKey       = "middleware.resource.disable"
	ResourceMiddleCompressKey      = "middleware.resource.compress"

	ResourceMiddleDefaultPrefix = "/resources"
	ResourceMiddleOrder         = 40

	ContextResourceKey = "ResourcesKey"
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
	serverPrefix := p.Conf.GetStringOr(ServerPrefix, "")
	resourcePrefix := p.Conf.GetStringOr(ResourceMiddlePrefixKey, ResourceMiddleDefaultPrefix)
	prefix := path.Join(serverPrefix, resourcePrefix)
	resourceExclude := p.Conf.GetStringOr(ResourceMiddleExcludeKey, "")
	excludes := strings.Split(resourceExclude, ",")
	if !slices.Contains(excludes, "/swagger") {
		excludes = append(excludes, "/swagger")
	}
	if !slices.Contains(excludes, "/mcp") {
		excludes = append(excludes, "/mcp")
	}
	for i, exclude := range excludes {
		excludes[i] = path.Join(serverPrefix, strings.TrimSpace(exclude))
	}
	indexNotFound := p.Conf.GetBoolOr(ResourceMiddleIndexNotFoundKey, false)
	compress := p.Conf.GetBoolOr(ResourceMiddleCompressKey, true)
	return Serve(prefix, excludes, indexNotFound, compress, LocalFile("resources", true))
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
	if p := strings.TrimPrefix(filepath, prefix); prefix != "/" && len(p) < len(filepath) {
		return true
	}
	_, err := l.FileSystem.Open(filepath)
	return err == nil
}

func (l *LocalFileSystem) Open(name string) (http.File, error) {
	f, err := l.FileSystem.Open(name)
	if err != nil {
		return l.FileSystem.Open("/index.html")
	}
	return f, err
}

type GzipResponseWriter struct {
	gin.ResponseWriter
	gz io.Writer
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	return w.gz.Write(b)
}

func (w *GzipResponseWriter) WriteString(s string) (int, error) {
	return w.gz.Write([]byte(s))
}

func Serve(prefix string, excludes []string, indexNotFound bool, compress bool, fs ServeFileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fs)
	if prefix != "" {
		fileserver = http.StripPrefix(prefix, fileserver)
	}
	return func(c *gin.Context) {
		if c.FullPath() != "" {
			return
		}
		for _, exclude := range excludes {
			if exclude != "/" && strings.HasPrefix(c.Request.URL.Path, exclude) {
				return
			}
		}
		writer := c.Writer
		acceptGzip := false
		if compress {
			acceptEncodings := strings.Split(c.GetHeader("Accept-Encoding"), ", ")
			for _, accept := range acceptEncodings {
				encoding := strings.TrimSpace(strings.Split(accept, ";")[0])
				if encoding == "gzip" {
					acceptGzip = true
					break
				}
			}
			if acceptGzip {
				writer.Header().Set("Content-Encoding", "gzip")
				writer.Header().Set("Vary", "Accept-Encoding")
				gz, _ := gzip.NewWriterLevel(writer, gzip.BestCompression)
				writer = &GzipResponseWriter{ResponseWriter: writer, gz: gz}
				defer gz.Close()
			}
		}

		uri := c.Request.URL.Path
		if (prefix != "" && prefix != "/") && (uri == "/" || uri == "/index.html") {
			c.Request.URL.Path = prefix
			c.Request.RequestURI = strings.Replace(c.Request.RequestURI, uri, prefix, 1)
			fileserver.ServeHTTP(writer, c.Request)
			c.Set(ContextResourceKey, true)
			c.Abort()
			return
		}
		if fs.Exists(prefix, c.Request.URL.Path) {
			fileserver.ServeHTTP(writer, c.Request)
			c.Set(ContextResourceKey, true)
			c.Abort()
			return
		}
		if indexNotFound {
			c.Request.URL.Path = prefix
			c.Request.RequestURI = strings.Replace(c.Request.RequestURI, uri, prefix, 1)
			fileserver.ServeHTTP(writer, c.Request)
			c.Set(ContextResourceKey, true)
			c.Abort()
			return
		}
	}
}
