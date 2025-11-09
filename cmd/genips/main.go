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

// Command genips generates random IPv4 addresses and writes them to a file.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func main() {
	outputPath := flag.String("out", "IPs.txt", "output file to write")
	count := flag.Int("count", 0, "number of IPv4 addresses to generate")
	reportUnique := flag.Bool("unique", false, "report unique IPv4 count (slower)")
	flag.Parse()

	if *count <= 0 {
		fmt.Println("count must be greater than 0")
		os.Exit(1)
	}

	if dir := filepath.Dir(*outputPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Println("Error creating output directory:", err)
			os.Exit(1)
		}
	}

	file, err := os.Create(*outputPath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 1<<20) // 1 MiB buffer
	defer writer.Flush()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var uniqueTracker map[uint32]struct{}
	if *reportUnique {
		uniqueTracker = make(map[uint32]struct{}, min(*count, 1_000_000))
	}

	for i := 0; i < *count; i++ {
		ip := [4]byte{
			byte(rng.Intn(256)),
			byte(rng.Intn(256)),
			byte(rng.Intn(256)),
			byte(rng.Intn(256)),
		}
		if uniqueTracker != nil {
			key := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
			uniqueTracker[key] = struct{}{}
		}
		if _, err := fmt.Fprintf(writer, "%d.%d.%d.%d\n", ip[0], ip[1], ip[2], ip[3]); err != nil {
			fmt.Println("Error writing IP:", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Generated %d IPv4 addresses in %s\n", *count, *outputPath)
	if uniqueTracker != nil {
		fmt.Printf("Unique IPv4 addresses: %d\n", len(uniqueTracker))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
