package main

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

// ByteCounter counts the number of bytes downloaded.
type ByteCounter struct {
	TotalBytes uint64
}

// Write implements the io.Writer interface for ByteCounter.
func (bc *ByteCounter) Write(buf []byte) (int, error) {
	n := len(buf)
	bc.TotalBytes += uint64(n)
	bc.PrintProgress()
	return n, nil
}

// PrintProgress prints the number of bytes written so far.
func (bc ByteCounter) PrintProgress() {
	fmt.Print("\r                                        ")
	fmt.Printf("\r            %s downloaded", humanize.Bytes(bc.TotalBytes))
}
