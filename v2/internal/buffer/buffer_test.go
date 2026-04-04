package buffer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuffer_writeAndInspect(t *testing.T) {
	b := New()
	defer b.Free()

	_, err := b.Write([]byte("ab"))
	require.NoError(t, err)
	_, err = b.WriteString("cd")
	require.NoError(t, err)
	err = b.WriteByte('e')
	require.NoError(t, err)
	require.Equal(t, "abcde", b.String())
	require.Equal(t, 5, b.Len())

	b.SetLen(2)
	require.Equal(t, "ab", b.String())

	b.Reset()
	require.Equal(t, 0, b.Len())
}

func TestBuffer_Free(t *testing.T) {
	type Args struct {
		payloadLen int
	}

	testCases := []struct {
		name string
		args Args
	}{
		{
			name: "small-buffer-pooled",
			args: Args{payloadLen: 16},
		},
		{
			name: "large-buffer-not-pooled",
			args: Args{payloadLen: 17 << 10},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := New()
			*b = append(*b, bytes.Repeat([]byte("x"), tc.args.payloadLen)...)
			b.Free()
		})
	}
}
