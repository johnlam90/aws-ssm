package tui

import (
	"strings"
	"sync"
)

// builderPool provides a pool of strings.Builder instances to reduce allocations
// during rendering. Each render cycle can acquire a builder, use it, then release
// it back to the pool for reuse.
var builderPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// AcquireBuilder gets a strings.Builder from the pool.
// The builder is reset and ready to use.
func AcquireBuilder() *strings.Builder {
	b := builderPool.Get().(*strings.Builder)
	b.Reset()
	return b
}

// ReleaseBuilder returns a strings.Builder to the pool for reuse.
// After calling this, the builder should not be used.
func ReleaseBuilder(b *strings.Builder) {
	if b == nil {
		return
	}
	// Only pool builders with reasonable capacity to avoid memory bloat
	// Builders that grew too large are discarded
	if b.Cap() <= 32*1024 { // 32KB limit
		b.Reset()
		builderPool.Put(b)
	}
}

// WithBuilder executes a function with a pooled builder and returns the result string.
// This is a convenience wrapper that handles acquisition and release automatically.
func WithBuilder(fn func(*strings.Builder)) string {
	b := AcquireBuilder()
	fn(b)
	result := b.String()
	ReleaseBuilder(b)
	return result
}

// RenderBuffer is a helper for building render output with efficient string concatenation.
// It wraps a pooled strings.Builder with convenient methods for common rendering patterns.
type RenderBuffer struct {
	b *strings.Builder
}

// NewRenderBuffer creates a new RenderBuffer backed by a pooled strings.Builder.
func NewRenderBuffer() *RenderBuffer {
	return &RenderBuffer{b: AcquireBuilder()}
}

// WriteString appends a string to the buffer.
func (rb *RenderBuffer) WriteString(s string) *RenderBuffer {
	rb.b.WriteString(s)
	return rb
}

// WriteLine appends a string followed by a newline.
func (rb *RenderBuffer) WriteLine(s string) *RenderBuffer {
	rb.b.WriteString(s)
	rb.b.WriteString("\n")
	return rb
}

// Newline appends a newline character.
func (rb *RenderBuffer) Newline() *RenderBuffer {
	rb.b.WriteString("\n")
	return rb
}

// WriteFormatted writes a formatted string using the provided style function.
func (rb *RenderBuffer) WriteStyled(text string, styleFn func() string) *RenderBuffer {
	rb.b.WriteString(styleFn())
	return rb
}

// String returns the accumulated string and releases the underlying builder.
// After calling String(), the RenderBuffer should not be used.
func (rb *RenderBuffer) String() string {
	result := rb.b.String()
	ReleaseBuilder(rb.b)
	rb.b = nil
	return result
}

// Builder returns the underlying strings.Builder for direct access.
// Use this when you need to pass the builder to other functions.
func (rb *RenderBuffer) Builder() *strings.Builder {
	return rb.b
}

// Release returns the underlying builder to the pool without getting the string.
// Use this in error paths where the built string is not needed.
func (rb *RenderBuffer) Release() {
	if rb.b != nil {
		ReleaseBuilder(rb.b)
		rb.b = nil
	}
}
