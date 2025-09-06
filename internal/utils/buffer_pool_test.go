package utils

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBufferPool(t *testing.T) {
	pool := NewBufferPool()
	assert.NotNil(t, pool)
}

func TestBufferPool_BytesBuffer(t *testing.T) {
	pool := NewBufferPool()

	// Test getting a buffer
	buf := pool.GetBytesBuffer()
	require.NotNil(t, buf)
	assert.Equal(t, 0, buf.Len()) // Should be reset

	// Test using the buffer
	_, err := buf.WriteString("test content")
	require.NoError(t, err)
	assert.Equal(t, "test content", buf.String())

	// Test returning the buffer to pool
	pool.PutBytesBuffer(buf)

	// Test getting another buffer (should be the same one, reset)
	buf2 := pool.GetBytesBuffer()
	assert.Equal(t, 0, buf2.Len()) // Should be reset

	// Should be the same buffer instance (pooled)
	assert.Equal(t, buf, buf2)
}

func TestBufferPool_StringsBuilder(t *testing.T) {
	pool := NewBufferPool()

	// Test getting a builder
	builder := pool.GetStringsBuilder()
	require.NotNil(t, builder)
	assert.Equal(t, 0, builder.Len()) // Should be reset

	// Test using the builder
	builder.WriteString("test content")
	assert.Equal(t, "test content", builder.String())

	// Test returning the builder to pool
	pool.PutStringsBuilder(builder)

	// Test getting another builder (should be the same one, reset)
	builder2 := pool.GetStringsBuilder()
	assert.Equal(t, 0, builder2.Len()) // Should be reset

	// Should be the same builder instance (pooled)
	assert.Equal(t, builder, builder2)
}

func TestBufferPool_PutBytesBuffer_NilSafety(t *testing.T) {
	pool := NewBufferPool()

	// Should not panic with nil buffer
	assert.NotPanics(t, func() {
		pool.PutBytesBuffer(nil)
	})
}

func TestBufferPool_PutStringsBuilder_NilSafety(t *testing.T) {
	pool := NewBufferPool()

	// Should not panic with nil builder
	assert.NotPanics(t, func() {
		pool.PutStringsBuilder(nil)
	})
}

func TestBufferPool_PutBytesBuffer_SizeLimit(t *testing.T) {
	pool := NewBufferPool()

	// Create a very large buffer (over 64KB limit)
	hugeBuf := bytes.NewBuffer(make([]byte, 70000))

	// This should not panic, but the buffer won't be returned to pool
	assert.NotPanics(t, func() {
		pool.PutBytesBuffer(hugeBuf)
	})
}

func TestBufferPool_PutStringsBuilder_SizeLimit(t *testing.T) {
	pool := NewBufferPool()

	// Create a builder with large capacity (over 64KB limit)
	var builder strings.Builder
	builder.Grow(70000)
	builder.WriteString("test") // Use some capacity

	// This should not panic, but the builder won't be returned to pool
	assert.NotPanics(t, func() {
		pool.PutStringsBuilder(&builder)
	})
}

func TestBufferPool_WithBytesBuffer(t *testing.T) {
	pool := NewBufferPool()

	result, err := pool.WithBytesBuffer(func(buf *bytes.Buffer) error {
		buf.WriteString("test content")
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, "test content", result.String())

	// Test with error
	_, err = pool.WithBytesBuffer(func(buf *bytes.Buffer) error {
		return errors.New("test error")
	})

	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())
}

func TestBufferPool_WithStringsBuilder(t *testing.T) {
	pool := NewBufferPool()

	result := pool.WithStringsBuilder(func(builder *strings.Builder) string {
		builder.WriteString("test content")
		return builder.String()
	})

	assert.Equal(t, "test content", result)
}

