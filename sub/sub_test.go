package sub_test

import (
	"strings"
	"testing"

	"lesiw.io/command"
	"lesiw.io/command/mock"
	"lesiw.io/command/sub"
)

func TestMachineProbingWithPrefix(t *testing.T) {
	m := new(mock.Machine)

	// Mock the prefixed commands that sub.Machine will execute
	m.Return(strings.NewReader("Linux\n"), "ssh", "host", "uname", "-s")
	m.Return(strings.NewReader("aarch64\n"), "ssh", "host", "uname", "-m")

	sm := sub.Machine(m, "ssh", "host")

	// sub.Machine should probe with prefix
	os := command.OS(t.Context(), sm)
	if os != "linux" {
		t.Errorf("expected OS %q, got %q", "linux", os)
	}

	arch := command.Arch(t.Context(), sm)
	if arch != "arm64" {
		t.Errorf("expected Arch %q, got %q", "arm64", arch)
	}

	// FS should forward
	fs := command.FS(sm)
	if fs == nil {
		t.Error("expected non-nil FS")
	}
}
