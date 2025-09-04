package utils

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// ResponseWriterPool provides pooled response writers for middleware efficiency
type ResponseWriterPool struct {
	pool sync.Pool
}

// NewResponseWriterPool creates a new response writer pool
func NewResponseWriterPool() *ResponseWriterPool {
	return &ResponseWriterPool{
		pool: sync.Pool{
			New: func() any {
				return &PooledResponseWriter{
					body:    bytes.NewBuffer(make([]byte, 0, 4096)), // Pre-allocate 4KB
					headers: make(http.Header, 20),                  // Pre-allocate common headers
				}
			},
		},
	}
}

// PooledResponseWriter is a response writer that can be pooled
type PooledResponseWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	headers http.Header
	status  int
	written bool
}

// GetWriter gets a pooled response writer
func (p *ResponseWriterPool) GetWriter(originalWriter gin.ResponseWriter) *PooledResponseWriter {
	writer := p.pool.Get().(*PooledResponseWriter)
	writer.ResponseWriter = originalWriter
	writer.body.Reset()
	writer.status = http.StatusOK
	writer.written = false

	// Clear headers but keep allocated memory
	for k := range writer.headers {
		delete(writer.headers, k)
	}

	return writer
}

// PutWriter returns a response writer to the pool
func (p *ResponseWriterPool) PutWriter(writer *PooledResponseWriter) {
	if writer != nil && writer.body.Cap() <= 65536 { // Don't pool huge buffers
		writer.ResponseWriter = nil // Clear reference
		p.pool.Put(writer)
	}
}

// Header implements http.ResponseWriter
func (w *PooledResponseWriter) Header() http.Header {
	if len(w.headers) == 0 {
		// Copy original headers on first access
		originalHeaders := w.ResponseWriter.Header()
		for key, values := range originalHeaders {
			w.headers[key] = values
		}
	}
	return w.headers
}

// WriteHeader implements http.ResponseWriter
func (w *PooledResponseWriter) WriteHeader(status int) {
	w.status = status
	w.written = true
}

// Write implements http.ResponseWriter
func (w *PooledResponseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(data)
}

// Flush flushes buffered data to the original writer
func (w *PooledResponseWriter) Flush() {
	if w.written {
		// Copy headers to original writer
		originalHeader := w.ResponseWriter.Header()
		for key, values := range w.headers {
			originalHeader[key] = values
		}

		// Write status and body
		w.ResponseWriter.WriteHeader(w.status)
		if w.body.Len() > 0 {
			_, _ = w.ResponseWriter.Write(w.body.Bytes())
		}
	}
}

// Status returns the status code
func (w *PooledResponseWriter) Status() int {
	return w.status
}

// Size returns the size of the written data
func (w *PooledResponseWriter) Size() int {
	return w.body.Len()
}

// Written returns whether WriteHeader has been called
func (w *PooledResponseWriter) Written() bool {
	return w.written
}

// Global response writer pool
var globalResponseWriterPool = NewResponseWriterPool()

// GetGlobalResponseWriterPool returns the global response writer pool
func GetGlobalResponseWriterPool() *ResponseWriterPool {
	return globalResponseWriterPool
}

// HeaderPool provides pooled header maps for efficient header manipulation
type HeaderPool struct {
	pool sync.Pool
}

// NewHeaderPool creates a new header pool
func NewHeaderPool() *HeaderPool {
	return &HeaderPool{
		pool: sync.Pool{
			New: func() any {
				return make(http.Header, 20) // Pre-allocate for common headers
			},
		},
	}
}

// GetHeaders gets a pooled header map
func (p *HeaderPool) GetHeaders() http.Header {
	headers := p.pool.Get().(http.Header)
	// Clear headers but keep allocated memory
	for k := range headers {
		delete(headers, k)
	}
	return headers
}

// PutHeaders returns headers to the pool
func (p *HeaderPool) PutHeaders(headers http.Header) {
	if headers != nil && len(headers) <= 50 { // Don't pool huge header maps
		p.pool.Put(headers)
	}
}

// WithHeaders provides a convenient way to use pooled headers
func (p *HeaderPool) WithHeaders(fn func(http.Header)) {
	headers := p.GetHeaders()
	defer p.PutHeaders(headers)
	fn(headers)
}

// Global header pool
var globalHeaderPool = NewHeaderPool()

// GetGlobalHeaderPool returns the global header pool
func GetGlobalHeaderPool() *HeaderPool {
	return globalHeaderPool
}

// CompressedResponsePool manages pooled compressed response handling
type CompressedResponsePool struct {
	bufferPool sync.Pool
}

// NewCompressedResponsePool creates a compressed response pool
func NewCompressedResponsePool() *CompressedResponsePool {
	return &CompressedResponsePool{
		bufferPool: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(make([]byte, 0, 8192)) // 8KB for compressed data
			},
		},
	}
}

