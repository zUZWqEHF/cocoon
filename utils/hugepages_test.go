//go:build !linux

package utils

import (
	"testing"
)

func TestDetectHugePages_NonLinux(t *testing.T) {
	if DetectHugePages() {
		t.Error("expected false on non-Linux platform")
	}
}
