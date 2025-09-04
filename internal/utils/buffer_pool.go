package utils

import (
	"bytes"
	"strings"
	"sync"
)

// BufferPool provides pooled buffer allocations for template rendering and string building
type BufferPool struct {
	bytesBufferPool   sync.Pool
	stringsBuilderPool sync.Pool
}

// NewBufferPool creates a new buffer pool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		bytesBufferPool: sync.Pool{
			New: func() any {
				// Pre-allocate with reasonable capacity for typical template sizes
				return bytes.NewBuffer(make([]byte, 0, 4096))
			},
		},
		stringsBuilderPool: sync.Pool{
			New: func() any {
				var b strings.Builder
				b.Grow(2048) // Pre-allocate reasonable capacity
				return &b
			},
		},
	}
}

// GetBytesBuffer gets a pooled bytes.Buffer
func (p *BufferPool) GetBytesBuffer() *bytes.Buffer {
	buf := p.bytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset() // Clear the buffer but keep allocated memory
	return buf
}

// PutBytesBuffer returns a bytes.Buffer to the pool
func (p *BufferPool) PutBytesBuffer(buf *bytes.Buffer) {
	if buf != nil && buf.Cap() <= 65536 { // Prevent memory leaks from huge buffers (64KB max)
		p.bytesBufferPool.Put(buf)
	}
}

// GetStringsBuilder gets a pooled strings.Builder
func (p *BufferPool) GetStringsBuilder() *strings.Builder {
	builder := p.stringsBuilderPool.Get().(*strings.Builder)
	builder.Reset() // Clear the builder but keep allocated memory
	return builder
}

// PutStringsBuilder returns a strings.Builder to the pool
func (p *BufferPool) PutStringsBuilder(builder *strings.Builder) {
	if builder != nil && builder.Cap() <= 65536 { // Prevent memory leaks from huge builders (64KB max)
		p.stringsBuilderPool.Put(builder)
	}
}

// WithBytesBuffer provides a convenient way to use pooled bytes.Buffer
func (p *BufferPool) WithBytesBuffer(fn func(*bytes.Buffer) error) (*bytes.Buffer, error) {
	buf := p.GetBytesBuffer()
	defer p.PutBytesBuffer(buf)
	err := fn(buf)
	
	// Return a copy since the pooled buffer will go back to pool
	result := bytes.NewBuffer(make([]byte, 0, buf.Len()))
	result.Write(buf.Bytes())
	return result, err
}

// WithStringsBuilder provides a convenient way to use pooled strings.Builder
func (p *BufferPool) WithStringsBuilder(fn func(*strings.Builder) string) string {
	builder := p.GetStringsBuilder()
	defer p.PutStringsBuilder(builder)
	return fn(builder)
}

// RenderToString renders using pooled buffer and returns string
func (p *BufferPool) RenderToString(renderFn func(*bytes.Buffer) error) (string, error) {
	buf := p.GetBytesBuffer()
	defer p.PutBytesBuffer(buf)
	
	if err := renderFn(buf); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// BuildString builds a string using pooled strings.Builder
func (p *BufferPool) BuildString(buildFn func(*strings.Builder)) string {
	builder := p.GetStringsBuilder()
	defer p.PutStringsBuilder(builder)
	
	buildFn(builder)
	return builder.String()
}

// Global buffer pool for application-wide use
var globalBufferPool = NewBufferPool()

// GetGlobalBufferPool returns the global buffer pool
func GetGlobalBufferPool() *BufferPool {
	return globalBufferPool
}

// RSS/XML Buffer Pool for feed generation
type FeedBufferPool struct {
	pool sync.Pool
}

// NewFeedBufferPool creates a buffer pool optimized for RSS/XML feed generation
func NewFeedBufferPool() *FeedBufferPool {
	return &FeedBufferPool{
		pool: sync.Pool{
			New: func() any {
				var b strings.Builder
				b.Grow(8192) // Pre-allocate 8KB for typical feed sizes
				return &b
			},
		},
	}
}

// GetBuilder gets a pooled strings.Builder for feed generation
func (p *FeedBufferPool) GetBuilder() *strings.Builder {
	builder := p.pool.Get().(*strings.Builder)
	builder.Reset()
	return builder
}

// PutBuilder returns a strings.Builder to the pool
func (p *FeedBufferPool) PutBuilder(builder *strings.Builder) {
	if builder != nil && builder.Cap() <= 131072 { // Prevent memory leaks (128KB max for feeds)
		p.pool.Put(builder)
	}
}

// BuildFeed builds a feed using pooled strings.Builder
func (p *FeedBufferPool) BuildFeed(buildFn func(*strings.Builder)) string {
	builder := p.GetBuilder()
	defer p.PutBuilder(builder)
	
	buildFn(builder)
	return builder.String()
}

// Global feed buffer pool
var globalFeedBufferPool = NewFeedBufferPool()

// GetGlobalFeedBufferPool returns the global feed buffer pool
func GetGlobalFeedBufferPool() *FeedBufferPool {
	return globalFeedBufferPool
}

// Small Buffer Pool for header values and small string operations
type SmallBufferPool struct {
	pool sync.Pool
}

// NewSmallBufferPool creates a buffer pool for small string operations
func NewSmallBufferPool() *SmallBufferPool {
	return &SmallBufferPool{
		pool: sync.Pool{
			New: func() any {
				var b strings.Builder
				b.Grow(256) // Pre-allocate 256 bytes for small operations
				return &b
			},
		},
	}
}

// GetBuilder gets a pooled strings.Builder for small operations
func (p *SmallBufferPool) GetBuilder() *strings.Builder {
	builder := p.pool.Get().(*strings.Builder)
	builder.Reset()
	return builder
}

// PutBuilder returns a strings.Builder to the pool
func (p *SmallBufferPool) PutBuilder(builder *strings.Builder) {
	if builder != nil && builder.Cap() <= 2048 { // Prevent memory leaks (2KB max for small buffers)
		p.pool.Put(builder)
	}
}

// BuildSmallString builds a small string using pooled strings.Builder
func (p *SmallBufferPool) BuildSmallString(buildFn func(*strings.Builder)) string {
	builder := p.GetBuilder()
	defer p.PutBuilder(builder)
	
	buildFn(builder)
	return builder.String()
}

// Global small buffer pool
var globalSmallBufferPool = NewSmallBufferPool()

// GetGlobalSmallBufferPool returns the global small buffer pool
func GetGlobalSmallBufferPool() *SmallBufferPool {
	return globalSmallBufferPool
}