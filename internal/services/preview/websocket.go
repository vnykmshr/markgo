// Package preview provides WebSocket functionality for real-time blog preview updates.
// It manages WebSocket connections, session handling, and broadcast functionality
// for live preview features in the MarkGo blog engine.
package preview

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func (s *Service) getUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return false
			}

			originURL, err := url.Parse(origin)
			if err != nil {
				return false
			}

			// Allow localhost and configured base URL
			allowedHosts := []string{"localhost", "127.0.0.1"}
			if s.baseURL != "" {
				if baseURL, err := url.Parse(s.baseURL); err == nil {
					allowedHosts = append(allowedHosts, baseURL.Host)
				}
			}

			for _, host := range allowedHosts {
				if strings.Contains(originURL.Host, host) {
					return true
				}
			}

			return false
		},
	}
}

func (s *Service) registerWebSocketClientImpl(sessionID string, w http.ResponseWriter, r *http.Request) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify session exists
	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("preview session not found")
	}

	// Upgrade connection to WebSocket
	upgrader := s.getUpgrader()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade failed", "error", err)
		return fmt.Errorf("websocket upgrade failed: %w", err)
	}

	// Add to clients list
	if s.clients[sessionID] == nil {
		s.clients[sessionID] = make([]*websocket.Conn, 0)
	}
	s.clients[sessionID] = append(s.clients[sessionID], conn)

	// Update session stats
	session.ClientCount = len(s.clients[sessionID])
	session.LastAccessed = time.Now()

	s.logger.Info("WebSocket client connected",
		"session_id", sessionID,
		"client_count", session.ClientCount,
		"remote_addr", r.RemoteAddr)

	// Handle client in goroutine
	go s.handleWebSocketClient(sessionID, conn)

	return nil
}

func (s *Service) handleWebSocketClient(sessionID string, conn *websocket.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Debug("Failed to close WebSocket connection", "error", err)
		}
		s.removeWebSocketClient(sessionID, conn)
	}()

	// Send initial connection confirmation
	if err := conn.WriteJSON(map[string]interface{}{
		"type":       "connected",
		"session_id": sessionID,
		"timestamp":  time.Now().Unix(),
	}); err != nil {
		s.logger.Debug("Failed to send connection confirmation", "error", err)
		return
	}

	// Handle incoming messages (keep-alive, ping, etc.)
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Debug("WebSocket connection closed unexpectedly", "error", err)
			}
			break
		}

		// Handle different message types
		switch messageType {
		case websocket.TextMessage:
			s.handleTextMessage(sessionID, conn, message)
		case websocket.PingMessage:
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				s.logger.Debug("Failed to send pong", "error", err)
				return // Exit the function instead of just breaking the switch
			}
		}
	}
}

func (s *Service) handleTextMessage(sessionID string, conn *websocket.Conn, message []byte) {
	s.logger.Debug("Received WebSocket message",
		"session_id", sessionID,
		"message", string(message))

	// For now, just echo back for testing
	response := map[string]interface{}{
		"type":      "echo",
		"message":   string(message),
		"timestamp": time.Now().Unix(),
	}

	if err := conn.WriteJSON(response); err != nil {
		s.logger.Debug("Failed to send echo response", "error", err)
	}
}

func (s *Service) removeWebSocketClient(sessionID string, targetConn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conns, exists := s.clients[sessionID]
	if !exists {
		return
	}

	// Remove the specific connection
	filtered := make([]*websocket.Conn, 0, len(conns))
	for _, conn := range conns {
		if conn != targetConn {
			filtered = append(filtered, conn)
		}
	}

	s.clients[sessionID] = filtered

	// Update session stats
	if session, exists := s.sessions[sessionID]; exists {
		session.ClientCount = len(filtered)
	}

	s.logger.Debug("WebSocket client disconnected",
		"session_id", sessionID,
		"remaining_clients", len(filtered))
}

func (s *Service) BroadcastReload(sessionID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conns, exists := s.clients[sessionID]
	if !exists || len(conns) == 0 {
		return nil // No clients to notify
	}

	message := map[string]interface{}{
		"type":      "reload",
		"timestamp": time.Now().Unix(),
	}

	failedConns := 0
	for _, conn := range conns {
		if err := conn.WriteJSON(message); err != nil {
			s.logger.Debug("Failed to broadcast reload", "error", err)
			failedConns++
		}
	}

	s.logger.Debug("Broadcasted reload message",
		"session_id", sessionID,
		"clients_notified", len(conns)-failedConns,
		"failed", failedConns)

	return nil
}
