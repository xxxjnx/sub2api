package service

import (
	"bytes"
	"context"
	"strings"
)

type usageRequestDataContextKey struct{}

type usageRequestDataSnapshot struct {
	data        []byte
	contentType string
}

// WithUsageRequestData stores an immutable copy of the original client request
// body. The bytes are intentionally kept verbatim: usage detail must not redact,
// normalize, or rewrite the payload supplied by the client.
func WithUsageRequestData(ctx context.Context, data []byte, contentType string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(data) == 0 {
		return ctx
	}
	return context.WithValue(ctx, usageRequestDataContextKey{}, &usageRequestDataSnapshot{
		data:        bytes.Clone(data),
		contentType: strings.TrimSpace(contentType),
	})
}

// PropagateUsageRequestData copies the immutable request snapshot between the
// request context and the bounded usage-record worker context.
func PropagateUsageRequestData(parent, target context.Context) context.Context {
	if target == nil {
		target = context.Background()
	}
	snapshot, ok := usageRequestDataFromContext(parent)
	if !ok {
		return target
	}
	return context.WithValue(target, usageRequestDataContextKey{}, snapshot)
}

func usageRequestDataFromContext(ctx context.Context) (*usageRequestDataSnapshot, bool) {
	if ctx == nil {
		return nil, false
	}
	snapshot, ok := ctx.Value(usageRequestDataContextKey{}).(*usageRequestDataSnapshot)
	return snapshot, ok && snapshot != nil && len(snapshot.data) > 0
}
