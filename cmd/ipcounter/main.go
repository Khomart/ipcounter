package main

import (
	"fmt"
	"ipcounter/internal/ipcounter"
	"ipcounter/internal/ipparser"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ipcounter <file>")
		os.Exit(1)
	}

	fileName := os.Args[1]
	timeStart := time.Now()
	counter := ipcounter.New()
	parser := ipparser.New(counter, fileName, 20)
	err := parser.Parse()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println("Unique IPs:", counter.GetUniqueCount())
	timeEnd := time.Now()
	fmt.Println("Time taken:", timeEnd.Sub(timeStart))
}
