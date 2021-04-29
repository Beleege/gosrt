package mpegts

import (
	"io"
	"os"
	"testing"
)

func TestExtractPCR(t *testing.T) {
	f, _ := os.Open("/tmp/test.ts")
	buf := make([]byte, TSPackageSize)
	first := 0.0
	last := 0.0
	for {
		if n, err := io.ReadFull(f, buf); err == nil {
			if n == TSPackageSize {
				if d, ok := ExtractPCR(buf[:n]); ok {
					if first == 0.0 {
						first = d
					} else {
						last = d
					}
				}
			}
		} else {
			break
		}
	}
	t.Logf("duration is %f", last-first)
}
