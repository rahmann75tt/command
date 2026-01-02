package command

import (
	"context"
	"maps"
	"strings"
)

type envKey struct{}

// Env returns the value of the environment variable named by key.
// It first checks the context, then falls back to querying the machine.
// Returns an empty string if the variable is unset.
func Env(ctx context.Context, m Machine, key string) string {
	if env := Envs(ctx); env != nil {
		if val, ok := env[key]; ok {
			return val
		}
	}
	switch OS(ctx, m) {
	case "windows":
		out, err := probeRead(ctx, m, "powershell", "-Command",
			"Write-Output $env:"+key,
		)
		if err == nil && out != "" {
			return out
		}
		out, err = probeRead(ctx, m, "cmd", "/c", "echo %"+key+"%")
		if err == nil {
			// Check if variable exists (Windows echoes %VAR% if not set).
			if v := strings.TrimSpace(out); v != "%"+key+"%" && v != "" {
				return v
			}
		}
	default:
		if out, err := probeRead(ctx, m, "printenv", key); err == nil {
			return out
		}
	}
	return ""
}

// Envs returns a map of the environment variables stored in ctx.
func Envs(ctx context.Context) map[string]string {
	if env, ok := ctx.Value(envKey{}).(map[string]string); ok {
		return env
	}
	return nil
}

// WithEnv returns a new context with the provided environment variables
// merged with any existing environment variables in ctx.
func WithEnv(ctx context.Context, env map[string]string) context.Context {
	val := maps.Clone(Envs(ctx))
	if val == nil {
		val = make(map[string]string, len(env))
	}
	maps.Copy(val, env)
	return context.WithValue(ctx, envKey{}, val)
}

// WithoutEnv returns a new context with all environment variables removed.
// This is similar to context.WithoutCancel - it preserves all other values
// in the context (working directory, deadlines, etc.) while clearing only
// the environment variables.
//
// This is useful when environment variables have been converted to another
// form (e.g., command-line arguments for SSH) and should not be passed
// through to the underlying Machine.
func WithoutEnv(ctx context.Context) context.Context {
	if Envs(ctx) == nil {
		return ctx
	}
	return context.WithValue(ctx, envKey{}, nil)
}

// UnsetEnv returns a new context with the named environment variable removed.
func UnsetEnv(ctx context.Context, name string) context.Context {
	env := Envs(ctx)
	if env == nil {
		return ctx
	}
	val := maps.Clone(env)
	delete(val, name)
	if len(val) < 1 {
		val = nil
	}
	return context.WithValue(ctx, envKey{}, val)
}
