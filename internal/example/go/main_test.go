package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"lesiw.io/command"
	"lesiw.io/command/mock"
)

// swap swaps a variable's value and restores it via t.Cleanup.
// See https://lesiw.dev/go/testing
func swap[T any](t *testing.T, ptr *T, val T) {
	t.Helper()
	old := *ptr
	*ptr = val
	t.Cleanup(func() { *ptr = old })
}

func TestFetchLatestVersion(t *testing.T) {
	swap(t, &httpGetTestHook, func(url string) (*http.Response, error) {
		if !strings.Contains(url, "VERSION") {
			t.Errorf(
				"fetchLatestVersion() called httpGet(%q), want %q in URL",
				url, "VERSION",
			)
		}
		body := "go1.22.0\ngo1.21.5\n"
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})

	version, err := latestGo()
	if err != nil {
		t.Fatal(err)
	}

	if version != "go1.22.0" {
		t.Errorf("version = %q, want %q", version, "go1.22.0")
	}
}

func TestInstallGoLinux(t *testing.T) {
	ctx := t.Context()

	var downloadedURL string
	swap(t, &latestGoTestHook, func() (string, error) {
		return "go1.21.0", nil
	})
	swap(t, &httpGetTestHook, func(url string) (*http.Response, error) {
		downloadedURL = url
		// Return gzipped data since installGo creates a gzip reader
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		_, _ = gzw.Write([]byte("dummy tar data"))
		_ = gzw.Close()
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(&buf),
		}, nil
	})

	var extractCalled bool
	swap(t, &extractTarballTestHook, func(
		context.Context, io.Writer, io.Reader,
	) error {
		extractCalled = true
		return nil
	})

	m := new(mock.Machine)
	m.SetOS("linux")
	m.SetArch("amd64")

	sh := command.Shell(m)
	dir, err := sh.Create(ctx, "/test/")
	if err != nil {
		t.Fatal(err)
	}
	err = installGo(ctx, sh, dir)
	if err != nil {
		t.Fatal(err)
	}

	expectedURL := "https://go.dev/dl/go1.21.0.linux-amd64.tar.gz"
	if downloadedURL != expectedURL {
		t.Errorf("downloaded URL = %q, want %q", downloadedURL, expectedURL)
	}

	if !extractCalled {
		t.Error("installGo() did not call extractTarballTestHook")
	}
}

func TestInstallGoWindows(t *testing.T) {
	ctx := t.Context()

	var downloadedURL string
	swap(t, &latestGoTestHook, func() (string, error) {
		return "go1.21.0", nil
	})
	swap(t, &httpGetTestHook, func(url string) (*http.Response, error) {
		downloadedURL = url
		// Return minimal zip data since installGo creates a zip reader
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		_ = zw.Close()
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(&buf),
		}, nil
	})

	var extractCalled bool
	swap(t, &extractZipTestHook, func(
		context.Context, io.Writer, io.Reader,
	) error {
		extractCalled = true
		return nil
	})

	m := new(mock.Machine)
	m.SetOS("windows")
	m.SetArch("arm64")

	sh := command.Shell(m)
	dir, err := sh.Create(ctx, "/test/")
	if err != nil {
		t.Fatal(err)
	}
	err = installGo(ctx, sh, dir)
	if err != nil {
		t.Fatal(err)
	}

	expectedURL := "https://go.dev/dl/go1.21.0.windows-arm64.zip"
	if downloadedURL != expectedURL {
		t.Errorf("downloaded URL = %q, want %q", downloadedURL, expectedURL)
	}

	if !extractCalled {
		t.Error("installGo() did not call extractZipTestHook")
	}
}

func TestExtractTarball(t *testing.T) {
	ctx := t.Context()

	sh := command.Shell(new(mock.Machine))
	workDir := "/tmp/test"
	dir, err := sh.Create(ctx, workDir+"/")
	if err != nil {
		t.Fatal(err)
	}

	tarGzData := createMockGoTarballGz(t)

	gzr, err := gzip.NewReader(bytes.NewReader(tarGzData))
	if err != nil {
		t.Fatal(err)
	}
	defer gzr.Close()

	err = extractTarball(ctx, dir, gzr)
	if err != nil {
		t.Fatal(err)
	}

	_, err = sh.Stat(ctx, workDir+"/go/bin/go")
	if err != nil {
		t.Errorf("go binary not found: %v", err)
	}
}

func TestRegisterGo(t *testing.T) {
	ctx := t.Context()

	swap(t, &latestGoTestHook, func() (string, error) {
		return "go1.21.0", nil
	})
	swap(t, &httpGetTestHook, func(url string) (*http.Response, error) {
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		_, _ = gzw.Write([]byte("dummy tar data"))
		_ = gzw.Close()
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(&buf),
		}, nil
	})
	swap(t, &extractTarballTestHook, func(
		context.Context, io.Writer, io.Reader,
	) error {
		return nil
	})

	sh := command.Shell(new(mock.Machine))
	err := registerGo(ctx, sh)
	if err != nil {
		t.Fatal(err)
	}

	// The actual lazy installation and Once behavior is tested
	// through the full integration test in the example itself.
	// Unit testing the Once behavior would require mocking
	// sub.Machine, which couples the test too tightly to
	// implementation details.
}

func createMockGoTarballGz(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	header := &tar.Header{
		Name: "go/bin/go",
		Mode: 0755,
		Size: int64(len("#!/bin/sh\necho go version go1.21.0\n")),
	}

	err := tw.WriteHeader(header)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tw.Write([]byte("#!/bin/sh\necho go version go1.21.0\n"))
	if err != nil {
		t.Fatal(err)
	}

	err = tw.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = gzw.Close()
	if err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}
