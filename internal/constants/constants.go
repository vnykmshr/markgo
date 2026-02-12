// Package constants provides application-wide constants for the MarkGo blog engine.
package constants

// Application metadata
const (
	AppName    = "MarkGo"
	AppVersion = "v3.1.0"
)

// Build-time variables injected via ldflags
var (
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// SupportedMarkdownExtensions lists recognized markdown file extensions.
var SupportedMarkdownExtensions = []string{".md", ".markdown", ".mdown", ".mkd"}
