package command

import (
	"bufio"
	"io"
)

// crlfReader wraps a ReadCloser to normalize line endings on read.
//
// Reads convert line endings to LF (\n):
//   - CRLF (\r\n) → LF (\n)  // Windows
//   - CR (\r) → LF (\n)      // Classic Mac OS
func crlfReader(r io.ReadCloser) io.ReadCloser {
	return &crlfReadCloser{r, bufio.NewReader(r), false}
}

type crlfReadCloser struct {
	io.ReadCloser
	br *bufio.Reader
	cr bool
}

func (c *crlfReadCloser) Read(p []byte) (n int, err error) {
	var ch byte
	for n < len(p) {
		ch, err = c.br.ReadByte()
		if err != nil {
			if err == io.EOF && c.cr == true {
				p[n] = '\n'
				n++
				c.cr = false
			}
			break
		}
		switch c.cr {
		case false:
			if ch == '\r' {
				c.cr = true
				continue
			}
			p[n] = ch
			n++
		case true:
			if ch == '\n' {
				p[n] = '\n'
				n++
				c.cr = false
				continue
			}
			if err := c.br.UnreadByte(); err != nil {
				return n, err
			}
			p[n] = '\n'
			n++
			c.cr = false
		}
	}
	return
}
