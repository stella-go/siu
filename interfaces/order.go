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

package interfaces

type Order interface {
	Order() int
}

type OrderSlice[e Order] []e

func (p OrderSlice[e]) Len() int {
	return len(p)
}

func (p OrderSlice[e]) Less(i, j int) bool {
	return p[i].Order() < p[j].Order()
}

func (p OrderSlice[e]) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
