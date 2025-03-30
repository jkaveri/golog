package golog

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewJSONWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewJSONWriter(buf)
	assert.NotNil(t, writer, "NewJSONWriter should not return nil")
}

func TestJSONWriter_Write(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		message  string
		fields   map[string]any
		validate func(t *testing.T, output string)
	}{
		{
			name:    "basic-log-entry",
			level:   LevelInfo,
			message: "test message",
			fields:  nil,
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Equal(t, "test message", entry[FieldMessage])
				assert.Equal(t, "INFO", entry[FieldLevel])
				assert.Contains(t, entry, FieldTime)
				assert.Contains(t, entry, FieldCaller)
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
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Equal(t, "user action", entry[FieldMessage])
				assert.Equal(t, "DEBUG", entry[FieldLevel])
				assert.Equal(t, float64(123), entry["user_id"])
				assert.Equal(t, "login", entry["action"])
			},
		},
		{
			name:    "verify-timestamp-format",
			level:   LevelInfo,
			message: "timestamp test",
			fields:  nil,
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")

				timestamp, ok := entry[FieldTime].(string)
				assert.True(t, ok, "Timestamp should be a string")
				_, err = time.Parse(time.RFC3339, timestamp)
				assert.NoError(t, err, "Timestamp should be in RFC3339 format")
			},
		},
		{
			name:    "all-log-levels",
			level:   LevelDebug,
			message: "test all levels",
			fields:  nil,
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Equal(t, "DEBUG", entry[FieldLevel])
			},
		},
		{
			name:    "special-characters",
			level:   LevelInfo,
			message: "test with special chars: \n\t\"'{}[]",
			fields: map[string]any{
				"special_field": "value with \n\t\"'{}[]",
			},
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Equal(t, "test with special chars: \n\t\"'{}[]", entry[FieldMessage])
				assert.Equal(t, "value with \n\t\"'{}[]", entry["special_field"])
			},
		},
		{
			name:    "nested-fields",
			level:   LevelInfo,
			message: "test nested fields",
			fields: map[string]any{
				"user": map[string]any{
					"name": "John Doe",
					"address": map[string]any{
						"city": "New York",
					},
				},
			},
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				user, ok := entry["user"].(map[string]any)
				assert.True(t, ok, "User should be a map")
				assert.Equal(t, "John Doe", user["name"])
				address, ok := user["address"].(map[string]any)
				assert.True(t, ok, "Address should be a map")
				assert.Equal(t, "New York", address["city"])
			},
		},
		{
			name:    "empty-message",
			level:   LevelInfo,
			message: "",
			fields: map[string]any{
				"empty": true,
			},
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Equal(t, "", entry[FieldMessage])
				assert.Equal(t, true, entry["empty"])
			},
		},
		{
			name:    "large-field-values",
			level:   LevelInfo,
			message: "test large values",
			fields: map[string]any{
				"large_string": strings.Repeat("x", 1000),
				"large_number": 999999999999999,
			},
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Equal(t, float64(999999999999999), entry["large_number"])
				assert.Len(t, entry["large_string"].(string), 1000)
			},
		},
		{
			name:    "field-name-validation",
			level:   LevelInfo,
			message: "test field names",
			fields: map[string]any{
				"valid_field":       "value",
				"field.with.dots":   "value",
				"field-with-dashes": "value",
			},
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Equal(t, "value", entry["valid_field"])
				assert.Equal(t, "value", entry["field.with.dots"])
				assert.Equal(t, "value", entry["field-with-dashes"])
			},
		},
		{
			name:    "channel-field",
			level:   LevelInfo,
			message: "test channel",
			fields: map[string]any{
				"channel": make(chan int),
			},
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Contains(t, entry["error"], "failed to marshal log entry")
			},
		},
		{
			name:    "complex64-not-supported",
			level:   LevelInfo,
			message: "test complex64",
			fields: map[string]any{
				"complex64": complex64(1 + 2i),
			},
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Contains(t, entry["error"], "failed to marshal log entry")
			},
		},
		{
			name:    "complex128-not-supported",
			level:   LevelInfo,
			message: "test complex128",
			fields: map[string]any{
				"complex128": complex128(3 + 4i),
			},
			validate: func(t *testing.T, output string) {
				var entry map[string]any
				err := json.Unmarshal([]byte(output), &entry)
				assert.NoError(t, err, "Output should be valid JSON")
				assert.Contains(t, entry["error"], "failed to marshal log entry")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			writer := NewJSONWriter(buf)

			writer.Write(tt.level, tt.message, tt.fields)
			writer.Flush()

			output := strings.TrimSpace(buf.String())
			tt.validate(t, output)
		})
	}
}

func TestJSONWriter_Flush(t *testing.T) {
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
			writer := NewJSONWriter(buf)

			writer.Write(tt.level, tt.message, tt.fields)

			assert.NotPanics(t, func() {
				writer.Flush()
			}, "Flush should not panic")
		})
	}
}
