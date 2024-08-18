package counter

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

// ByteCounter counts and displays the number of bytes downloaded.
type ByteCounter struct {
	totalBytes      uint64
	displayProgress bool
}

// New returns a new ByteCounter instance.
func New(displayProgress bool) *ByteCounter {
	return &ByteCounter{
		displayProgress: displayProgress,
	}
}

// Write implements the io.Writer interface for ByteCounter.
func (bc *ByteCounter) Write(buf []byte) (int, error) {
	n := len(buf)
	bc.totalBytes += uint64(n)

	if bc.displayProgress {
		fmt.Print("\r                                        ")
		fmt.Printf("\r            %s downloaded", humanize.Bytes(bc.totalBytes))
	}

	return n, nil
}

// TotalBytes returns the number of bytes downloaded.
func (bc *ByteCounter) TotalBytes() uint64 {
	return bc.totalBytes
}
