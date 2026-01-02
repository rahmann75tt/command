package command

import "fmt"

// String returns the string representation of buf.
// If buf implements [fmt.Stringer], returns the result of String().
// Otherwise, returns the type in angle brackets (e.g., "<*pkg.Type>").
func String(buf Buffer) string {
	if s, ok := buf.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprintf("<%T>", buf)
}
