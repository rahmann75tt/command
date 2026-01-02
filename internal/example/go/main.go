// This example demonstrates installing a Go toolchain into a machine without
// relying on any commands being present.
//
// The example shows:
//   - Probing OS/Arch to determine correct Go distribution
//   - Fetching Go toolchain on host and streaming into target via command.FS
//   - Using command.Shell to explicitly register each available command
//   - Idempotent Go installation with sync.Once
//   - Building and running Go code in the prepared environment
//
// By default, runs on the local system. Use -ctr to run in a minimal busybox
// container. We don't rely on wget, curl, or package managers - everything is
// streamed in via the filesystem from the host.
package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"lesiw.io/command"
	"lesiw.io/command/ctr"
	"lesiw.io/command/sub"
	"lesiw.io/command/sys"
	"lesiw.io/defers"
	"lesiw.io/fs/path"
)

var (
	latestGoTestHook       func() (string, error)
	httpGetTestHook        func(url string) (*http.Response, error)
	extractTarballTestHook func(context.Context, io.Writer, io.Reader) error
	extractZipTestHook     func(context.Context, io.Writer, io.Reader) error
)

func main() {
	defer defers.Run()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		defers.Exit(1)
	}
}

func run() error {
	var (
		useCtr = flag.Bool("ctr", false, "install to busybox container")
		ctx    = context.Background()
	)
	flag.Parse()

	sh := shell(ctx, *useCtr)
	if err := registerGo(ctx, sh); err != nil {
		return err
	}
	if ver, err := sh.Read(ctx, "go", "version"); err != nil {
		return fmt.Errorf("failed to retrieve installed Go version: %w", err)
	} else {
		fmt.Printf("Go installed: %s\n", ver)
	}
	if err := runProgram(ctx, sh); err != nil {
		return fmt.Errorf("failed to run Go program: %w", err)
	}
	return nil
}

func shell(ctx context.Context, useCtr bool) *command.Sh {
	var m command.Machine = sys.Machine()
	if useCtr {
		m = ctr.Machine(m, "busybox:latest")
		defers.Add(func() { _ = command.Shutdown(ctx, m) })
	}
	return command.Shell(m)
}

func registerGo(ctx context.Context, sh *command.Sh) error {
	dir, err := sh.Temp(ctx, "go/")
	if err != nil {
		return fmt.Errorf("failed to create temp dir for Go: %w", err)
	}
	defers.Add(func() { _ = sh.RemoveAll(ctx, dir.Path()) })
	var install = sync.OnceValue(func() error {
		return installGo(ctx, sh, dir)
	})
	bin := path.Join(dir.Path(), "go", "bin", "go")
	if sh.OS(ctx) == "windows" {
		bin += ".exe"
	}
	sh.HandleFunc("go",
		func(ctx context.Context, args ...string) command.Buffer {
			if err := install(); err != nil {
				return command.Fail(err)
			}
			ctx = command.WithEnv(ctx, map[string]string{
				"GOROOT":  path.Join(dir.Path(), "go"),
				"GOCACHE": path.Join(dir.Path(), "gocache"),
			})
			return sub.Machine(command.Unshell(sh), bin).Command(
				ctx, args[1:]...,
			)
		},
	)
	return nil
}

var prog = []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello from Go!")
}
`)

func runProgram(ctx context.Context, sh *command.Sh) error {
	dir, err := sh.Temp(ctx, "hello-go/")
	if err != nil {
		return fmt.Errorf("failed to create tempfile for program: %w", err)
	}
	defers.Add(func() { _ = sh.RemoveAll(ctx, dir.Path()) })

	src := path.Join(dir.Path(), "hello.go")
	if err := sh.WriteFile(ctx, src, prog); err != nil {
		return fmt.Errorf("failed to write hello.go: %w", err)
	}
	sh.HandleFunc("hello",
		func(ctx context.Context, args ...string) command.Buffer {
			return sh.Command(ctx, "go", "run", src)
		},
	)
	return sh.Exec(ctx, "hello")
}

func httpGet(url string) (*http.Response, error) {
	if h := httpGetTestHook; h != nil {
		return h(url)
	}
	return http.Get(url)
}

func installGo(ctx context.Context, sh *command.Sh, w io.WriteCloser) error {
	ver, err := latestGo()
	if err != nil {
		return err
	}
	fmt.Printf("Installing Go %s for %s/%s...\n",
		ver, sh.OS(ctx), sh.Arch(ctx),
	)
	if sh.OS(ctx) == "windows" {
		url := fmt.Sprintf(
			"https://go.dev/dl/%s.%s-%s.zip",
			ver, sh.OS(ctx), sh.Arch(ctx),
		)
		resp, err := httpGet(url)
		if err != nil {
			return fmt.Errorf("failed to download Go: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf(
				"failed to download Go: HTTP %d", resp.StatusCode,
			)
		}
		if err := extractZip(ctx, w, resp.Body); err != nil {
			return err
		}
	} else {
		url := fmt.Sprintf(
			"https://go.dev/dl/%s.%s-%s.tar.gz",
			ver, sh.OS(ctx), sh.Arch(ctx),
		)
		resp, err := httpGet(url)
		if err != nil {
			return fmt.Errorf("failed to download Go: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf(
				"failed to download Go: HTTP %d", resp.StatusCode,
			)
		}
		gzr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzr.Close()
		if err := extractTarball(ctx, w, gzr); err != nil {
			return err
		}
	}
	return nil
}

func latestGo() (string, error) {
	if h := latestGoTestHook; h != nil {
		return h()
	}

	resp, err := httpGet("https://go.dev/VERSION?m=text")
	if err != nil {
		return "", fmt.Errorf("failed to fetch Go version: %w", err)
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read version: %w", err)
	}
	ver, _, _ := strings.Cut(string(buf), "\n")
	return ver, nil
}

func extractTarball(ctx context.Context, dst io.Writer, src io.Reader) error {
	if h := extractTarballTestHook; h != nil {
		return h(ctx, dst, src)
	}

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to extract tar: %w", err)
	}
	return nil
}

func extractZip(ctx context.Context, dst io.Writer, src io.Reader) error {
	if h := extractZipTestHook; h != nil {
		return h(ctx, dst, src)
	}

	buf, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("failed to read zip data: %w", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}

	return zipToTar(dst, zr)
}

func zipToTar(dst io.Writer, zr *zip.Reader) error {
	tw := tar.NewWriter(dst)
	defer tw.Close()

	writeFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", f.Name, err)
		}
		defer rc.Close()

		header, err := tar.FileInfoHeader(f.FileInfo(), "")
		if err != nil {
			return fmt.Errorf(
				"failed to create header for %s: %w", f.Name, err,
			)
		}
		header.Name = f.Name
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write header for %s: %w", f.Name, err)
		}
		if !f.FileInfo().IsDir() {
			if _, err := io.Copy(tw, rc); err != nil {
				return fmt.Errorf("failed to copy %s: %w", f.Name, err)
			}
		}
		return nil
	}
	for _, f := range zr.File {
		if err := writeFile(f); err != nil {
			return err
		}
	}

	return nil
}
