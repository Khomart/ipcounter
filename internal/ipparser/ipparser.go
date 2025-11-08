package ipparser

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"ipcounter/internal/ipcounter"
	"net"
	"os"
	"strings"
	"sync"
)

type IPParser struct {
	ipCounter *ipcounter.IPCounter
	fileName  string
	workers   int64
}

func New(ipCounter *ipcounter.IPCounter, fileName string, workers int64) *IPParser {
	return &IPParser{
		ipCounter: ipCounter,
		fileName:  fileName,
		workers:   workers,
	}
}

func (p *IPParser) Parse() error {
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
	fmt.Println("File size:", size, "bytes")
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

func (p *IPParser) processBatch(start int64, end int64, wg *sync.WaitGroup) {
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
			fmt.Println("Error reading line:", err)
			return
		}
		pos += int64(len(line))
		ip := net.ParseIP(strings.TrimSuffix(string(line), "\n")).To4()
		if ip == nil {
			fmt.Println("Error parsing IP:", ip)
			continue
		}
		p.ipCounter.Add(binary.BigEndian.Uint32(ip))
	}
}
