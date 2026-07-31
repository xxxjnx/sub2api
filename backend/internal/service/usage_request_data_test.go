package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithUsageRequestData_PreservesExactImmutableBytes(t *testing.T) {
	t.Parallel()

	raw := []byte("{\n  \"authorization\": \"Bearer raw-secret\"\n}")
	ctx := WithUsageRequestData(context.Background(), raw, " application/json ")
	raw[0] = '['

	snapshot, ok := usageRequestDataFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "{\n  \"authorization\": \"Bearer raw-secret\"\n}", string(snapshot.data))
	require.Equal(t, "application/json", snapshot.contentType)
}

func TestPropagateUsageRequestData_CopiesSnapshotToWorkerContext(t *testing.T) {
	t.Parallel()

	type targetKey struct{}
	parent := WithUsageRequestData(context.Background(), []byte("raw request"), "text/plain")
	target := context.WithValue(context.Background(), targetKey{}, "kept")

	worker := PropagateUsageRequestData(parent, target)
	require.Equal(t, "kept", worker.Value(targetKey{}))

	snapshot, ok := usageRequestDataFromContext(worker)
	require.True(t, ok)
	require.Equal(t, []byte("raw request"), snapshot.data)
	require.Equal(t, "text/plain", snapshot.contentType)
}
