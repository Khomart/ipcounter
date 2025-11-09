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

package storage

import (
	"math"
	"sync"
	"testing"
)

func TestStorageAddSequential(t *testing.T) {
	storage := New()

	ips := []uint32{
		0,
		1,
		1,
		63,
		64,
		math.MaxUint32,
		math.MaxUint32,
	}

	for _, ip := range ips {
		storage.Add(ip)
	}

	if got, want := storage.GetUniqueCount(), 5; got != want {
		t.Fatalf("GetUniqueCount() = %d, want %d", got, want)
	}
}

func TestStorageAddConcurrent(t *testing.T) {
	storage := New()

	const (
		firstRangeCount  = 200_000
		secondRangeCount = 200_000
		secondRangeStart = 100_000
	)

	var wg sync.WaitGroup
	start := make(chan struct{})

	wg.Add(2)

	go func() {
		defer wg.Done()
		<-start
		for ip := range uint32(firstRangeCount) {
			storage.Add(ip)
		}
	}()

	go func() {
		defer wg.Done()
		<-start
		for offset := range uint32(secondRangeCount) {
			storage.Add(secondRangeStart + offset)
		}
	}()

	close(start)
	wg.Wait()

	expectedUnique := secondRangeStart + secondRangeCount
	if got := storage.GetUniqueCount(); got != int(expectedUnique) {
		t.Fatalf("GetUniqueCount() = %d, want %d", got, expectedUnique)
	}
}

func TestStorageGetUniqueCountEmpty(t *testing.T) {
	var storage Storage

	if got := storage.GetUniqueCount(); got != 0 {
		t.Fatalf("GetUniqueCount() on zero-value storage = %d, want 0", got)
	}
}
