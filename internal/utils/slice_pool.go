package utils

import (
	"reflect"
	"sync"
)

// SlicePool provides pooled slice allocations for template functions
type SlicePool struct {
	intSlicePool       sync.Pool
	stringSlicePool    sync.Pool
	interfaceSlicePool sync.Pool
	byteSlicePool      sync.Pool
}

// NewSlicePool creates a new slice pool
func NewSlicePool() *SlicePool {
	return &SlicePool{
		intSlicePool: sync.Pool{
			New: func() any {
				slice := make([]int, 0, 100) // Pre-allocate reasonable capacity
				return &slice
			},
		},
		stringSlicePool: sync.Pool{
			New: func() any {
				slice := make([]string, 0, 50) // Pre-allocate reasonable capacity
				return &slice
			},
		},
		interfaceSlicePool: sync.Pool{
			New: func() any {
				slice := make([]any, 0, 50) // Pre-allocate reasonable capacity
				return &slice
			},
		},
		byteSlicePool: sync.Pool{
			New: func() any {
				slice := make([]byte, 0, 1024) // Pre-allocate 1KB for byte operations
				return &slice
			},
		},
	}
}

// GetIntSlice gets a pooled int slice
func (p *SlicePool) GetIntSlice() *[]int {
	slice := p.intSlicePool.Get().(*[]int)
	*slice = (*slice)[:0] // Reset length but keep capacity
	return slice
}

// PutIntSlice returns an int slice to the pool
func (p *SlicePool) PutIntSlice(slice *[]int) {
	if slice != nil && cap(*slice) <= 1000 { // Prevent memory leaks from huge slices
		p.intSlicePool.Put(slice)
	}
}

// GetStringSlice gets a pooled string slice
func (p *SlicePool) GetStringSlice() *[]string {
	slice := p.stringSlicePool.Get().(*[]string)
	*slice = (*slice)[:0] // Reset length but keep capacity
	return slice
}

// PutStringSlice returns a string slice to the pool
func (p *SlicePool) PutStringSlice(slice *[]string) {
	if slice != nil && cap(*slice) <= 500 { // Prevent memory leaks from huge slices
		p.stringSlicePool.Put(slice)
	}
}

// GetInterfaceSlice gets a pooled interface{} slice
func (p *SlicePool) GetInterfaceSlice() *[]any {
	slice := p.interfaceSlicePool.Get().(*[]any)
	*slice = (*slice)[:0] // Reset length but keep capacity
	return slice
}

// PutInterfaceSlice returns an interface{} slice to the pool
func (p *SlicePool) PutInterfaceSlice(slice *[]any) {
	if slice != nil && cap(*slice) <= 500 { // Prevent memory leaks from huge slices
		p.interfaceSlicePool.Put(slice)
	}
}

// GetByteSlice gets a pooled byte slice
func (p *SlicePool) GetByteSlice() *[]byte {
	slice := p.byteSlicePool.Get().(*[]byte)
	*slice = (*slice)[:0] // Reset length but keep capacity
	return slice
}

// PutByteSlice returns a byte slice to the pool
func (p *SlicePool) PutByteSlice(slice *[]byte) {
	if slice != nil && cap(*slice) <= 10240 { // Prevent memory leaks from huge slices (10KB max)
		p.byteSlicePool.Put(slice)
	}
}

// WithIntSlice provides a convenient way to use pooled int slice
func (p *SlicePool) WithIntSlice(fn func(*[]int) []int) []int {
	slice := p.GetIntSlice()
	defer p.PutIntSlice(slice)
	return fn(slice)
}

// WithStringSlice provides a convenient way to use pooled string slice
func (p *SlicePool) WithStringSlice(fn func(*[]string) []string) []string {
	slice := p.GetStringSlice()
	defer p.PutStringSlice(slice)
	return fn(slice)
}

