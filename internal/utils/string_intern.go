package utils

import (
	"sync"
)

// StringInterner provides string interning to reduce memory usage
// by storing only one copy of frequently repeated strings
type StringInterner struct {
	mu      sync.RWMutex
	strings map[string]string
	stats   InternerStats
}

// InternerStats tracks interning statistics
type InternerStats struct {
	TotalLookups  int64
	HitCount      int64
	MissCount     int64
	UniqueStrings int64
	MemorySaved   int64 // Estimated bytes saved
}

// Global interner for application-wide string interning
var globalInterner = NewStringInterner()

// NewStringInterner creates a new string interner
func NewStringInterner() *StringInterner {
	return &StringInterner{
		strings: make(map[string]string),
	}
}

// Intern returns the canonical instance of the string, reducing memory usage
// for frequently repeated strings like tags, categories, and authors
func (si *StringInterner) Intern(s string) string {
	if s == "" {
		return s
	}

	// Try read lock first (common case)
	si.mu.RLock()
	if interned, exists := si.strings[s]; exists {
		si.stats.TotalLookups++
		si.stats.HitCount++
		si.mu.RUnlock()
		return interned
	}
	si.mu.RUnlock()

	// Need to intern the string
	si.mu.Lock()
	defer si.mu.Unlock()

	// Double-check pattern - another goroutine might have interned it
	if interned, exists := si.strings[s]; exists {
		si.stats.TotalLookups++
		si.stats.HitCount++
		return interned
	}

	// Intern the string
	si.strings[s] = s
	si.stats.TotalLookups++
	si.stats.MissCount++
	si.stats.UniqueStrings++
	si.stats.MemorySaved += int64(len(s)) // Rough estimate

	return s
}

// InternSlice interns all strings in a slice, returning a new slice with interned strings
func (si *StringInterner) InternSlice(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}

	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = si.Intern(s)
	}
	return result
}

// Stats returns current interning statistics
func (si *StringInterner) Stats() InternerStats {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.stats
}

// Clear removes all interned strings (use carefully, may break existing references)
func (si *StringInterner) Clear() {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.strings = make(map[string]string)
	si.stats = InternerStats{}
}

// Size returns the number of unique interned strings
func (si *StringInterner) Size() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return len(si.strings)
}

// Global convenience functions using the global interner

// InternString interns a string using the global interner
func InternString(s string) string {
	return globalInterner.Intern(s)
}

// InternStringSlice interns all strings in a slice using the global interner
func InternStringSlice(slice []string) []string {
	return globalInterner.InternSlice(slice)
}

// GetInternerStats returns statistics from the global interner
func GetInternerStats() InternerStats {
	return globalInterner.Stats()
}

// ClearGlobalInterner clears the global interner (use carefully)
func ClearGlobalInterner() {
	globalInterner.Clear()
}

// GetGlobalInternerSize returns the size of the global interner
func GetGlobalInternerSize() int {
	return globalInterner.Size()
}

// SavedMemory estimates bytes saved through string interning
func (si *StringInterner) SavedMemory() int64 {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.stats.MemorySaved
}

// Compact performs string interner compaction (placeholder for future optimization)
func (si *StringInterner) Compact() {
	// Currently a no-op, but could implement advanced compaction logic
	// such as removing rarely accessed strings, reorganizing hash tables, etc.
}
