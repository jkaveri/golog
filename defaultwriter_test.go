package golog

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewDefaultWriter(buf)
	assert.NotNil(t, writer, "NewDefaultWriter should not return nil")
}

func TestDefaultWriter_Write(t *testing.T) {
	tests := []struct {
		name        string
		level       int
		message     string
		fields      map[string]any
		contains    []string
		validate    func(t *testing.T, output string)
		shouldPanic bool
	}{
		{
			name:    "basic-log-entry",
			level:   LevelInfo,
			message: "test message",
			fields:  nil,
			contains: []string{
				"[INFO]",
				"test message",
				"defaultwriter_test.go",
			},
		},
		{
			name:    "log-entry-with-fields",
			level:   LevelDebug,
			message: "user action",
			fields: map[string]any{
				"user_id": 123,
				"action":  "login",
			},
			contains: []string{
				"[DEBUG]",
				"user action",
				`user_id="123"`,
				`action="login"`,
			},
		},
		{
			name:    "verify-timestamp-format",
			level:   LevelInfo,
			message: "timestamp test",
			fields:  nil,
			validate: func(t *testing.T, output string) {
				// Extract timestamp from the log entry
				// Format: file:line [level][timestamp] message
				parts := strings.Split(output, "]")
				assert.Greater(t, len(parts), 2, "Log entry should contain timestamp")

				// The timestamp is in the second part, between [ and ]
				timestamp := strings.TrimPrefix(parts[1], "[")
				_, err := time.Parse(time.RFC3339, timestamp)
				assert.NoError(t, err, "Timestamp should be in RFC3339 format log=%s", output)
			},
		},
		{
			name:    "complex-field-types",
			level:   LevelInfo,
			message: "complex types",
			fields: map[string]any{
				"bool":  true,
				"float": 3.14,
				"int":   42,
				"time":  time.Now(),
			},
			contains: []string{
				`bool="true"`,
				`float="3.14"`,
				`int="42"`,
			},
		},
		{
			name:    "panic-on-complex64",
			level:   LevelInfo,
			message: "test complex64",
			fields: map[string]any{
				"complex64": complex64(1 + 2i),
			},
			shouldPanic: true,
		},
		{
			name:    "panic-on-complex128",
			level:   LevelInfo,
			message: "test complex128",
			fields: map[string]any{
				"complex128": complex128(3 + 4i),
			},
			shouldPanic: true,
		},
		{
			name:    "panic-on-unsupported-type",
			level:   LevelInfo,
			message: "test unsupported type",
			fields: map[string]any{
				"channel": make(chan int),
			},
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			writer := NewDefaultWriter(buf)

			if tt.shouldPanic {
				assert.Panics(t, func() {
					writer.Write(tt.level, tt.message, tt.fields)
				}, "Write should panic for unsupported types")
				return
			}

			writer.Write(tt.level, tt.message, tt.fields)
			writer.Flush()

			output := buf.String()

			// Check for required string contents
			for _, contain := range tt.contains {
				assert.Contains(t, output, contain)
			}

			// Run additional validation if specified
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestDefaultWriter_Flush(t *testing.T) {
	tests := []struct {
		name    string
		level   int
		message string
		fields  map[string]any
	}{
		{
			name:    "flush-after-write",
			level:   LevelInfo,
			message: "test message",
			fields:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			writer := NewDefaultWriter(buf)

			writer.Write(tt.level, tt.message, tt.fields)

			assert.NotPanics(t, func() {
				writer.Flush()
			}, "Flush should not panic")
		})
	}
}

func TestDefaultWriter_FieldsToString(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]any
		expected string
	}{
		{
			name:     "empty-fields",
			fields:   nil,
			expected: "",
		},
		{
			name: "single-field",
			fields: map[string]any{
				"key": "value",
			},
			expected: `key="value"`,
		},
		{
			name: "multiple-fields",
			fields: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			expected: `key1="value1" key2="value2"`,
		},
		{
			name: "numeric-fields",
			fields: map[string]any{
				"int":   42,
				"float": 3.14,
				"bool":  true,
			},
			expected: `int="42" float="3.14" bool="true"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewDefaultWriter(&bytes.Buffer{})
			result := writer.fieldsToString(tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}
