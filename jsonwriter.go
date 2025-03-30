package golog

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/bytedance/sonic"
)

type jsonWriter struct {
	writer *bufio.Writer
	output io.Writer
}

// NewJSONWriter creates a new JSON logger that writes to the specified io.Writer
func NewJSONWriter(output io.Writer) *jsonWriter {
	return &jsonWriter{
		writer: bufio.NewWriterSize(output, defaultBufferSize),
		output: output,
	}
}

// Write implements LogWriter interface
func (l *jsonWriter) Write(level int, msg string, fields map[string]any) {
	// Get caller information (skip 2 frames to get the actual logging call)
	file, line := getCallerInfo(skipFrames)

	// Create the base log entry
	entry := map[string]any{
		FieldTime:    time.Now().Format(time.RFC3339),
		FieldLevel:   LevelString(level),
		FieldMessage: msg,
		FieldCaller:  fmt.Sprintf("%s:%d", file, line),
	}

	// Add all fields to the entry
	for k, v := range fields {
		switch v := v.(type) {
		case error:
			entry[k] = fmt.Sprintf("%+v", v)
		default:
			entry[k] = v
		}
	}

	// Marshal to JSON using sonic
	data, err := sonic.Marshal(entry)
	if err != nil {
		panic(err)
	}

	// Write the JSON entry with a newline
	data = append(data, '\n')
	l.writer.Write(data)
}

// Flush implements LogWriter interface
func (l *jsonWriter) Flush() {
	l.writer.Flush()
	if flusher, ok := l.output.(io.Closer); ok {
		flusher.Close()
	}
}
