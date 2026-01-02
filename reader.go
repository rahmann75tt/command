package command

import (
	"context"
	"io"
	"sync"
)

type reader struct {
	sync.Mutex
	r       io.Reader
	cancel  context.CancelFunc
	started bool
	closed  bool
}

func (r *reader) Read(p []byte) (int, error) {
	r.Lock()
	if r.closed {
		r.Unlock()
		return 0, ErrClosed
	}
	r.started = true
	r.Unlock()

	return r.r.Read(p)
}

func (r *reader) Close() error {
	r.Lock()
	if r.closed {
		r.Unlock()
		return nil
	}
	r.closed = true
	started := r.started
	r.Unlock()

	if started && r.cancel != nil {
		r.cancel()
	}

	if started {
		if closer, ok := r.r.(io.Closer); ok {
			return closer.Close()
		}
	}

	return nil
}

// NewReader creates a read-only command that cancels on Close.
//
// The command starts lazily on the first Read() call. Close() cancels
// the underlying context to immediately terminate the command, which is
// appropriate for read-only operations where the user has signaled
// they're done reading.
//
// If Close() is called before any Read(), the command never starts.
func NewReader(ctx context.Context, m Machine, args ...string) io.ReadCloser {
	ctx, cancel := context.WithCancel(ctx)
	return &reader{
		r:      m.Command(ctx, args...),
		cancel: cancel,
	}
}
