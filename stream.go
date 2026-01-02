package command

import (
	"context"
	"io"
)

// NewStream creates a bidirectional command stream with full
// Read/Write/Close access.
//
// The returned io.ReadWriteCloser provides direct access to the command's
// stdin (Write), stdout (Read), and stdin close signal (Close).
//
// If the underlying command does not support writing (is read-only), Write()
// will return an error. Close() closes stdin if supported, otherwise it is
// a no-op.
//
// NewStream is primarily useful with command.Copy for pipeline composition.
// For most use cases, prefer NewReader (read-only with cancellation) or
// NewWriter (write-only with completion wait).
func NewStream(
	ctx context.Context, m Machine, args ...string,
) io.ReadWriteCloser {
	buf := m.Command(ctx, args...)
	return &stream{buf: buf}
}

type stream struct {
	buf Buffer
}

func (s *stream) Read(p []byte) (int, error) {
	return s.buf.Read(p)
}

func (s *stream) Write(p []byte) (int, error) {
	if wb, ok := s.buf.(WriteBuffer); ok {
		return wb.Write(p)
	}
	return 0, ErrReadOnly
}

func (s *stream) Close() error {
	if wb, ok := s.buf.(WriteBuffer); ok {
		return wb.Close()
	}
	// Read-only commands - Close is a no-op
	return nil
}

// ReadFrom implements io.ReaderFrom for optimized copying that auto-closes
// stdin when the source reaches EOF.
// This allows io.Copy to automatically close stdin in pipeline stages.
func (s *stream) ReadFrom(src io.Reader) (n int64, err error) {
	wb, ok := s.buf.(WriteBuffer)
	if !ok {
		return 0, ErrReadOnly
	}

	// Copy from source to command stdin
	n, err = io.Copy(wb, src)

	// Auto-close stdin after copy completes
	closeErr := wb.Close()
	if err == nil {
		err = closeErr
	}

	return n, err
}
