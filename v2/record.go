package golog

import "time"

// Record is one immutable log event produced by a [Logger] and passed to [Writer.Write].
// Call sites supply the message and attributes; the logger sets Time and Level.
// Inspect attributes with [Record.NumAttrs], [Record.Attr], or [Record.RangeAttrs].
type Record struct {
	Time    time.Time
	Level   Level
	Message string
	attrs   []Attr
}

// RecordBuilder accumulates attributes for a single log line. [Enricher] implementations
// receive a pointer to a builder and call [RecordBuilder.AddAttr] before [RecordBuilder.Build]
// freezes the result into a [Record].
type RecordBuilder struct {
	time    time.Time
	level   Level
	message string
	attrs   []Attr
}

func newRecordBuilder(ts time.Time, level Level, msg string, attrCap int) RecordBuilder {
	b := RecordBuilder{
		time:    ts,
		level:   level,
		message: msg,
	}
	if attrCap > 0 {
		b.attrs = make([]Attr, 0, attrCap)
	}
	return b
}

// AddAttr appends one attribute to the builder in addition to any call-site attributes.
func (b *RecordBuilder) AddAttr(a Attr) {
	b.attrs = append(b.attrs, a)
}

// AddAttrs appends multiple attributes; order is preserved in the final [Record].
func (b *RecordBuilder) AddAttrs(attrs ...Attr) {
	b.attrs = append(b.attrs, attrs...)
}

// Build returns an immutable [Record] with the builder’s time, level, message, and attrs.
func (b RecordBuilder) Build() Record {
	out := Record{
		Time:    b.time,
		Level:   b.level,
		Message: b.message,
		attrs:   b.attrs,
	}
	return out
}

// NumAttrs returns the number of attributes attached to the record.
func (r Record) NumAttrs() int {
	return len(r.attrs)
}

// Attr returns the i-th attribute in emission order (0 <= i < [Record.NumAttrs]).
// It panics if i is out of range.
func (r Record) Attr(i int) Attr {
	if i < 0 || i >= r.NumAttrs() {
		panic("golog: attr index out of range")
	}
	return r.attrs[i]
}

// RangeAttrs calls yield for each attribute in order. Iteration stops early if yield returns false.
func (r Record) RangeAttrs(yield func(Attr) bool) {
	for i := 0; i < len(r.attrs); i++ {
		if !yield(r.attrs[i]) {
			return
		}
	}
}
