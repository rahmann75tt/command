package command

// Fail returns a Buffer that returns err on all Read operations.
func Fail(err error) Buffer { return fail{err} }

type fail struct{ error }

func (f fail) Read([]byte) (int, error) { return 0, f.error }
