package preview

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"

	"github.com/vnykmshr/markgo/internal/config"
	"github.com/vnykmshr/markgo/internal/services"
)

// Service provides preview functionality.
type Service struct {
	articleService  services.ArticleServiceInterface
	templateService services.TemplateServiceInterface
	watcher         *fsnotify.Watcher
	sessions        map[string]*Session
	clients         map[string][]*websocket.Conn
	mu              sync.RWMutex
	logger          *slog.Logger
	config          *config.PreviewConfig
	baseURL         string
	articlesPath    string
	running         bool
}

// Session represents a preview session.
type Session = services.PreviewSession

// Stats represents preview statistics.
type Stats = services.PreviewStats

// NewService creates a new preview service instance.
func NewService(
	articleService services.ArticleServiceInterface,
	templateService services.TemplateServiceInterface,
	cfg *config.PreviewConfig,
	logger *slog.Logger,
	articlesPath string,
) (*Service, error) {
	if articleService == nil {
		return nil, fmt.Errorf("article service is required")
	}
	if templateService == nil {
		return nil, fmt.Errorf("template service is required")
	}
	if cfg == nil {
		return nil, fmt.Errorf("preview config is required")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	baseURL := fmt.Sprintf("http://localhost:%d", cfg.Port)
	if cfg.BaseURL != "" {
		baseURL = cfg.BaseURL
	}

	service := &Service{
		articleService:  articleService,
		templateService: templateService,
		watcher:         watcher,
		sessions:        make(map[string]*Session),
		clients:         make(map[string][]*websocket.Conn),
		logger:          logger,
		config:          cfg,
		baseURL:         baseURL,
		articlesPath:    articlesPath,
	}

	return service, nil
}

// Start starts the preview service.
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("preview service already running")
	}

	// Add articles directory to file watcher
	if err := s.watcher.Add(s.articlesPath); err != nil {
		s.logger.Warn("Failed to watch articles directory", "path", s.articlesPath, "error", err)
	} else {
		s.logger.Info("Watching articles directory for changes", "path", s.articlesPath)
	}

	go s.watchFiles()
	go s.cleanupSessions()

	s.running = true
	s.logger.Info("Preview service started", "base_url", s.baseURL)
	return nil
}

// Stop stops the preview service.
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("preview service not running")
	}

	if err := s.watcher.Close(); err != nil {
		s.logger.Warn("Error closing file watcher", "error", err)
	}

	// Close all WebSocket connections
	for sessionID, conns := range s.clients {
		for _, conn := range conns {
			if err := conn.Close(); err != nil {
				s.logger.Warn("Error closing WebSocket connection",
					"session_id", sessionID, "error", err)
			}
		}
	}

	s.running = false
	s.logger.Info("Preview service stopped")
	return nil
}

// CreateSession creates a new preview session for the given article slug.
func (s *Service) CreateSession(articleSlug string) (*Session, error) {
	if articleSlug == "" {
		return nil, fmt.Errorf("article slug is required")
	}

	// Check if article exists as draft
	_, err := s.articleService.GetDraftBySlug(articleSlug)
	if err != nil {
		return nil, fmt.Errorf("draft article not found: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check session limit
	if len(s.sessions) >= s.config.MaxSessions {
		return nil, fmt.Errorf("maximum preview sessions exceeded (%d)", s.config.MaxSessions)
	}

	sessionID := generateSessionID()
	authToken := generateAuthToken()

	session := &Session{
		ID:           sessionID,
		ArticleSlug:  articleSlug,
		URL:          fmt.Sprintf("%s/preview/%s", s.baseURL, sessionID),
		AuthToken:    authToken,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}

	s.sessions[sessionID] = session
	s.clients[sessionID] = make([]*websocket.Conn, 0)

	s.logger.Info("Preview session created",
		"session_id", sessionID,
		"article_slug", articleSlug,
		"url", session.URL)

	return session, nil
}

// GetSession retrieves a preview session by ID.
func (s *Service) GetSession(sessionID string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("preview session not found")
	}

	// Update last accessed time
	session.LastAccessed = time.Now()
	return session, nil
}

// DeleteSession removes a preview session.
func (s *Service) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("preview session not found")
	}

	// Close all connections for this session
	if conns, exists := s.clients[sessionID]; exists {
		for _, conn := range conns {
			if err := conn.Close(); err != nil {
				s.logger.Warn("Error closing connection", "error", err)
			}
		}
		delete(s.clients, sessionID)
	}

	delete(s.sessions, sessionID)

	s.logger.Info("Preview session deleted",
		"session_id", sessionID,
		"article_slug", session.ArticleSlug)

	return nil
}

// RegisterWebSocketClient registers a new WebSocket client for the given session.
func (s *Service) RegisterWebSocketClient(sessionID string, w http.ResponseWriter, r *http.Request) error {
	return s.registerWebSocketClientImpl(sessionID, w, r)
}

