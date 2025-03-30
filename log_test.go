package golog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelConstants(t *testing.T) {
	assert.Equal(t, 0, LevelDebug)
	assert.Equal(t, 1, LevelInfo)
	assert.Equal(t, 2, LevelError)
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "parse debug level",
			input:    "debug",
			expected: LevelDebug,
		},
		{
			name:     "parse info level",
			input:    "INFO",
			expected: LevelInfo,
		},
		{
			name:     "parse error level",
			input:    "Error",
			expected: LevelError,
		},
		{
			name:     "invalid level",
			input:    "invalid",
			expected: -1,
		},
		{
			name:     "empty level",
			input:    "",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "debug level string",
			input:    LevelDebug,
			expected: "DEBUG",
		},
		{
			name:     "info level string",
			input:    LevelInfo,
			expected: "INFO",
		},
		{
			name:     "error level string",
			input:    LevelError,
			expected: "ERROR",
		},
		{
			name:     "invalid level",
			input:    999,
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LevelString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetMinLevel(t *testing.T) {
	// Save original minLevel
	originalMinLevel := minLevel

	// Test valid levels
	SetLevel(LevelDebug)
	assert.Equal(t, LevelDebug, minLevel)

	SetLevel(LevelInfo)
	assert.Equal(t, LevelInfo, minLevel)

	SetLevel(LevelError)
	assert.Equal(t, LevelError, minLevel)

	// Test invalid level
	SetLevel(999)
	assert.Equal(t, LevelError, minLevel) // Should not change

	// Restore original minLevel
	minLevel = originalMinLevel
}

func TestShouldLog(t *testing.T) {
	// Save original minLevel
	originalMinLevel := minLevel

	tests := []struct {
		name     string
		minLevel int
		level    int
		expected bool
	}{
		{
			name:     "debug level with debug min",
			minLevel: LevelDebug,
			level:    LevelDebug,
			expected: true,
		},
		{
			name:     "debug level with info min",
			minLevel: LevelInfo,
			level:    LevelDebug,
			expected: false,
		},
		{
			name:     "info level with debug min",
			minLevel: LevelDebug,
			level:    LevelInfo,
			expected: true,
		},
		{
			name:     "error level with info min",
			minLevel: LevelInfo,
			level:    LevelError,
			expected: true,
		},
		{
			name:     "invalid level",
			minLevel: LevelDebug,
			level:    999,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minLevel = tt.minLevel
			result := shouldLog(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Restore original minLevel
	minLevel = originalMinLevel
}
