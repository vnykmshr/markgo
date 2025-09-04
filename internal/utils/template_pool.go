package utils

import (
	"sync"

	"github.com/gin-gonic/gin"
)

// TemplateDataPool provides a pool of gin.H maps to reduce allocations
type TemplateDataPool struct {
	pool sync.Pool
}

// NewTemplateDataPool creates a new template data pool
func NewTemplateDataPool() *TemplateDataPool {
	return &TemplateDataPool{
		pool: sync.Pool{
			New: func() any {
				return make(gin.H, 20) // Pre-allocate with reasonable capacity
			},
		},
	}
}

// Get retrieves a gin.H map from the pool
func (p *TemplateDataPool) Get() gin.H {
	return p.pool.Get().(gin.H)
}

// Put returns a gin.H map to the pool after clearing it
func (p *TemplateDataPool) Put(data gin.H) {
	// Clear the map but keep the allocated memory
	for k := range data {
		delete(data, k)
	}
	p.pool.Put(data)
}

// Global template data pool for application-wide use
var globalTemplateDataPool = NewTemplateDataPool()

// GetTemplateData gets a gin.H map from the global pool
func GetTemplateData() gin.H {
	return globalTemplateDataPool.Get()
}

// PutTemplateData returns a gin.H map to the global pool
func PutTemplateData(data gin.H) {
	globalTemplateDataPool.Put(data)
}

// WithTemplateData provides a convenient way to use pooled template data
func WithTemplateData(fn func(gin.H) gin.H) gin.H {
	data := GetTemplateData()
	defer PutTemplateData(data)

	return fn(data)
}

// TemplateDataBuilder helps build template data efficiently
type TemplateDataBuilder struct {
	data gin.H
	pool *TemplateDataPool
}

// NewTemplateDataBuilder creates a new template data builder
func NewTemplateDataBuilder() *TemplateDataBuilder {
	return &TemplateDataBuilder{
		data: GetTemplateData(),
		pool: globalTemplateDataPool,
	}
}

// Set sets a key-value pair in the template data
func (b *TemplateDataBuilder) Set(key string, value any) *TemplateDataBuilder {
	b.data[key] = value
	return b
}

// SetIf conditionally sets a key-value pair based on a condition
func (b *TemplateDataBuilder) SetIf(condition bool, key string, value any) *TemplateDataBuilder {
	if condition {
		b.data[key] = value
	}
	return b
}

// SetIfNotEmpty sets a key-value pair if the value is not empty/nil
func (b *TemplateDataBuilder) SetIfNotEmpty(key string, value any) *TemplateDataBuilder {
	if value != nil {
		switch v := value.(type) {
		case string:
			if v != "" {
				b.data[key] = value
			}
		case []any:
			if len(v) > 0 {
				b.data[key] = value
			}
		default:
			b.data[key] = value
		}
	}
	return b
}

// Merge merges another gin.H map into the current data
func (b *TemplateDataBuilder) Merge(other gin.H) *TemplateDataBuilder {
	for k, v := range other {
		b.data[k] = v
	}
	return b
}

// Build returns the built template data
func (b *TemplateDataBuilder) Build() gin.H {
	// Return a copy to prevent modifications after build
	result := make(gin.H, len(b.data))
	for k, v := range b.data {
		result[k] = v
	}

	// Return the pooled data back to the pool
	b.pool.Put(b.data)
	b.data = nil // Prevent further use

	return result
}

// Release manually releases the builder's data back to the pool (in case Build() is not called)
func (b *TemplateDataBuilder) Release() {
	if b.data != nil {
		b.pool.Put(b.data)
		b.data = nil
	}
}

// BaseTemplateData creates common template data that most pages need
func BaseTemplateData(title string, config any) *TemplateDataBuilder {
	return NewTemplateDataBuilder().
		Set("title", title).
		Set("config", config)
}

// ArticlePageData creates template data for article pages
func ArticlePageData(title string, config any, recent any) *TemplateDataBuilder {
	return BaseTemplateData(title, config).
		Set("recent", recent)
}

// ListPageData creates template data for list pages with pagination
func ListPageData(title string, config any, recent any, items any, pagination any) *TemplateDataBuilder {
	return ArticlePageData(title, config, recent).
		Set("articles", items).
		Set("pagination", pagination)
}
