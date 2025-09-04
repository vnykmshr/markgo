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

// TagInterner specialized interner for article tags
type TagInterner struct {
	*StringInterner
	tagMap map[string]uint16 // Map strings to compact IDs
	idMap  map[uint16]string // Map IDs back to strings
	nextID uint16
}

// NewTagInterner creates a new tag interner with ID mapping
func NewTagInterner() *TagInterner {
	return &TagInterner{
		StringInterner: NewStringInterner(),
		tagMap:         make(map[string]uint16),
		idMap:          make(map[uint16]string),
		nextID:         1,
	}
}

// InternTag interns a tag and returns both the interned string and a unique ID
func (ti *TagInterner) InternTag(tag string) (string, uint16) {
	if tag == "" {
		return tag, 0
	}

	ti.mu.Lock()
	defer ti.mu.Unlock()

	// Check if we already have this tag
	if id, exists := ti.tagMap[tag]; exists {
		return ti.idMap[id], id
	}

	// Intern the string and assign ID
	interned := ti.StringInterner.Intern(tag)
	id := ti.nextID
	ti.nextID++

	ti.tagMap[tag] = id
	ti.idMap[id] = interned

	return interned, id
}

// GetTagByID returns the tag string for a given ID
func (ti *TagInterner) GetTagByID(id uint16) string {
	if id == 0 {
		return ""
	}

	ti.mu.RLock()
	defer ti.mu.RUnlock()
	return ti.idMap[id]
}

// GetTagID returns the ID for a given tag
func (ti *TagInterner) GetTagID(tag string) uint16 {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	return ti.tagMap[tag]
}

// InternTagSlice interns a slice of tags and returns both string and ID slices
func (ti *TagInterner) InternTagSlice(tags []string) ([]string, []uint16) {
	if len(tags) == 0 {
		return tags, nil
	}

	internedTags := make([]string, len(tags))
	tagIDs := make([]uint16, len(tags))

	for i, tag := range tags {
		internedTag, id := ti.InternTag(tag)
		internedTags[i] = internedTag
		tagIDs[i] = id
	}

	return internedTags, tagIDs
}
