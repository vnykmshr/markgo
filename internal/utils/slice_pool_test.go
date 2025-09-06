package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSlicePool(t *testing.T) {
	pool := NewSlicePool()
	assert.NotNil(t, pool)
}

func TestSlicePool_IntSlice(t *testing.T) {
	pool := NewSlicePool()

	// Test getting an int slice
	slice := pool.GetIntSlice()
	assert.NotNil(t, slice)
	assert.Equal(t, 0, len(*slice))
	assert.True(t, cap(*slice) >= 100) // Should have pre-allocated capacity

	// Test using the slice
	*slice = append(*slice, 1, 2, 3)
	assert.Equal(t, []int{1, 2, 3}, *slice)

	// Test returning to pool
	pool.PutIntSlice(slice)

	// Test getting again (should be reset)
	slice2 := pool.GetIntSlice()
	assert.Equal(t, 0, len(*slice2))
	// Should be the same slice instance (pooled)
	assert.Equal(t, slice, slice2)
}

func TestSlicePool_StringSlice(t *testing.T) {
	pool := NewSlicePool()

	slice := pool.GetStringSlice()
	assert.NotNil(t, slice)
	assert.Equal(t, 0, len(*slice))
	assert.True(t, cap(*slice) >= 50)

	*slice = append(*slice, "hello", "world")
	assert.Equal(t, []string{"hello", "world"}, *slice)

	pool.PutStringSlice(slice)

	slice2 := pool.GetStringSlice()
	assert.Equal(t, 0, len(*slice2))
	assert.Equal(t, slice, slice2)
}

func TestSlicePool_InterfaceSlice(t *testing.T) {
	pool := NewSlicePool()

	slice := pool.GetInterfaceSlice()
	assert.NotNil(t, slice)
	assert.Equal(t, 0, len(*slice))
	assert.True(t, cap(*slice) >= 50)

	*slice = append(*slice, "hello", 123, true)
	assert.Equal(t, []any{"hello", 123, true}, *slice)

	pool.PutInterfaceSlice(slice)

	slice2 := pool.GetInterfaceSlice()
	assert.Equal(t, 0, len(*slice2))
	assert.Equal(t, slice, slice2)
}

func TestSlicePool_ByteSlice(t *testing.T) {
	pool := NewSlicePool()

	slice := pool.GetByteSlice()
	assert.NotNil(t, slice)
	assert.Equal(t, 0, len(*slice))
	assert.True(t, cap(*slice) >= 1024)

	*slice = append(*slice, []byte("hello")...)
	assert.Equal(t, []byte("hello"), *slice)

	pool.PutByteSlice(slice)

	slice2 := pool.GetByteSlice()
	assert.Equal(t, 0, len(*slice2))
	assert.Equal(t, slice, slice2)
}

func TestSlicePool_NilSafety(t *testing.T) {
	pool := NewSlicePool()

	// Should not panic with nil slices
	assert.NotPanics(t, func() {
		pool.PutIntSlice(nil)
	})
	assert.NotPanics(t, func() {
		pool.PutStringSlice(nil)
	})
	assert.NotPanics(t, func() {
		pool.PutInterfaceSlice(nil)
	})
	assert.NotPanics(t, func() {
		pool.PutByteSlice(nil)
	})
}

func TestSlicePool_SizeLimits(t *testing.T) {
	pool := NewSlicePool()

	// Create slices with large capacity (over limits)
	largeIntSlice := make([]int, 0, 2000)       // Over 1000 limit
	largeStringSlice := make([]string, 0, 1000) // Over 500 limit
	largeInterfaceSlice := make([]any, 0, 1000) // Over 500 limit
	largeByteSlice := make([]byte, 0, 20000)    // Over 10KB limit

	// Should not panic, but slices won't be returned to pool due to size
	assert.NotPanics(t, func() {
		pool.PutIntSlice(&largeIntSlice)
	})
	assert.NotPanics(t, func() {
		pool.PutStringSlice(&largeStringSlice)
	})
	assert.NotPanics(t, func() {
		pool.PutInterfaceSlice(&largeInterfaceSlice)
	})
	assert.NotPanics(t, func() {
		pool.PutByteSlice(&largeByteSlice)
	})
}

func TestSlicePool_WithIntSlice(t *testing.T) {
	pool := NewSlicePool()

	result := pool.WithIntSlice(func(slice *[]int) []int {
		*slice = append(*slice, 1, 2, 3, 4, 5)
		return *slice
	})

	assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
}

