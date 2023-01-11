// Copyright 2010-2023 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/stella-go/siu/common"
	"gopkg.in/yaml.v2"
)

var null struct{}

var (
	env          *environment
	defaultFiles = []string{"application.yml", "config/application.yml"}
	rwLock       = &sync.RWMutex{}
)

type environment struct {
	files   []string
	configs []map[interface{}]interface{}
}

func init() {
	names := defaultFiles
	envConfigFiles := os.Getenv("STELLA_CONFIG_FILES")
	if envConfigFiles != "" {
		files := strings.Split(envConfigFiles, ",")
		names = append(files, names...)
	}
	maps := make([]map[interface{}]interface{}, 0)
	for _, file := range names {
		m := make(map[interface{}]interface{})
		bts, err := os.ReadFile(file)
		if err != nil {
			maps = append(maps, m)
			continue
		}
		err = yaml.Unmarshal(bts, &m)
		if err != nil {
			maps = append(maps, m)
			continue
		}
		maps = append(maps, m)
		common.INFO("Load configuration file: %s success", file)
	}
	env = &environment{names, maps}
}

func LoadConfig(files ...string) {
	rwLock.Lock()
	defer rwLock.Unlock()
	alreadyDone := make(map[string]struct{})
	for _, f := range env.files {
		alreadyDone[f] = null
	}
	maps := make([]map[interface{}]interface{}, 0)
	for _, file := range files {
		if _, done := alreadyDone[file]; done {
			continue
		}
		m := make(map[interface{}]interface{})
		bts, err := os.ReadFile(file)
		if err != nil {
			maps = append(maps, m)
			common.ERROR("Failed to read configuration file: %s, with error %v", file, err)
			continue
		}
		err = yaml.Unmarshal(bts, &m)
		if err != nil {
			maps = append(maps, m)
			common.ERROR("Failed to unmarshal configuration file: %s, with error %v", file, err)
			continue
		}
		maps = append(maps, m)
	}
	env.files = append(env.files, files...)
	env.configs = append(env.configs, maps...)
}

func (p *environment) tryLoadOSEnv(key string) (interface{}, bool) {
	key = strings.ReplaceAll(key, ".", "_")
	key = strings.ReplaceAll(key, "-", "_")
	key = "STELLA_" + strings.ToUpper(key)
	value := os.Getenv(key)
	if value == "" {
		return nil, false
	}
	return value, true
}

func (p *environment) Get(key string) (interface{}, bool) {
	rwLock.RLock()
	defer rwLock.RUnlock()
	value, ok := p.tryLoadOSEnv(key)
	if ok {
		return value, ok
	}
	for _, config := range p.configs {
		value, ok := get(config, key)
		if ok {
			return value, true
		}
	}
	return nil, false
}

func (p *environment) GetInt(key string) (int, bool) {
	value, ok := p.Get(key)
	if !ok {
		return 0, false
	}
	switch value := value.(type) {
	case int:
		return value, true
	case string:
		intValue, err := strconv.ParseInt(value, 0, 0)
		if err != nil {
			common.ERROR("Failed to get configuration: %s=%s, with error %v", key, value, err)
			return 0, false
		}
		return int(intValue), true
	default:
		common.ERROR("Failed to get configuration: %s=%v", key, value)
		return 0, false
	}
}

func (p *environment) GetString(key string) (string, bool) {
	value, ok := p.Get(key)
	if !ok {
		return "", false
	}
	switch value := value.(type) {
	case string:
		return value, true
	case int:
		return strconv.Itoa(value), true
	default:
		common.ERROR("Failed to get configuration: %s=%v", key, value)
		return "", false
	}
}

func (p *environment) GetBool(key string) (bool, bool) {
	value, ok := p.Get(key)
	if !ok {
		return false, false
	}
	switch value := value.(type) {
	case bool:
		return value, true
	case int:
		if value == 0 {
			return false, true
		}
		if value == 1 {
			return true, true
		}
		common.ERROR("Failed to get configuration: %s=%d", key, value)
		return false, false
	case string:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			common.ERROR("Failed to get configuration: %s=%s, with error %v", key, value, err)
			return false, false
		}
		return boolValue, true
	default:
		common.ERROR("Failed to get configuration: %s=%s", key, value)
		return false, false
	}
}

func (p *environment) GetOr(key string, defaultValue interface{}) interface{} {
	value, ok := p.Get(key)
	if !ok {
		return defaultValue
	} else {
		return value
	}
}

func (p *environment) GetIntOr(key string, defaultValue int) int {
	value, ok := p.GetInt(key)
	if !ok {
		return defaultValue
	} else {
		return value
	}
}

