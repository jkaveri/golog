package golog

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func BenchmarkJSONWriterVsSlog_FlatAttrs(b *testing.B) {
	gologLogger := NewLoggerWriter(NewJSONWriter(io.Discard), LevelDebug)
	gologAttrs := []Attr{
		String("k1", "v1"),
		Int("n", 42),
		Bool("ok", true),
	}

	slogLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	slogAttrs := []slog.Attr{
		slog.String("k1", "v1"),
		slog.Int("n", 42),
		slog.Bool("ok", true),
	}
	ctx := context.Background()

	b.Run("golog", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gologLogger.Info("bench", gologAttrs...)
		}
	})

	b.Run("slog", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slogLogger.LogAttrs(ctx, slog.LevelInfo, "bench", slogAttrs...)
		}
	})
}

func BenchmarkJSONWriterVsSlog_GroupAttrs(b *testing.B) {
	gologLogger := NewLoggerWriter(NewJSONWriter(io.Discard), LevelDebug)
	gologAttrs := []Attr{
		Group("request",
			String("id", "r1"),
			Group("user", Int("id", 42)),
		),
	}

	slogLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	slogAttrs := []slog.Attr{
		slog.Group("request",
			slog.String("id", "r1"),
			slog.Group("user", slog.Int("id", 42)),
		),
	}
	ctx := context.Background()

	b.Run("golog", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gologLogger.Info("bench", gologAttrs...)
		}
	})

	b.Run("slog", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slogLogger.LogAttrs(ctx, slog.LevelInfo, "bench", slogAttrs...)
		}
	})
}
