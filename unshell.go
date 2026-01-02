package command

// Unsheller is an optional interface that machines can implement to expose
// a version of themselves with fallback routing to inner machines.
//
// This is the opposite of Shell: while Shell creates explicit routing with
// "command not found" for unregistered commands, Unshell removes that
// protection and allows fallback to inner machine implementations.
//
// IMPORTANT: Unshell breaks portability guarantees. Code using Unshell may
// depend on commands available in the underlying machine that aren't
// explicitly registered. Use Unshell only when you understand the tradeoffs,
// such as for probing operations (OS, Arch, FS) where you want to query the
// actual underlying system.
//
// This is analogous to "unsafe" in memory management - it removes safety
// guarantees but enables necessary low-level operations.
type Unsheller interface {
	// Unshell returns a Machine that falls back to inner layers when
	// commands aren't found.
	//
	// If the machine cannot provide fallback behavior, or if exposing
	// internals doesn't make sense, Unshell returns nil.
	Unshell() Machine
}

// Unshell returns a version of the machine with fallback routing to inner
// layers. This is the opposite of Shell - it removes explicit routing
// protection and allows commands to fall back to underlying implementations.
//
// If m implements the Unsheller interface and Unshell() returns non-nil,
// that machine is returned. Otherwise, m itself is returned unchanged.
//
// IMPORTANT: Unshell breaks portability guarantees. While Shell ensures code
// only uses explicitly registered commands, Unshell removes this protection
// and allows fallback to underlying machine implementations. Code using
// Unshell may inadvertently depend on commands available in development but
// not in production.
//
// Use Unshell only for specific purposes like probing (OS, Arch, FS
// detection) where you intentionally want to query the underlying system.
// For regular command execution, use Shell's explicit routing.
//
// Example:
//
//	sh := command.Shell(sys.Machine())
//	sh.Handle("jq", jqMachine)
//
//	// Regular use - only "jq" works, others return "command not found"
//	sh.Read(ctx, "cat", "file")  // Error: command not found
//
//	// Unshell for probing - falls back to sys.Machine
//	m := command.Unshell(sh)
//	command.OS(ctx, m)  // Works! Falls back to sys.Machine's uname
func Unshell(m Machine) Machine {
	if u, ok := m.(Unsheller); ok {
		if unshelled := u.Unshell(); unshelled != nil {
			return unshelled
		}
	}
	return m
}
