package services

import (
	"fmt"
	"strings"
	"time"
)

// DemonstrateCacheCompression shows the cache compression functionality
func DemonstrateCacheCompression() {
	fmt.Println("=== Cache Compression Demonstration ===")
	
	// Create a cache with compression enabled
	cache := NewCacheService(time.Hour, 100)
	defer cache.Stop()
	
	// Set compression threshold to 100 bytes so we can see compression in action
	cache.SetCompressionSettings(true, 100)
	
	// Test with a small string (should not be compressed)
	smallData := "Hello, World!"
	cache.Set("small", smallData, 0)
	
	// Test with a large string (should be compressed)
	largeData := strings.Repeat("This is a test string that will be compressed because it's long enough. ", 50)
	cache.Set("large", largeData, 0)
	
	// Test with a large byte slice (should be compressed)
	largeBytes := make([]byte, 5000)
	for i := range largeBytes {
		largeBytes[i] = byte(i % 256)
	}
	cache.Set("bytes", largeBytes, 0)
	
	// Test with a complex object (should use gob compression)
	complexData := map[string]interface{}{
		"name":  "MarkGo Cache",
		"type":  "In-Memory Cache with Compression",
		"features": []string{
			"Gzip compression for large values",
			"Memory-aware eviction policies", 
			"String and byte slice optimization",
			"Complex object support via gob encoding",
		},
		"stats": map[string]int{
			"version": 2,
			"maxSize": 1000,
		},
	}
	cache.Set("complex", complexData, 0)
	
	// Retrieve all values to verify they work
	fmt.Println("\nVerifying compressed data retrieval:")
	
	if val, found := cache.Get("small"); found {
		fmt.Printf("✓ Small data: %s\n", val)
	}
	
	if val, found := cache.Get("large"); found {
		retrieved := val.(string)
		fmt.Printf("✓ Large data: %d characters retrieved successfully\n", len(retrieved))
		fmt.Printf("  Starts with: %s...\n", retrieved[:50])
	}
	
	if val, found := cache.Get("bytes"); found {
		retrievedBytes := val.([]byte)
		fmt.Printf("✓ Byte data: %d bytes retrieved successfully\n", len(retrievedBytes))
		fmt.Printf("  First few bytes: %v...\n", retrievedBytes[:10])
	}
	
	if val, found := cache.Get("complex"); found {
		complexResult := val.(map[string]interface{})
		fmt.Printf("✓ Complex data: %s\n", complexResult["name"])
	}
	
	// Show cache statistics
	stats := cache.Stats()
	fmt.Printf("\n=== Cache Statistics ===\n")
	fmt.Printf("Total items: %v\n", stats["total_items"])
	fmt.Printf("Compressed items: %v\n", stats["compressed_items"])
	fmt.Printf("Hit ratio: %.2f\n", stats["hit_ratio"])
	fmt.Printf("Compression ratio: %.2f\n", stats["compression_ratio"])
	fmt.Printf("Memory usage: %v bytes\n", stats["memory_usage"])
	fmt.Printf("Max memory: %v bytes\n", stats["max_memory"])
	fmt.Printf("Compression enabled: %v\n", stats["enable_compression"])
	fmt.Printf("Compression threshold: %v bytes\n", stats["compression_threshold"])
	
	fmt.Println("\n✅ Cache compression demonstration completed successfully!")
}