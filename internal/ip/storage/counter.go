// Copyright [2025] [Artem Khomich]

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package storage provides a concurrency-safe bitmap for tracking unique IPv4
// addresses.
package storage

import (
	"math/bits"
	"sync/atomic"
)

// Storage holds a bitmap that records which IPv4 addresses have been seen.
type Storage struct {
	bitmap []atomic.Uint64
}

// New allocates and returns a Storage ready to record IPv4 addresses.
func New() *Storage {
	return &Storage{
		bitmap: make([]atomic.Uint64, 1<<32/64),
	}
}

// Add marks the provided IPv4 address as seen.
func (h *Storage) Add(ip uint32) {
	idx := ip / 64
	mask := uint64(1) << (ip % 64)

	for {
		cur := h.bitmap[idx].Load()
		if cur&mask != 0 {
			return
		}
		if h.bitmap[idx].CompareAndSwap(cur, cur|mask) {
			return
		}
	}
}

// GetUniqueCount returns the number of distinct IPv4 addresses recorded.
func (h *Storage) GetUniqueCount() int {
	count := 0
	for i := range h.bitmap {
		count += bits.OnesCount64(h.bitmap[i].Load())
	}
	return count
}