func (p *environment) GetBoolOr(key string, defaultValue bool) bool {
	value, ok := p.GetBool(key)
	if !ok {
		return defaultValue
	} else {
		return value
	}
}

func (p *environment) GetStringOr(key string, defaultValue string) string {
	value, ok := p.GetString(key)
	if !ok {
		return defaultValue
	} else {
		return value
	}
}

func get(config interface{}, key string) (interface{}, bool) {
	switch config.(type) {
	case map[interface{}]interface{}:
		m := config.(map[interface{}]interface{})
		tokens := strings.Split(key, ".")
		tokensLen := len(tokens)
		for i := range tokens {
			tmpK := strings.Join(tokens[:tokensLen-i], ".")
			v, ok := m[tmpK]
			if ok {
				switch v.(type) {
				case map[interface{}]interface{}:
					leftKey := strings.Join(tokens[tokensLen-i:], ".")
					if len(leftKey) == 0 {
						return v, true
					}
					return get(v, leftKey)
				default:
					if i == 0 {
						return v, true
					}
				}
			}
		}
		return nil, false
	default:
		return nil, false
	}
}

type Config interface {
	Get(key string) (interface{}, bool)
	GetOr(key string, defaultValue interface{}) interface{}
}

type TypedConfig interface {
	Config
	GetInt(key string) (int, bool)
	GetBool(key string) (bool, bool)
	GetString(key string) (string, bool)
	GetIntOr(key string, defaultValue int) int
	GetBoolOr(key string, defaultValue bool) bool
	GetStringOr(key string, defaultValue string) string
}

type ConfigurationEnvironment struct{}

func (p *ConfigurationEnvironment) Get(key string) (interface{}, bool) {
	return env.Get(key)
}

func (p *ConfigurationEnvironment) GetOr(key string, defaultValue interface{}) interface{} {
	return env.GetOr(key, defaultValue)
}

func (p *ConfigurationEnvironment) GetInt(key string) (int, bool) {
	return env.GetInt(key)
}

func (p *ConfigurationEnvironment) GetString(key string) (string, bool) {
	return env.GetString(key)
}

func (p *ConfigurationEnvironment) GetBool(key string) (bool, bool) {
	return env.GetBool(key)
}

func (p *ConfigurationEnvironment) GetIntOr(key string, defaultValue int) int {
	return env.GetIntOr(key, defaultValue)
}

func (p *ConfigurationEnvironment) GetBoolOr(key string, defaultValue bool) bool {
	return env.GetBoolOr(key, defaultValue)
}

func (p *ConfigurationEnvironment) GetStringOr(key string, defaultValue string) string {
	return env.GetStringOr(key, defaultValue)
}

type Cipher interface {
	Decrypt(string) (string, error)
}

type DecryptEnvironment struct {
	Cipher Cipher
}

func (p *DecryptEnvironment) Get(key string) (interface{}, bool) {
	return env.Get(key)
}

func (p *DecryptEnvironment) GetOr(key string, defaultValue interface{}) interface{} {
	return env.GetOr(key, defaultValue)
}

func (p *DecryptEnvironment) GetInt(key string) (int, bool) {
	return env.GetInt(key)
}

func (p *DecryptEnvironment) GetString(key string) (string, bool) {
	if p.Cipher == nil {
		return env.GetString(key)
	}
	if value, ok := env.GetString(key); ok {
		if srcVal, err := p.Cipher.Decrypt(value); err != nil {
			common.ERROR("Failed to decrypt configuration: %s=%s, with error %v", key, value, err)
			return "", false
		} else {
			return srcVal, ok
		}
	} else {
		return value, ok
	}
}

func (p *DecryptEnvironment) GetBool(key string) (bool, bool) {
	return env.GetBool(key)
}

func (p *DecryptEnvironment) GetIntOr(key string, defaultValue int) int {
	return env.GetIntOr(key, defaultValue)
}

func (p *DecryptEnvironment) GetBoolOr(key string, defaultValue bool) bool {
	return env.GetBoolOr(key, defaultValue)
}

func (p *DecryptEnvironment) GetStringOr(key string, defaultValue string) string {
	if p.Cipher == nil {
		return env.GetStringOr(key, defaultValue)
	}
	value := env.GetStringOr(key, defaultValue)
	if srcVal, err := p.Cipher.Decrypt(value); err != nil {
		common.ERROR("Failed to decrypt configuration: %s=%s, with error %v", key, value, err)
		return defaultValue
	} else {
		return srcVal
	}
}
