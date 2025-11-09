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

package integration

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"ipcounter/internal/ip/loader"
	"ipcounter/internal/ip/storage"
)

func TestIPCounterEndToEnd(t *testing.T) {
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("unable to determine current file path: %v", err)
	}

	testCases := []struct {
		name       string
		inputFile  string
		workers    int
		wantUnique int
	}{
		{
			name:       "single_worker_unique_ips",
			inputFile:  "happy_path.txt",
			workers:    1,
			wantUnique: 3,
		},
		{
			name:       "two_workers_with_duplicates",
			inputFile:  "duplicate_two_lines.txt",
			workers:    2,
			wantUnique: 1,
		},
		{
			name:       "invalid_entries_are_ignored",
			inputFile:  "with_invalid.txt",
			workers:    1,
			wantUnique: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			args := []string{
				"run",
				"../cmd/ipcounter",
				"-workers=" + strconv.Itoa(tc.workers),
				"testdata/" + tc.inputFile,
			}
			cmd := exec.Command("go", args...)
			cmd.Dir = root

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go run failed: %v\nOutput:\n%s", err, output)
			}

			uniqueCount, err := parseUniqueCount(string(output))
			if err != nil {
				t.Fatalf("failed to parse unique count: %v\nOutput:\n%s", err, output)
			}

			if uniqueCount != tc.wantUnique {
				t.Fatalf("unexpected unique count: want %d, got %d\nOutput:\n%s", tc.wantUnique, uniqueCount, output)
			}
		})
	}
}

func parseUniqueCount(output string) (int, error) {
	const prefix = "Number of unique addresses:"

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if after, ok := strings.CutPrefix(line, prefix); ok {
			value := strings.TrimSpace(after)
			uniqueCount, err := strconv.Atoi(value)
			if err != nil {
				return 0, err
			}
			return uniqueCount, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return 0, errors.New("unique count not found in output")
}

func BenchmarkIPCounter5kAddressesFile(b *testing.B) {
	b.ReportAllocs()

	dataPath := "testdata/5k_addresses.txt"
	b.Logf("Using dataset: %s", dataPath)

	workerCounts := []int{1, max(2, runtime.NumCPU()/2)}

	for _, workers := range workerCounts {
		workers := workers
		b.Run(fmt.Sprintf("workers_%d", workers), func(b *testing.B) {
			for b.Loop() {
				storage := storage.New()
				loader := loader.New(storage, dataPath, int64(workers))
				if err := loader.Do(); err != nil {
					b.Fatalf("parse failed: %v", err)
				}
			}
		})
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
