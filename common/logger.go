// Copyright 2010-2025 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"fmt"
	"log"

	"github.com/stella-go/logger"
)

var (
	level logger.Level = logger.InfoLevel
)

// Not concurrent safe, used only during siu context initialization
func SetLevel(lv logger.Level) {
	level = lv
}

func DEBUG(format string, v ...interface{}) {
	if level <= logger.DebugLevel {
		if len(v) > 0 {
			if _, ok := v[len(v)-1].(error); ok {
				format += " %v"
			}
		}
		msg := fmt.Sprintf(format, v...)
		log.Output(2, "DEBUG - [SIU] "+msg)
	}
}

func INFO(format string, v ...interface{}) {
	if level <= logger.InfoLevel {
		if len(v) > 0 {
			if _, ok := v[len(v)-1].(error); ok {
				format += " %v"
			}
		}
		msg := fmt.Sprintf(format, v...)
		log.Output(2, "INFO  - [SIU] "+msg)
	}
}

func WARN(format string, v ...interface{}) {
	if level <= logger.WarnLevel {
		if len(v) > 0 {
			if _, ok := v[len(v)-1].(error); ok {
				format += " %v"
			}
		}
		msg := fmt.Sprintf(format, v...)
		log.Output(2, "WARN  - [SIU] "+msg)
	}
}

func ERROR(format string, v ...interface{}) {
	if level <= logger.ErrorLevel {
		if len(v) > 0 {
			if _, ok := v[len(v)-1].(error); ok {
				format += " %v"
			}
		}
		msg := fmt.Sprintf(format, v...)
		log.Output(2, "ERROR - [SIU] "+msg)
	}
}
