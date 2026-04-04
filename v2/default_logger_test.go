package golog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func restoreDefaultLogger(t *testing.T) {
	prev := defaultLogger
	t.Cleanup(func() {
		defaultLogger = prev
	})
}

func TestInitDefaultAndDefault(t *testing.T) {
	restoreDefaultLogger(t)
	err := InitDefault(Config{Format: FormatText, Output: "", Level: LevelDebug})
	require.NoError(t, err)
	require.NotNil(t, Default())
}

func TestTopLevelLevelMethodsDelegateToDefault(t *testing.T) {
	restoreDefaultLogger(t)
	w := &testWriter{}
	defaultLogger = NewLoggerWriter(w, LevelDebug)

	Debug("d")
	require.Equal(t, LevelDebug, w.lastLevel)
	require.Equal(t, "d", w.lastMessage)

	Info("i")
	require.Equal(t, LevelInfo, w.lastLevel)
	require.Equal(t, "i", w.lastMessage)

	Error("e")
	require.Equal(t, LevelError, w.lastLevel)
	require.Equal(t, "e", w.lastMessage)
}

func TestTopLevelWithAndWithErrorAndWithContext(t *testing.T) {
	restoreDefaultLogger(t)
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, 1)

	w := &testWriter{captureAttrs: true}
	defaultLogger = NewLoggerWriter(w, LevelDebug)

	With(String("scope", "s")).WithError(context.Canceled).WithContext(ctx).Info("msg", String("call", "c"))

	want := []string{"scope", "error", "call"}
	require.Equal(t, want, w.lastAttrKeys)
}

func TestSetLevel_packageDefault(t *testing.T) {
	restoreDefaultLogger(t)
	w := &testWriter{}
	defaultLogger = NewLoggerWriter(w, LevelInfo)
	Debug("skip")
	require.Equal(t, 0, w.writeCount)
	SetLevel(LevelDebug)
	Debug("x")
	require.Equal(t, 1, w.writeCount)
	require.Equal(t, LevelDebug, w.lastLevel)
}

func TestInitDefault_discard(t *testing.T) {
	restoreDefaultLogger(t)
	err := InitDefault(Config{Format: FormatText, Output: "", Level: LevelDebug})
	require.NoError(t, err)
	Info("smoke", String("k", "v"))
}

func TestInitDefault_invalidFormat(t *testing.T) {
	restoreDefaultLogger(t)
	err := InitDefault(Config{Format: "yaml", Output: ""})
	require.Error(t, err)
}

func TestDefault_fallbackWhenLoggerUnset(t *testing.T) {
	restoreDefaultLogger(t)
	defaultLogger = nil

	log := Default()
	require.NotNil(t, log)
	log.Info("smoke")
}

func TestTopLevelWithContext(t *testing.T) {
	restoreDefaultLogger(t)
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "ctx-val")

	w := &testWriter{}
	defaultLogger = NewLoggerWriter(w, LevelDebug)

	WithContext(ctx).Info("m")
	require.Equal(t, 1, w.writeCount)
	require.Equal(t, "m", w.lastMessage)
}

func TestTopLevelWithError(t *testing.T) {
	restoreDefaultLogger(t)
	w := &testWriter{captureAttrs: true}
	defaultLogger = NewLoggerWriter(w, LevelDebug)

	WithError(context.Canceled).Info("e")
	require.Contains(t, w.lastAttrKeys, "error")
}
