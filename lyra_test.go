package lyra

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	l := New()
	require.NotNil(t, l)
	require.NotNil(t, l.tasks)
	require.Len(t, l.tasks, 0)
	require.Nil(t, l.error)
}