// WithInterfaceSlice provides a convenient way to use pooled interface slice
func (p *SlicePool) WithInterfaceSlice(fn func(*[]any) []any) []any {
	slice := p.GetInterfaceSlice()
	defer p.PutInterfaceSlice(slice)
	return fn(slice)
}

// WithByteSlice provides a convenient way to use pooled byte slice
func (p *SlicePool) WithByteSlice(fn func(*[]byte) []byte) []byte {
	slice := p.GetByteSlice()
	defer p.PutByteSlice(slice)
	return fn(slice)
}

// SliceCopyPooled performs a pooled slice copy operation
func (p *SlicePool) SliceCopyPooled(src any, start, end int) any {
	val := reflect.ValueOf(src)

	// Check if it's a slice or array
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return src // Return original if not slice/array
	}

	length := val.Len()

	// Boundary checks
	if start < 0 {
		start = 0
	}
	if start > length {
		start = length
	}
	if end < start {
		end = start
	}
	if end > length {
		end = length
	}

	sliceLen := end - start
	if sliceLen == 0 {
		// Return empty slice of the same type
		return reflect.MakeSlice(val.Type(), 0, 0).Interface()
	}

	// Create new slice of the same type
	result := reflect.MakeSlice(val.Type(), sliceLen, sliceLen)
	reflect.Copy(result, val.Slice(start, end))

	return result.Interface()
}

// IntSequencePooled generates an int sequence using pooled slice
func (p *SlicePool) IntSequencePooled(start, end int) []int {
	if start > end {
		return []int{}
	}

	return p.WithIntSlice(func(pooledSlice *[]int) []int {
		// Calculate required capacity
		seqLen := end - start + 1

		// Ensure capacity
		if cap(*pooledSlice) < seqLen {
			// If pool slice is too small, create new slice and don't return to pool
			result := make([]int, seqLen)
			for i := 0; i < seqLen; i++ {
				result[i] = start + i
			}
			return result
		}

		// Use pooled slice
		*pooledSlice = (*pooledSlice)[:seqLen]
		for i := 0; i < seqLen; i++ {
			(*pooledSlice)[i] = start + i
		}

		// Return a copy since the pooled slice will go back to pool
		result := make([]int, seqLen)
		copy(result, *pooledSlice)
		return result
	})
}

// Global slice pool for application-wide use
var globalSlicePool = NewSlicePool()

// GetGlobalSlicePool returns the global slice pool
func GetGlobalSlicePool() *SlicePool {
	return globalSlicePool
}

// RuneSlicePool provides pooled rune slice allocations for string processing
type RuneSlicePool struct {
	pool sync.Pool
}

// NewRuneSlicePool creates a new rune slice pool
func NewRuneSlicePool() *RuneSlicePool {
	return &RuneSlicePool{
		pool: sync.Pool{
			New: func() any {
				slice := make([]rune, 0, 500) // Pre-allocate for typical text processing
				return &slice
			},
		},
	}
}

// GetRuneSlice gets a pooled rune slice
func (p *RuneSlicePool) GetRuneSlice() *[]rune {
	slice := p.pool.Get().(*[]rune)
	*slice = (*slice)[:0] // Reset length but keep capacity
	return slice
}

// PutRuneSlice returns a rune slice to the pool
func (p *RuneSlicePool) PutRuneSlice(slice *[]rune) {
	if slice != nil && cap(*slice) <= 2000 { // Prevent memory leaks from huge slices
		p.pool.Put(slice)
	}
}

// WithRuneSlice provides a convenient way to use pooled rune slice
func (p *RuneSlicePool) WithRuneSlice(fn func(*[]rune) []rune) []rune {
	slice := p.GetRuneSlice()
	defer p.PutRuneSlice(slice)
	return fn(slice)
}

// Global rune slice pool
var globalRuneSlicePool = NewRuneSlicePool()

// GetGlobalRuneSlicePool returns the global rune slice pool
func GetGlobalRuneSlicePool() *RuneSlicePool {
	return globalRuneSlicePool
}