// GetBuffer gets a buffer for compressed responses
func (p *CompressedResponsePool) GetBuffer() *bytes.Buffer {
	buf := p.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns a buffer to the pool
func (p *CompressedResponsePool) PutBuffer(buf *bytes.Buffer) {
	if buf != nil && buf.Cap() <= 65536 { // Don't pool huge buffers
		p.bufferPool.Put(buf)
	}
}

// Global compressed response pool
var globalCompressedResponsePool = NewCompressedResponsePool()

// GetGlobalCompressedResponsePool returns the global compressed response pool
func GetGlobalCompressedResponsePool() *CompressedResponsePool {
	return globalCompressedResponsePool
}

// MiddlewareContextPool provides pooled contexts for middleware chains
type MiddlewareContextPool struct {
	pool sync.Pool
}

// MiddlewareContext holds reusable middleware state
type MiddlewareContext struct {
	Values   map[string]any
	Headers  map[string]string
	Metadata map[string]any
}

// NewMiddlewareContextPool creates a middleware context pool
func NewMiddlewareContextPool() *MiddlewareContextPool {
	return &MiddlewareContextPool{
		pool: sync.Pool{
			New: func() any {
				return &MiddlewareContext{
					Values:   make(map[string]any, 10),
					Headers:  make(map[string]string, 20),
					Metadata: make(map[string]any, 5),
				}
			},
		},
	}
}

// GetContext gets a pooled middleware context
func (p *MiddlewareContextPool) GetContext() *MiddlewareContext {
	ctx := p.pool.Get().(*MiddlewareContext)

	// Clear maps but keep allocated memory
	for k := range ctx.Values {
		delete(ctx.Values, k)
	}
	for k := range ctx.Headers {
		delete(ctx.Headers, k)
	}
	for k := range ctx.Metadata {
		delete(ctx.Metadata, k)
	}

	return ctx
}

// PutContext returns a middleware context to the pool
func (p *MiddlewareContextPool) PutContext(ctx *MiddlewareContext) {
	if ctx != nil &&
		len(ctx.Values) <= 50 &&
		len(ctx.Headers) <= 50 &&
		len(ctx.Metadata) <= 20 { // Size limits to prevent memory leaks
		p.pool.Put(ctx)
	}
}

// WithContext provides a convenient way to use pooled middleware context
func (p *MiddlewareContextPool) WithContext(fn func(*MiddlewareContext)) {
	ctx := p.GetContext()
	defer p.PutContext(ctx)
	fn(ctx)
}

// Global middleware context pool
var globalMiddlewareContextPool = NewMiddlewareContextPool()

// GetGlobalMiddlewareContextPool returns the global middleware context pool
func GetGlobalMiddlewareContextPool() *MiddlewareContextPool {
	return globalMiddlewareContextPool
}

// Cleanup methods for memory leak prevention

// Cleanup clears all pooled response writers (for memory leak prevention)
func (p *ResponseWriterPool) Cleanup() {
	// Force cleanup of all pooled writers by creating a new pool
	p.pool = sync.Pool{
		New: func() any {
			return &PooledResponseWriter{
				body:    bytes.NewBuffer(make([]byte, 0, 4096)),
				headers: make(http.Header, 20),
			}
		},
	}
}

// GetStats returns statistics for the response writer pool
func (p *ResponseWriterPool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"type": "ResponseWriterPool",
		"note": "sync.Pool doesn't expose internal metrics",
	}
}

// Cleanup clears all pooled headers (for memory leak prevention)
func (p *HeaderPool) Cleanup() {
	p.pool = sync.Pool{
		New: func() any {
			return make(http.Header, 20)
		},
	}
}

// GetStats returns statistics for the header pool
func (p *HeaderPool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"type": "HeaderPool",
		"note": "sync.Pool doesn't expose internal metrics",
	}
}

// Cleanup clears all compressed response buffers (for memory leak prevention)
func (p *CompressedResponsePool) Cleanup() {
	p.bufferPool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, 8192))
		},
	}
}

// GetStats returns statistics for the compressed response pool
func (p *CompressedResponsePool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"type": "CompressedResponsePool",
		"note": "sync.Pool doesn't expose internal metrics",
	}
}

// Cleanup clears all middleware contexts (for memory leak prevention)
func (p *MiddlewareContextPool) Cleanup() {
	p.pool = sync.Pool{
		New: func() any {
			return &MiddlewareContext{
				Values:   make(map[string]any, 10),
				Headers:  make(map[string]string, 20),
				Metadata: make(map[string]any, 5),
			}
		},
	}
}

// GetStats returns statistics for the middleware context pool
func (p *MiddlewareContextPool) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"type": "MiddlewareContextPool",
		"note": "sync.Pool doesn't expose internal metrics",
	}
}
