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

package loader

import (
	"encoding/binary"
	"errors"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

func writeTempIPFile(t *testing.T, name string, lines []string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, name)

	var content []byte
	for _, line := range lines {
		content = append(content, line...)
		content = append(content, '\n')
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	return path
}

func TestDo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		lines      []string
		workers    int64
		wantAdds   []string
		wantErr    error
		createFile bool
	}{
		{
			name:       "single worker",
			workers:    1,
			createFile: true,
			lines: []string{
				"192.168.0.1",
				"10.0.0.1",
				"10.0.0.1",
			},
			wantAdds: []string{
				"192.168.0.1",
				"10.0.0.1",
				"10.0.0.1",
			},
		},
		{
			name:       "multiple workers",
			workers:    2,
			createFile: true,
			lines: []string{
				"10.0.0.1",
				"10.0.0.10",
				"10.0.0.100",
				"10.0.0.200",
				"123.123.123.123",
				"17.123.253.123",
			},
			wantAdds: []string{
				"10.0.0.1",
				"10.0.0.10",
				"10.0.0.100",
				"10.0.0.200",
				"123.123.123.123",
				"17.123.253.123",
			},
		},
		{
			name:       "handles duplicates",
			workers:    3,
			createFile: true,
			lines: []string{
				"8.8.8.8",
				"8.8.8.8",
				"1.1.1.1",
				"1.1.1.1",
				"9.9.9.9",
			},
			wantAdds: []string{
				"8.8.8.8",
				"8.8.8.8",
				"1.1.1.1",
				"1.1.1.1",
				"9.9.9.9",
			},
		},
		{
			name:       "skips invalid ips",
			workers:    2,
			createFile: true,
			lines: []string{
				"8.8.8.8",
				"invalid",
				"",
				"8.8.8.8",
				"1.1.1.1",
			},
			wantAdds: []string{
				"8.8.8.8",
				"8.8.8.8",
				"1.1.1.1",
			},
		},
		{
			name:       "empty file",
			lines:      nil,
			workers:    2,
			createFile: true,
			wantAdds:   nil,
		},
		{
			name:       "missing file",
			workers:    2,
			wantErr:    os.ErrNotExist,
			createFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			storage := newRecordingStorage()

			var path string
			if tt.createFile {
				fileName := strings.ReplaceAll(tt.name, " ", "_") + ".txt"
				path = writeTempIPFile(t, fileName, tt.lines)
			} else {
				path = filepath.Join(t.TempDir(), "missing.txt")
			}

			loader := New(storage, path, tt.workers)

			err := loader.Do()
			if tt.wantErr != nil {
				if err == nil || !errors.Is(err, tt.wantErr) {
					t.Fatalf("Parse() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			gotAdds := storage.Values()
			wantAdds := parseIPStrings(t, tt.wantAdds)

			slices.Sort(gotAdds)
			slices.Sort(wantAdds)

			for i := range wantAdds {
				if gotAdds[i] != wantAdds[i] {
					t.Fatalf("added ip[%d] = %d, want %d", i, gotAdds[i], wantAdds[i])
				}
			}
		})
	}
}

type recordingStorage struct {
	mu   sync.Mutex
	adds []uint32
}

func newRecordingStorage() *recordingStorage {
	return &recordingStorage{}
}

func (c *recordingStorage) Add(ip uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.adds = append(c.adds, ip)
}

func (c *recordingStorage) Values() []uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]uint32, len(c.adds))
	copy(cp, c.adds)
	return cp
}

func parseIPStrings(t *testing.T, ips []string) []uint32 {
	t.Helper()
	if ips == nil {
		return nil
	}

	values := make([]uint32, 0, len(ips))
	for _, s := range ips {
		ip := net.ParseIP(s).To4()
		if ip == nil {
			t.Fatalf("parseIPStrings: invalid IP %q", s)
		}
		values = append(values, binary.BigEndian.Uint32(ip))
	}
	return values
}
