package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionStore_CreateAndValidate(t *testing.T) {
	store := NewSessionStore()

	token, err := store.Create("admin")
	require.NoError(t, err)
	assert.Len(t, token, 64) // 32 bytes = 64 hex chars

	username, valid := store.Validate(token)
	assert.True(t, valid)
	assert.Equal(t, "admin", username)
}

func TestSessionStore_ValidateEmpty(t *testing.T) {
	store := NewSessionStore()

	_, valid := store.Validate("")
	assert.False(t, valid)
}

func TestSessionStore_ValidateUnknownToken(t *testing.T) {
	store := NewSessionStore()

	_, valid := store.Validate("nonexistent")
	assert.False(t, valid)
}

func TestSessionStore_Delete(t *testing.T) {
	store := NewSessionStore()

	token, err := store.Create("admin")
	require.NoError(t, err)

	store.Delete(token)

	_, valid := store.Validate(token)
	assert.False(t, valid)
}

func TestSessionStore_Expired(t *testing.T) {
	store := NewSessionStore()

	// Manually store an expired session
	store.sessions.Store("expired-token", &session{
		username:  "admin",
		expiresAt: time.Now().Add(-1 * time.Hour),
	})

	_, valid := store.Validate("expired-token")
	assert.False(t, valid)

	// Confirm it was cleaned up
	_, exists := store.sessions.Load("expired-token")
	assert.False(t, exists)
}

func TestSessionAuth_ValidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewSessionStore()
	token, _ := store.Create("admin")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin", http.NoBody)
	c.Request.AddCookie(&http.Cookie{Name: "_session", Value: token})

	handler := SessionAuth(store)
	handler(c)

	assert.False(t, c.IsAborted())
	user, exists := c.Get("admin_user")
	assert.True(t, exists)
	assert.Equal(t, "admin", user)
}

func TestSessionAuth_NoSession_GET(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewSessionStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin", http.NoBody)

	handler := SessionAuth(store)
	handler(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?next=")
}

func TestSessionAuth_NoSession_POST(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewSessionStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/cache/clear", http.NoBody)

	handler := SessionAuth(store)
	handler(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSessionAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewSessionStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/compose", http.NoBody)
	c.Request.AddCookie(&http.Cookie{Name: "_session", Value: "bogus"})

	handler := SessionAuth(store)
	handler(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusFound, w.Code)
}

// --- SoftSessionAuth tests ---

func TestSoftSessionAuth_ValidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewSessionStore()
	token, _ := store.Create("admin")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/compose", http.NoBody)
	c.Request.AddCookie(&http.Cookie{Name: "_session", Value: token})

	handler := SoftSessionAuth(store, false)
	handler(c)

	assert.False(t, c.IsAborted())
	user, exists := c.Get("admin_user")
	assert.True(t, exists)
	assert.Equal(t, "admin", user)
	authenticated, exists := c.Get("authenticated")
	assert.True(t, exists)
	assert.Equal(t, true, authenticated)
	_, authRequired := c.Get("auth_required")
	assert.False(t, authRequired)
}

func TestSoftSessionAuth_NoSession_GET(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewSessionStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/compose", http.NoBody)

	handler := SoftSessionAuth(store, false)
	handler(c)

	// Should NOT abort — allows handler to run
	assert.False(t, c.IsAborted())

	// Should set auth_required
	authRequired, exists := c.Get("auth_required")
	assert.True(t, exists)
	assert.Equal(t, true, authRequired)

	// Should generate CSRF token
	csrfToken, exists := c.Get("csrf_token")
	assert.True(t, exists)
	assert.NotEmpty(t, csrfToken)

	// Should NOT set authenticated
	_, exists = c.Get("authenticated")
	assert.False(t, exists)
}

func TestSoftSessionAuth_NoSession_POST(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewSessionStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/compose", http.NoBody)

	handler := SoftSessionAuth(store, false)
	handler(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSoftSessionAuth_InvalidToken_GET(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewSessionStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin", http.NoBody)
	c.Request.AddCookie(&http.Cookie{Name: "_session", Value: "bogus"})

	handler := SoftSessionAuth(store, false)
	handler(c)

	// Should NOT abort — allows handler to run
	assert.False(t, c.IsAborted())

	// Should set auth_required (stale cookie cleared, treated as unauthenticated)
	authRequired, exists := c.Get("auth_required")
	assert.True(t, exists)
	assert.Equal(t, true, authRequired)
}

func TestSoftSessionAuth_HEAD_Request(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewSessionStore()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodHead, "/compose", http.NoBody)

	handler := SoftSessionAuth(store, false)
	handler(c)

	// HEAD should behave same as GET — not abort
	assert.False(t, c.IsAborted())
	authRequired, exists := c.Get("auth_required")
	assert.True(t, exists)
	assert.Equal(t, true, authRequired)
}
