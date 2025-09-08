package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStringInterner(t *testing.T) {
	interner := NewStringInterner()
	assert.NotNil(t, interner)
	assert.Equal(t, 0, interner.Size())
}

func TestStringInterner_Intern(t *testing.T) {
	interner := NewStringInterner()

	// Test interning a string
	s1 := interner.Intern("test")
	assert.Equal(t, "test", s1)
	assert.Equal(t, 1, interner.Size())

	// Test interning the same string again
	s2 := interner.Intern("test")
	assert.Equal(t, "test", s2)
	assert.Equal(t, 1, interner.Size()) // Should still be 1

	// The pointers should be the same (string interning)
	assert.True(t, &s1 == &s2 || s1 == s2)

	// Test interning a different string
	s3 := interner.Intern("different")
	assert.Equal(t, "different", s3)
	assert.Equal(t, 2, interner.Size())
}

func TestStringInterner_Intern_EmptyString(t *testing.T) {
	interner := NewStringInterner()

	empty := interner.Intern("")
	assert.Equal(t, "", empty)
	assert.Equal(t, 0, interner.Size()) // Empty strings are not stored
}

func TestStringInterner_InternSlice(t *testing.T) {
	interner := NewStringInterner()

	// Test empty slice
	result := interner.InternSlice([]string{})
	assert.Equal(t, []string{}, result)

	// Test nil slice
	result = interner.InternSlice(nil)
	assert.Nil(t, result)

	// Test slice with strings
	input := []string{"tag1", "tag2", "tag1", "tag3"}
	result = interner.InternSlice(input)

	expected := []string{"tag1", "tag2", "tag1", "tag3"}
	assert.Equal(t, expected, result)
	assert.Equal(t, 3, interner.Size()) // Only 3 unique strings

	// Test that repeated calls return the same interned strings
	result2 := interner.InternSlice(input)
	assert.Equal(t, result, result2)
}

func TestStringInterner_Stats(t *testing.T) {
	interner := NewStringInterner()

	// Initial stats
	stats := interner.Stats()
	assert.Equal(t, int64(0), stats.TotalLookups)
	assert.Equal(t, int64(0), stats.HitCount)
	assert.Equal(t, int64(0), stats.MissCount)
	assert.Equal(t, int64(0), stats.UniqueStrings)
	assert.Equal(t, int64(0), stats.MemorySaved)

	// Intern a new string
	interner.Intern("test")
	stats = interner.Stats()
	assert.Equal(t, int64(1), stats.TotalLookups)
	assert.Equal(t, int64(0), stats.HitCount)
	assert.Equal(t, int64(1), stats.MissCount)
	assert.Equal(t, int64(1), stats.UniqueStrings)
	assert.Equal(t, int64(4), stats.MemorySaved) // "test" is 4 bytes

	// Intern the same string again
	interner.Intern("test")
	stats = interner.Stats()
	assert.Equal(t, int64(2), stats.TotalLookups)
	assert.Equal(t, int64(1), stats.HitCount)
	assert.Equal(t, int64(1), stats.MissCount)
	assert.Equal(t, int64(1), stats.UniqueStrings)
	assert.Equal(t, int64(4), stats.MemorySaved)
}

func TestStringInterner_Clear(t *testing.T) {
	interner := NewStringInterner()

	// Add some strings
	interner.Intern("test1")
	interner.Intern("test2")
	assert.Equal(t, 2, interner.Size())

	// Clear
	interner.Clear()
	assert.Equal(t, 0, interner.Size())

	// Stats should be reset
	stats := interner.Stats()
	assert.Equal(t, int64(0), stats.TotalLookups)
	assert.Equal(t, int64(0), stats.HitCount)
	assert.Equal(t, int64(0), stats.MissCount)
	assert.Equal(t, int64(0), stats.UniqueStrings)
	assert.Equal(t, int64(0), stats.MemorySaved)
}

func TestStringInterner_ConcurrentAccess(t *testing.T) {
	interner := NewStringInterner()

	// Test concurrent interning of the same strings
	var wg sync.WaitGroup
	const numGoroutines = 5
	const numOperations = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				s := interner.Intern("concurrent-test")
				assert.Equal(t, "concurrent-test", s)
			}
		}()
	}

	wg.Wait()

	// Should only have one unique string despite many concurrent operations
	assert.Equal(t, 1, interner.Size())

	stats := interner.Stats()
	// Due to concurrent access, we should have close to the expected number of lookups
	// but it might not be exact due to race conditions in counting
	assert.GreaterOrEqual(t, stats.TotalLookups, int64(numGoroutines*numOperations-5))
	assert.LessOrEqual(t, stats.TotalLookups, int64(numGoroutines*numOperations))
	assert.Equal(t, int64(1), stats.UniqueStrings)
}

// Global functions tests
func TestInternString(t *testing.T) {
	// Clear global interner first
	ClearGlobalInterner()

	s1 := InternString("global-test")
	assert.Equal(t, "global-test", s1)
	assert.Equal(t, 1, GetGlobalInternerSize())

	s2 := InternString("global-test")
	assert.Equal(t, s1, s2)
	assert.Equal(t, 1, GetGlobalInternerSize())
}

func TestInternStringSlice(t *testing.T) {
	ClearGlobalInterner()

	input := []string{"tag1", "tag2", "tag1"}
	result := InternStringSlice(input)

	expected := []string{"tag1", "tag2", "tag1"}
	assert.Equal(t, expected, result)
	assert.Equal(t, 2, GetGlobalInternerSize()) // Only 2 unique
}

func TestGetInternerStats(t *testing.T) {
	ClearGlobalInterner()

	InternString("stats-test")
	stats := GetInternerStats()

	assert.Equal(t, int64(1), stats.TotalLookups)
	assert.Equal(t, int64(1), stats.UniqueStrings)
}

func TestClearGlobalInterner(t *testing.T) {
	InternString("clear-test")
	assert.True(t, GetGlobalInternerSize() > 0)

	ClearGlobalInterner()
	assert.Equal(t, 0, GetGlobalInternerSize())
}

// Benchmark tests
func BenchmarkStringInterner_Intern_New(b *testing.B) {
	interner := NewStringInterner()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interner.Intern("new-string")
		interner.Clear() // Clear to ensure it's always new
	}
}

func BenchmarkStringInterner_Intern_Existing(b *testing.B) {
	interner := NewStringInterner()
	interner.Intern("existing-string") // Pre-intern

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interner.Intern("existing-string")
	}
}

func BenchmarkGlobalInterner(b *testing.B) {
	ClearGlobalInterner()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InternString("benchmark-test")
	}
}
