package utils

import (
	"strings"
	"sync"
)

// StringBuilderPool provides pooled string builders for efficient string operations
type StringBuilderPool struct {
	pool sync.Pool
}

// NewStringBuilderPool creates a new string builder pool
func NewStringBuilderPool() *StringBuilderPool {
	return &StringBuilderPool{
		pool: sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
	}
}

// Get retrieves a string builder from the pool
func (p *StringBuilderPool) Get() *strings.Builder {
	return p.pool.Get().(*strings.Builder)
}

// Put returns a string builder to the pool after resetting it
func (p *StringBuilderPool) Put(builder *strings.Builder) {
	builder.Reset()
	p.pool.Put(builder)
}

// Global string builder pool for application-wide use
var globalStringBuilderPool = NewStringBuilderPool()

// GetStringBuilder gets a string builder from the global pool
func GetStringBuilder() *strings.Builder {
	return globalStringBuilderPool.Get()
}

// PutStringBuilder returns a string builder to the global pool
func PutStringBuilder(builder *strings.Builder) {
	globalStringBuilderPool.Put(builder)
}

// BuildString provides a convenient way to build strings with pooled builders
func BuildString(fn func(*strings.Builder)) string {
	builder := GetStringBuilder()
	defer PutStringBuilder(builder)

	fn(builder)
	return builder.String()
}

// JoinStrings efficiently joins strings using a pooled builder
func JoinStrings(separator string, parts ...string) string {
	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return BuildString(func(builder *strings.Builder) {
		builder.WriteString(parts[0])
		for _, s := range parts[1:] {
			builder.WriteString(separator)
			builder.WriteString(s)
		}
	})
}

// ConcatenateStrings efficiently concatenates strings using a pooled builder
func ConcatenateStrings(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	return BuildString(func(builder *strings.Builder) {
		for _, s := range parts {
			builder.WriteString(s)
		}
	})
}

// ByteBufferPool provides pooled byte buffers for efficient byte operations
type ByteBufferPool struct {
	pool sync.Pool
}

// NewByteBufferPool creates a new byte buffer pool
func NewByteBufferPool(initialSize int) *ByteBufferPool {
	return &ByteBufferPool{
		pool: sync.Pool{
			New: func() any {
				buf := make([]byte, 0, initialSize)
				return &buf
			},
		},
	}
}

// Get retrieves a byte buffer from the pool
func (p *ByteBufferPool) Get() []byte {
	bufPtr := p.pool.Get().(*[]byte)
	buf := *bufPtr
	return buf[:0] // Reset length but keep capacity
}

// Put returns a byte buffer to the pool after clearing it
func (p *ByteBufferPool) Put(buf []byte) {
	if cap(buf) > 0 && cap(buf) <= 32768 { // Don't pool huge buffers
		// Reset the slice but keep the capacity
		buf = buf[:0]
		p.pool.Put(&buf)
	}
}

// Global byte buffer pool for small operations (4KB initial size)
var globalByteBufferPool = NewByteBufferPool(4096)

// GetByteBuffer gets a byte buffer from the global pool
func GetByteBuffer() []byte {
	return globalByteBufferPool.Get()
}

// PutByteBuffer returns a byte buffer to the global pool
func PutByteBuffer(buf []byte) {
	globalByteBufferPool.Put(buf)
}

// Large byte buffer pool for bigger operations (64KB initial size)
var globalLargeByteBufferPool = NewByteBufferPool(65536)

// GetLargeByteBuffer gets a large byte buffer from the pool
func GetLargeByteBuffer() []byte {
	return globalLargeByteBufferPool.Get()
}

// PutLargeByteBuffer returns a large byte buffer to the pool
func PutLargeByteBuffer(buf []byte) {
	globalLargeByteBufferPool.Put(buf)
}
