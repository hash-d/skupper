package frame2

import "context"

// If the given context is not nil, return it.
//
// Otherwise, return a default context.
//
// For now, that's a brand new context.Background(), but that might change
func ContextOrDefault(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
