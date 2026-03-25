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

package n

import (
	"fmt"
	"strconv"
	"time"
)

// scanInt64 converts common database driver value types to int64.
func scanInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	case string:
		return strconv.ParseInt(v, 10, 64)
	case time.Time:
		return v.Unix(), nil
	default:
		return 0, fmt.Errorf("can't convert type %T to int64", value)
	}
}

// scanUint64 converts common database driver value types to uint64.
func scanUint64(value any) (uint64, error) {
	switch v := value.(type) {
	case int64:
		return uint64(v), nil
	case int32:
		return uint64(v), nil
	case int16:
		return uint64(v), nil
	case int8:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case uint64:
		return v, nil
	case uint32:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case float64:
		return uint64(v), nil
	case float32:
		return uint64(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case []byte:
		return strconv.ParseUint(string(v), 10, 64)
	case string:
		return strconv.ParseUint(v, 10, 64)
	case time.Time:
		return uint64(v.Unix()), nil
	default:
		return 0, fmt.Errorf("can't convert type %T to uint64", value)
	}
}

// scanFloat64 converts common database driver value types to float64.
func scanFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case []byte:
		return strconv.ParseFloat(string(v), 64)
	case string:
		return strconv.ParseFloat(v, 64)
	case time.Time:
		return float64(v.Unix()), nil
	default:
		return 0, fmt.Errorf("can't convert type %T to float64", value)
	}
}

// scanBool converts common database driver value types to bool.
func scanBool(value any) (bool, error) {
	switch v := value.(type) {
	case int64:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int16:
		return v != 0, nil
	case int8:
		return v != 0, nil
	case int:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint16:
		return v != 0, nil
	case uint8:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case float32:
		return v != 0, nil
	case bool:
		return v, nil
	case []byte:
		return strconv.ParseBool(string(v))
	case string:
		return strconv.ParseBool(v)
	case time.Time:
		return !v.IsZero(), nil
	default:
		return false, fmt.Errorf("can't convert type %T to bool", value)
	}
}

// scanString converts common database driver value types to string.
func scanString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case bool:
		return strconv.FormatBool(v), nil
	case time.Time:
		return v.Format(time.DateTime), nil
	default:
		return "", fmt.Errorf("can't convert type %T to string", value)
	}
}
