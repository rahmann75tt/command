package sys_test

import (
	"runtime"
	"testing"

	"lesiw.io/command"
	"lesiw.io/command/sys"
)

// TestOS_Detect verifies that OS detection matches runtime.GOOS.
func TestOSDetect(t *testing.T) {
	got := command.OS(t.Context(), sys.Machine())
	if want := runtime.GOOS; got != want {
		t.Errorf("OS() = %q, want %q (runtime.GOOS)", got, want)
	}
}

// TestArch_Detect verifies that Arch detection matches runtime.GOARCH.
func TestArchDetect(t *testing.T) {
	got := command.Arch(t.Context(), sys.Machine())
	if want := runtime.GOARCH; got != want {
		t.Errorf("Arch() = %q, want %q (runtime.GOOS)", got, want)
	}
}
