# IP Counter

## Problem
Process a text file containing IPv4 addresses (one per line) and report the count of unique addresses. The solution must handle very large inputs efficiently, tolerate malformed lines, and make good use of multicore machines.

## Solution Overview
- Streams the input file in parallel batches using worker goroutines.
- Normalizes addresses with the standard library and maps them into a bitmap to deduplicate values in constant time.
- Reports total unique IPv4 addresses along with timing and file size diagnostics.

### Tradeoffs
- **Bitmap memory footprint:** Storing the full IPv4 address space guarantees O(1) lookups but costs ~512 Mb even for small inputs. A hash-based structure would reduce memory but introduce collisions or higher CPU overhead. 

## Execution examples
```
$ ./bin/ipcounter localtestdata/random_10millions_ips.txt
Number of unique addresses: 9988323
Time taken: 565.960667ms

$ ./bin/ipcounter -workers 1 localtestdata/random_10millions_ips.txt
Number of unique addresses: 9988323
Time taken: 2.369079459s

$ ./bin/ipcounter localtestdata/random_1billion_ips.txt
Number of unique addresses: 892115405
Time taken: 36.022614875s

$ ./bin/ipcounter localtestdata/ip_addresses
Number of unique addresses: 1000000000
Time taken: 3m1.751277208s
```

## Requirements
- Go `1.25.1` or newer (see `go.mod`).
- 512 Mb RAM.

## Usage
1. Build the binary:
   ```
   go build -o bin/ipcounter ./cmd/ipcounter
   ```
2. Run against a file:
   ```
   ./bin/ipcounter -workers 8 ./path/to/addresses.txt
   ```
   Flags:
   - `-workers`: number of goroutines to process the file (defaults to logical CPU count).

3. Alternatively, use the `Makefile` targets described below.

## Makefile Targets

| Target     | Description                                | Command Triggered                       |
|------------|--------------------------------------------|-----------------------------------------|
| `build`    | Build the `ipcounter` binary into `bin/`.  | `go build -o bin/ipcounter ./cmd/ipcounter` |
| `run`      | Run the app without creating a binary.     | `go run ./cmd/ipcounter`                |
| `test`     | Run unit and integration tests.            | `go test ./...`                         |
| `test-race`| Run tests with the race detector enabled.  | `go test -race ./...`                   |
| `bench`    | Execute benchmarks defined in tests.       | `go test -bench=. ./...`                |
| `test-all` | Run `test`, `test-race`, then `bench`.     | invocations above in sequence           |
| `fmt`      | Format Go code under `cmd/ipcounter` and `internal`. | `go fmt ./cmd/ipcounter ./internal/...` |
| `clean`    | Remove build artifacts under `bin/`.       | `rm -rf bin`                            |

## Potential Improvements
- **Error handling:** Malformed lines are logged and skipped to ensure progress. Strict failure-on-error would surface issues faster but halt processing of large datasets.
- **Counter:** We can also use threading to count unique ips, although the result of this improvement can be negligible in overall performance.

## 📄 License

Copyright [2025] [Artem Khomich]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.