// IsRunning returns whether the preview service is currently running.
func (s *Service) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetStats returns statistics about the preview service.
func (s *Service) GetStats() *Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalClients := 0
	for _, conns := range s.clients {
		totalClients += len(conns)
	}

	// Count watched paths
	watchedPaths := 1 // articles directory
	if s.running {
		// Add template directory if exists
		watchedPaths = 1
	}

	return &Stats{
		ActiveSessions: len(s.sessions),
		TotalClients:   totalClients,
		FilesWatched:   watchedPaths,
	}
}

func (s *Service) watchFiles() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				s.handleFileChange(event.Name)
			}

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			s.logger.Error("File watcher error", "error", err)
		}
	}
}

func (s *Service) handleFileChange(filepath string) {
	s.logger.Debug("File changed", "path", filepath)

	// Extract article slug from file path
	articleSlug := s.extractSlugFromPath(filepath)
	if articleSlug == "" {
		s.logger.Debug("Could not extract slug from path", "path", filepath)
		return
	}

	s.logger.Info("Article file changed, broadcasting reload",
		"path", filepath,
		"slug", articleSlug)

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find sessions that are previewing this specific article
	reloadedSessions := 0
	for sessionID, session := range s.sessions {
		if session.ArticleSlug == articleSlug {
			s.broadcastToSession(sessionID, "reload", map[string]interface{}{
				"article_slug": articleSlug,
				"file_path":    filepath,
				"reason":       "file_changed",
			})
			reloadedSessions++
		}
	}

	if reloadedSessions > 0 {
		s.logger.Info("Reloaded preview sessions",
			"article_slug", articleSlug,
			"sessions_reloaded", reloadedSessions)
	}
}

func (s *Service) broadcastToSession(sessionID, messageType string, data interface{}) {
	conns, exists := s.clients[sessionID]
	if !exists {
		return
	}

	message := map[string]interface{}{
		"type":      messageType,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	// Remove closed connections while broadcasting
	activeConns := make([]*websocket.Conn, 0, len(conns))

	for _, conn := range conns {
		if err := conn.WriteJSON(message); err != nil {
			s.logger.Debug("Failed to send WebSocket message", "error", err)
			if err := conn.Close(); err != nil {
				s.logger.Debug("Failed to close WebSocket connection", "error", err)
			}
		} else {
			activeConns = append(activeConns, conn)
		}
	}

	s.clients[sessionID] = activeConns
}

func (s *Service) cleanupSessions() {
	ticker := time.NewTicker(s.config.SessionTimeout / 4)
	defer ticker.Stop()

	for range ticker.C {
		s.performCleanup()
	}
}

func (s *Service) performCleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expired := make([]string, 0)

	for sessionID, session := range s.sessions {
		if now.Sub(session.LastAccessed) > s.config.SessionTimeout {
			expired = append(expired, sessionID)
		}
	}

	for _, sessionID := range expired {
		s.logger.Info("Cleaning up expired session", "session_id", sessionID)

		// Close connections
		if conns, exists := s.clients[sessionID]; exists {
			for _, conn := range conns {
				if err := conn.Close(); err != nil {
					s.logger.Debug("Failed to close WebSocket connection", "error", err)
				}
			}
			delete(s.clients, sessionID)
		}

		delete(s.sessions, sessionID)
	}

	if len(expired) > 0 {
		s.logger.Info("Cleaned up expired sessions", "count", len(expired))
	}
}

func generateSessionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based ID if crypto rand fails
		return "session_" + hex.EncodeToString([]byte(time.Now().String()))[:16]
	}
	return "session_" + hex.EncodeToString(bytes)[:16]
}

func generateAuthToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based token if crypto rand fails
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(bytes)
}

// extractSlugFromPath converts a file path to an article slug
func (s *Service) extractSlugFromPath(filepath string) string {
	// Handle different path separators
	filepath = strings.ReplaceAll(filepath, "\\", "/")

	// Extract filename from path
	filename := filepath
	if idx := strings.LastIndex(filepath, "/"); idx != -1 {
		filename = filepath[idx+1:]
	}

	// Remove .md extension
	if !strings.HasSuffix(filename, ".md") {
		return ""
	}
	filename = strings.TrimSuffix(filename, ".md")

	// Convert filename to slug format
	// Remove date prefixes like "2024-01-01-" if present
	if len(filename) > 11 && filename[4] == '-' && filename[7] == '-' && filename[10] == '-' {
		filename = filename[11:]
	}

	// Convert to kebab-case slug
	slug := strings.ToLower(filename)
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove any invalid characters
	validSlug := ""
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			validSlug += string(r)
		}
	}

	// Clean up multiple dashes
	for strings.Contains(validSlug, "--") {
		validSlug = strings.ReplaceAll(validSlug, "--", "-")
	}
	validSlug = strings.Trim(validSlug, "-")

	return validSlug
}
