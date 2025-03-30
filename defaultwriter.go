package golog

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
)

// defaultWriter implements the LogWriter interface with buffered writing and efficient JSON serialization.
// It provides a default implementation for logging with file location, timestamp, and structured fields.
type defaultWriter struct {
	output io.Writer
	buf    *bufio.Writer
}

// NewDefaultWriter creates a new defaultWriter instance with the given io.Writer.
// The writer is wrapped in a buffer for better performance.
// Example:
//
//	writer := NewDefaultWriter(os.Stdout)
func NewDefaultWriter(output io.Writer) *defaultWriter {
	return &defaultWriter{
		output: output,
		buf:    bufio.NewWriter(output),
	}
}

// Write implements LogWriter interface. It writes a log entry with the following format:
//
//	file:line [level][timestamp] message field1="value1" field2="value2"
//
// The fields are automatically converted to strings and properly escaped.
// The caller information (file and line) is automatically captured.
func (l *defaultWriter) Write(level int, msg string, fields map[string]any) {
	file, line := getCallerInfo(skipFrames)
	fmt.Fprintf(
		l.buf,
		"%s [%s][%s] %s %s\n",
		fmt.Sprintf("%s:%d", file, line),
		LevelString(level),
		time.Now().Format(time.RFC3339),
		msg,
		l.fieldsToString(fields),
	)
}

// Flush writes any buffered data to the underlying writer and closes it if it implements io.Closer.
// This should be called when you want to ensure all buffered logs are written.
// It's typically called when shutting down the application or when immediate flushing is needed.
func (l *defaultWriter) Flush() {
	l.buf.Flush()
	if flusher, ok := l.output.(io.Closer); ok {
		flusher.Close()
	}
}

// fieldsToString converts a map of fields to a space-separated string of key-value pairs.
// Each value is wrapped in quotes and properly escaped.
// Example: map[string]any{"user": "john", "age": 30} -> user="john" age="30"
func (l *defaultWriter) fieldsToString(fields map[string]any) string {
	var sb strings.Builder

	started := false
	for key, value := range fields {
		if started {
			sb.WriteRune(' ')
		} else {
			started = true
		}

		sb.WriteString(key)
		sb.WriteRune('=')
		sb.WriteRune('"')
		sb.WriteString(l.valToString(value))
		sb.WriteRune('"')
	}

	return sb.String()
}

// valToString converts any value to its string representation.
// It handles various types including:
// - Basic types (string, bool, numbers)
// - Complex numbers
// - Time values (formatted as RFC3339)
// - Error types
// - Any other type (converted using JSON serialization via Sonic)
func (l *defaultWriter) valToString(value any) string {
	var sb strings.Builder

	switch v := value.(type) {
	case string:
		sb.WriteString(v)
	case bool:
		sb.WriteString(strconv.FormatBool(v))
	case float64:
		sb.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	case float32:
		sb.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	case int64:
		sb.WriteString(strconv.FormatInt(v, 10))
	case int32:
		sb.WriteString(strconv.FormatInt(int64(v), 10))
	case int:
		sb.WriteString(strconv.Itoa(v))
	case uint64:
		sb.WriteString(strconv.FormatUint(v, 10))
	case uint32:
		sb.WriteString(strconv.FormatUint(uint64(v), 10))
	case uint:
		sb.WriteString(strconv.FormatUint(uint64(v), 10))
	case uint8:
		sb.WriteString(strconv.FormatUint(uint64(v), 10))
	case uint16:
		sb.WriteString(strconv.FormatUint(uint64(v), 10))
	case complex64:
		panic("complex64 is not supported")
	case complex128:
		panic("complex128 is not supported")
	case time.Time:
		sb.WriteString(v.Format(time.RFC3339))
	case error:
		sb.WriteString(v.Error())
	default:
		sb.WriteString(l.reflectToString(v))
	}

	return sb.String()
}

// reflectToString uses Sonic to convert any value to its JSON string representation.
// This is used as a fallback for types that aren't handled by valToString.
// Sonic is used instead of the standard json package for better performance.
// Returns an empty string if serialization fails.
func (l *defaultWriter) reflectToString(v any) string {
	jstr, err := sonic.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(jstr)
}