func TestBufferPool_RenderToString(t *testing.T) {
	pool := NewBufferPool()

	result, err := pool.RenderToString(func(buf *bytes.Buffer) error {
		buf.WriteString("rendered content")
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, "rendered content", result)

	// Test with error
	_, err = pool.RenderToString(func(buf *bytes.Buffer) error {
		return errors.New("render error")
	})

	assert.Error(t, err)
	assert.Equal(t, "render error", err.Error())
}

func TestBufferPool_BuildString(t *testing.T) {
	pool := NewBufferPool()

	result := pool.BuildString(func(builder *strings.Builder) {
		builder.WriteString("built content")
	})

	assert.Equal(t, "built content", result)
}

func TestGetGlobalBufferPool(t *testing.T) {
	pool1 := GetGlobalBufferPool()
	pool2 := GetGlobalBufferPool()

	assert.NotNil(t, pool1)
	assert.Equal(t, pool1, pool2) // Should be the same instance
}

func TestNewFeedBufferPool(t *testing.T) {
	pool := NewFeedBufferPool()
	assert.NotNil(t, pool)
}

func TestFeedBufferPool_Operations(t *testing.T) {
	pool := NewFeedBufferPool()

	// Test getting and using builder
	builder := pool.GetBuilder()
	require.NotNil(t, builder)
	assert.Equal(t, 0, builder.Len())

	builder.WriteString("feed content")
	assert.Equal(t, "feed content", builder.String())

	// Test returning to pool
	pool.PutBuilder(builder)

	// Test getting again (should be reset)
	builder2 := pool.GetBuilder()
	assert.Equal(t, 0, builder2.Len())
	assert.Equal(t, builder, builder2) // Same instance
}

func TestFeedBufferPool_PutBuilder_NilSafety(t *testing.T) {
	pool := NewFeedBufferPool()

	assert.NotPanics(t, func() {
		pool.PutBuilder(nil)
	})
}

func TestFeedBufferPool_PutBuilder_SizeLimit(t *testing.T) {
	pool := NewFeedBufferPool()

	// Create a builder with capacity over 128KB limit
	var builder strings.Builder
	builder.Grow(140000)
	builder.WriteString("test")

	assert.NotPanics(t, func() {
		pool.PutBuilder(&builder)
	})
}

func TestFeedBufferPool_BuildFeed(t *testing.T) {
	pool := NewFeedBufferPool()

	result := pool.BuildFeed(func(builder *strings.Builder) {
		builder.WriteString("<?xml version=\"1.0\"?><rss>content</rss>")
	})

	assert.Equal(t, "<?xml version=\"1.0\"?><rss>content</rss>", result)
}

func TestGetGlobalFeedBufferPool(t *testing.T) {
	pool1 := GetGlobalFeedBufferPool()
	pool2 := GetGlobalFeedBufferPool()

	assert.NotNil(t, pool1)
	assert.Equal(t, pool1, pool2) // Should be the same instance
}

func TestNewSmallBufferPool(t *testing.T) {
	pool := NewSmallBufferPool()
	assert.NotNil(t, pool)
}

func TestSmallBufferPool_Operations(t *testing.T) {
	pool := NewSmallBufferPool()

	// Test getting and using builder
	builder := pool.GetBuilder()
	require.NotNil(t, builder)
	assert.Equal(t, 0, builder.Len())

	builder.WriteString("small content")
	assert.Equal(t, "small content", builder.String())

	// Test returning to pool
	pool.PutBuilder(builder)

	// Test getting again (should be reset)
	builder2 := pool.GetBuilder()
	assert.Equal(t, 0, builder2.Len())
	assert.Equal(t, builder, builder2) // Same instance
}

func TestSmallBufferPool_PutBuilder_NilSafety(t *testing.T) {
	pool := NewSmallBufferPool()

	assert.NotPanics(t, func() {
		pool.PutBuilder(nil)
	})
}

func TestSmallBufferPool_PutBuilder_SizeLimit(t *testing.T) {
	pool := NewSmallBufferPool()

	// Create a builder with capacity over 2KB limit
	var builder strings.Builder
	builder.Grow(3000)
	builder.WriteString("test")

	assert.NotPanics(t, func() {
		pool.PutBuilder(&builder)
	})
}

func TestSmallBufferPool_BuildSmallString(t *testing.T) {
	pool := NewSmallBufferPool()

	result := pool.BuildSmallString(func(builder *strings.Builder) {
		builder.WriteString("header-value")
	})

	assert.Equal(t, "header-value", result)
}

func TestGetGlobalSmallBufferPool(t *testing.T) {
	pool1 := GetGlobalSmallBufferPool()
	pool2 := GetGlobalSmallBufferPool()

	assert.NotNil(t, pool1)
	assert.Equal(t, pool1, pool2) // Should be the same instance
}

// Benchmark tests
func BenchmarkBufferPool_GetPutBytesBuffer(b *testing.B) {
	pool := NewBufferPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.GetBytesBuffer()
		buf.WriteString("test content")
		pool.PutBytesBuffer(buf)
	}
}

func BenchmarkBufferPool_GetPutStringsBuilder(b *testing.B) {
	pool := NewBufferPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := pool.GetStringsBuilder()
		builder.WriteString("test content")
		pool.PutStringsBuilder(builder)
	}
}

func BenchmarkFeedBufferPool_BuildFeed(b *testing.B) {
	pool := NewFeedBufferPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.BuildFeed(func(builder *strings.Builder) {
			builder.WriteString("<?xml version=\"1.0\"?><rss><item>test</item></rss>")
		})
	}
}

func BenchmarkSmallBufferPool_BuildSmallString(b *testing.B) {
	pool := NewSmallBufferPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.BuildSmallString(func(builder *strings.Builder) {
			builder.WriteString("cache-control: max-age=3600")
		})
	}
}
