package command

import (
	"context"
	"io"
	"sync"
)

type writer struct {
	sync.Mutex
	w       io.Writer
	read    chan error
	started bool
	closed  bool
}

func (w *writer) init() {
	w.read = make(chan error, 1)
	if r, ok := w.w.(io.Reader); ok {
		go func() {
			_, err := io.Copy(io.Discard, r)
			w.read <- err
		}()
	} else {
		go func() {
			w.read <- nil
		}()
	}
}

func (w *writer) Write(p []byte) (int, error) {
	w.Lock()
	if w.closed {
		w.Unlock()
		return 0, ErrClosed
	}
	if !w.started {
		w.started = true
		w.init()
	}
	w.Unlock()

	return w.w.Write(p)
}

func (w *writer) Close() error {
	w.Lock()
	if w.closed {
		w.Unlock()
		return nil
	}
	w.closed = true
	started := w.started
	w.Unlock()

	if !started {
		return nil
	}

	if closer, ok := w.w.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}

	return <-w.read
}

// ReadFrom implements io.ReaderFrom for optimized copying that auto-closes
// stdin when the source reaches EOF.
// This allows io.Copy(NewWriter(...), src) to work correctly without requiring
// explicit Close() after the copy.
func (w *writer) ReadFrom(src io.Reader) (n int64, err error) {
	w.Lock()
	if w.closed {
		w.Unlock()
		return 0, ErrClosed
	}
	if !w.started {
		w.started = true
		w.init()
	}
	w.Unlock()

	// Copy from source to command stdin
	n, err = io.Copy(w.w, src)

	// Auto-close stdin after copy completes (even on error)
	if closer, ok := w.w.(io.Closer); ok {
		closeErr := closer.Close()
		if err == nil {
			err = closeErr
		}
	}

	// Wait for command to complete reading
	readErr := <-w.read
	if err == nil {
		err = readErr
	}

	// Mark as closed
	w.Lock()
	w.closed = true
	w.Unlock()

	return n, err
}

// NewWriter creates a write-only command that waits for completion on Close.
//
// The command starts lazily on the first Write() call. Close() waits for
// the command to complete gracefully by closing stdin and reading any output,
// which is appropriate for write-only operations that must finish processing
// before the operation is considered complete.
//
// If Close() is called before any Write(), the command never starts.
//
// NewWriter implements io.ReaderFrom for optimized copying. When io.Copy
// detects this, it will auto-close stdin after the source reaches EOF.
func NewWriter(ctx context.Context, m Machine, args ...string) io.WriteCloser {
	buf := m.Command(ctx, args...)
	// Assert that the command supports writing
	wb, ok := buf.(WriteBuffer)
	if !ok {
		// Return a writer that will fail on first write
		return &writer{w: &readOnlyBuffer{buf}}
	}
	return &writer{w: wb}
}

type readOnlyBuffer struct {
	Buffer
}

func (r *readOnlyBuffer) Write(p []byte) (int, error) {
	return 0, ErrReadOnly
}
