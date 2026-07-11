package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type contextKey struct{}

var key contextKey

func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(key).(string); ok {
		return id
	}
	return ""
}

func NewContext(ctx context.Context) context.Context {
	if id := FromContext(ctx); id != "" {
		return ctx
	}
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return context.WithValue(ctx, key, "0000000000000000")
	}
	return context.WithValue(ctx, key, hex.EncodeToString(b))
}