func TestSlicePool_WithStringSlice(t *testing.T) {
	pool := NewSlicePool()

	result := pool.WithStringSlice(func(slice *[]string) []string {
		*slice = append(*slice, "a", "b", "c")
		return *slice
	})

	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestSlicePool_WithInterfaceSlice(t *testing.T) {
	pool := NewSlicePool()

	result := pool.WithInterfaceSlice(func(slice *[]any) []any {
		*slice = append(*slice, 1, "hello", true)
		return *slice
	})

	assert.Equal(t, []any{1, "hello", true}, result)
}

func TestSlicePool_WithByteSlice(t *testing.T) {
	pool := NewSlicePool()

	result := pool.WithByteSlice(func(slice *[]byte) []byte {
		*slice = append(*slice, []byte("test data")...)
		return *slice
	})

	assert.Equal(t, []byte("test data"), result)
}

func TestSlicePool_SliceCopyPooled(t *testing.T) {
	pool := NewSlicePool()

	// Test with int slice
	intSlice := []int{1, 2, 3, 4, 5}
	result := pool.SliceCopyPooled(intSlice, 1, 4)
	assert.Equal(t, []int{2, 3, 4}, result)

	// Test with string slice
	stringSlice := []string{"a", "b", "c", "d"}
	result = pool.SliceCopyPooled(stringSlice, 0, 2)
	assert.Equal(t, []string{"a", "b"}, result)

	// Test with invalid bounds (start > end)
	result = pool.SliceCopyPooled(intSlice, 3, 1)
	assert.Equal(t, []int{}, result)

	// Test with out of bounds indices
	result = pool.SliceCopyPooled(intSlice, -1, 10)
	assert.Equal(t, intSlice, result) // Should return full slice with bounds correction

	// Test with non-slice input
	result = pool.SliceCopyPooled("not a slice", 0, 5)
	assert.Equal(t, "not a slice", result)

	// Test with empty slice
	emptySlice := []int{}
	result = pool.SliceCopyPooled(emptySlice, 0, 10)
	assert.Equal(t, []int{}, result)
}

func TestSlicePool_IntSequencePooled(t *testing.T) {
	pool := NewSlicePool()

	// Test normal sequence
	result := pool.IntSequencePooled(1, 5)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, result)

	// Test single element
	result = pool.IntSequencePooled(10, 10)
	assert.Equal(t, []int{10}, result)

	// Test empty sequence (start > end)
	result = pool.IntSequencePooled(5, 1)
	assert.Equal(t, []int{}, result)

	// Test sequence that exceeds pool capacity
	result = pool.IntSequencePooled(1, 200)
	assert.Equal(t, 200, len(result))
	assert.Equal(t, 1, result[0])
	assert.Equal(t, 200, result[199])

	// Test negative numbers
	result = pool.IntSequencePooled(-3, 3)
	assert.Equal(t, []int{-3, -2, -1, 0, 1, 2, 3}, result)
}

func TestGetGlobalSlicePool(t *testing.T) {
	pool1 := GetGlobalSlicePool()
	pool2 := GetGlobalSlicePool()

	assert.NotNil(t, pool1)
	assert.Equal(t, pool1, pool2) // Should be the same instance (singleton)
}

func TestNewRuneSlicePool(t *testing.T) {
	pool := NewRuneSlicePool()
	assert.NotNil(t, pool)
}

func TestRuneSlicePool_Operations(t *testing.T) {
	pool := NewRuneSlicePool()

	// Test getting a rune slice
	slice := pool.GetRuneSlice()
	assert.NotNil(t, slice)
	assert.Equal(t, 0, len(*slice))
	assert.True(t, cap(*slice) >= 500)

	// Test using the slice
	*slice = append(*slice, 'h', 'e', 'l', 'l', 'o')
	assert.Equal(t, []rune{'h', 'e', 'l', 'l', 'o'}, *slice)

	// Test returning to pool
	pool.PutRuneSlice(slice)

	// Test getting again (should be reset)
	slice2 := pool.GetRuneSlice()
	assert.Equal(t, 0, len(*slice2))
	assert.Equal(t, slice, slice2) // Same instance
}

func TestRuneSlicePool_NilSafety(t *testing.T) {
	pool := NewRuneSlicePool()

	assert.NotPanics(t, func() {
		pool.PutRuneSlice(nil)
	})
}

func TestRuneSlicePool_SizeLimit(t *testing.T) {
	pool := NewRuneSlicePool()

	// Create a slice with large capacity (over 2000 limit)
	largeSlice := make([]rune, 0, 3000)

	// Should not panic, but slice won't be returned to pool
	assert.NotPanics(t, func() {
		pool.PutRuneSlice(&largeSlice)
	})
}

func TestRuneSlicePool_WithRuneSlice(t *testing.T) {
	pool := NewRuneSlicePool()

	result := pool.WithRuneSlice(func(slice *[]rune) []rune {
		text := "Hello, 世界!"
		*slice = append(*slice, []rune(text)...)
		return *slice
	})

	expected := []rune("Hello, 世界!")
	assert.Equal(t, expected, result)
}

func TestGetGlobalRuneSlicePool(t *testing.T) {
	pool1 := GetGlobalRuneSlicePool()
	pool2 := GetGlobalRuneSlicePool()

	assert.NotNil(t, pool1)
	assert.Equal(t, pool1, pool2) // Should be the same instance (singleton)
}

// Benchmark tests
func BenchmarkSlicePool_IntSlice(b *testing.B) {
	pool := NewSlicePool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slice := pool.GetIntSlice()
		*slice = append(*slice, 1, 2, 3, 4, 5)
		pool.PutIntSlice(slice)
	}
}

func BenchmarkSlicePool_WithIntSlice(b *testing.B) {
	pool := NewSlicePool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.WithIntSlice(func(slice *[]int) []int {
			*slice = append(*slice, 1, 2, 3, 4, 5)
			return *slice
		})
	}
}

func BenchmarkSlicePool_IntSequencePooled(b *testing.B) {
	pool := NewSlicePool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.IntSequencePooled(1, 50)
	}
}

func BenchmarkRuneSlicePool_WithRuneSlice(b *testing.B) {
	pool := NewRuneSlicePool()
	text := []rune("This is a test string with unicode: 世界")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.WithRuneSlice(func(slice *[]rune) []rune {
			*slice = append(*slice, text...)
			return *slice
		})
	}
}
