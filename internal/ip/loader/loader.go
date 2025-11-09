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

// Package loader provides facilities for reading IPv4 address files and
// streaming them into a storage backend.
package loader

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

type storage interface {
	Add(ip uint32)
}

// Loader reads IPv4 addresses from a file and stores them using the provided
// storage backend.
type Loader struct {
	storage  storage
	fileName string
	workers  int64
}

// New creates a Loader that will read from the given file using the supplied
// storage backend and worker count.
func New(storage storage, fileName string, workers int64) *Loader {
	return &Loader{
		storage:  storage,
		fileName: fileName,
		workers:  workers,
	}
}

// Do partitions the configured input file among worker goroutines and streams
// every parsed IPv4 address into the storage backend.
func (p *Loader) Do() error {
	file, err := os.Open(p.fileName)
	if err != nil {
		err := fmt.Errorf("error opening file: %w", err)
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		err := fmt.Errorf("error getting file size: %w", err)
		return err
	}
	size := info.Size()
	batchSize := size / p.workers
	var i int64 = 0
	var wg sync.WaitGroup
	for i = 0; i < p.workers; i++ {
		start := batchSize * i
		end := start + batchSize
		if i == p.workers-1 {
			end = size
		}
		wg.Add(1)
		go p.processBatch(start, end, &wg)
	}
	wg.Wait()
	return nil
}

func (p *Loader) processBatch(start int64, end int64, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(p.fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	file.Seek(start, 0)
	reader := bufio.NewReader(file)
	if start != 0 {
		reader.ReadBytes('\n')
	}

	var pos int64 = start
	for pos < end {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				if len(line) == 0 {
					break
				}
			} else {
				fmt.Println("Error reading line:", err)
				return
			}
		}
		pos += int64(len(line))
		ip := net.ParseIP(strings.TrimSuffix(string(line), "\n")).To4()
		if ip == nil {
			fmt.Println("Error parsing IP:", ip)
			continue
		}
		p.storage.Add(binary.BigEndian.Uint32(ip))
		if err == io.EOF {
			break
		}
	}
}
