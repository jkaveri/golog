package golog

import (
	"context"
	"slices"
	"testing"
)

func restoreDefaultLogger(t *testing.T) {
	prev := defaultLogger
	t.Cleanup(func() {
		defaultLogger = prev
	})
}

func TestInitDefaultAndDefault(t *testing.T) {
	restoreDefaultLogger(t)
	if err := InitDefault(Config{Format: FormatText, Output: "", Level: LevelDebug}); err != nil {
		t.Fatal(err)
	}
	if got := Default(); got == nil {
		t.Fatal("Default() returned nil after InitDefault")
	}
}

func TestTopLevelLevelMethodsDelegateToDefault(t *testing.T) {
	restoreDefaultLogger(t)
	w := &testWriter{}
	defaultLogger = NewLoggerWriter(w, LevelDebug)

	Debug("d")
	if w.lastLevel != LevelDebug || w.lastMessage != "d" {
		t.Fatalf("Debug delegated incorrectly: level=%v msg=%q", w.lastLevel, w.lastMessage)
	}

	Info("i")
	if w.lastLevel != LevelInfo || w.lastMessage != "i" {
		t.Fatalf("Info delegated incorrectly: level=%v msg=%q", w.lastLevel, w.lastMessage)
	}

	Error("e")
	if w.lastLevel != LevelError || w.lastMessage != "e" {
		t.Fatalf("Error delegated incorrectly: level=%v msg=%q", w.lastLevel, w.lastMessage)
	}
}

func TestTopLevelWithAndWithErrorAndWithContext(t *testing.T) {
	restoreDefaultLogger(t)
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, 1)

	w := &testWriter{captureAttrs: true}
	defaultLogger = NewLoggerWriter(w, LevelDebug)

	With(String("scope", "s")).WithError(context.Canceled).WithContext(ctx).Info("msg", String("call", "c"))

	want := []string{"scope", "error", "call"}
	if !slices.Equal(w.lastAttrKeys, want) {
		t.Fatalf("attr keys: got %v, want %v", w.lastAttrKeys, want)
	}
}

func TestSetLevel_packageDefault(t *testing.T) {
	restoreDefaultLogger(t)
	w := &testWriter{}
	defaultLogger = NewLoggerWriter(w, LevelInfo)
	Debug("skip")
	if w.writeCount != 0 {
		t.Fatal("MinLevel Info should skip Debug")
	}
	SetLevel(LevelDebug)
	Debug("x")
	if w.writeCount != 1 || w.lastLevel != LevelDebug {
		t.Fatalf("after SetLevel: count=%d level=%v", w.writeCount, w.lastLevel)
	}
}

func TestInitDefault_discard(t *testing.T) {
	restoreDefaultLogger(t)
	if err := InitDefault(Config{Format: FormatText, Output: "", Level: LevelDebug}); err != nil {
		t.Fatal(err)
	}
	Info("smoke", String("k", "v"))
}

func TestInitDefault_invalidFormat(t *testing.T) {
	restoreDefaultLogger(t)
	err := InitDefault(Config{Format: "yaml", Output: ""})
	if err == nil {
		t.Fatal("want error for unknown format")
	}
}
