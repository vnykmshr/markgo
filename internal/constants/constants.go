// Package constants provides application-wide constants for the MarkGo blog engine.
package constants

// AppName is the display name of the application.
const AppName = "MarkGo"

// Build-time variables injected via ldflags
var (
	AppVersion = "dev"
	GitCommit  = "unknown"
	BuildTime  = "unknown"
)

// SupportedMarkdownExtensions lists recognized markdown file extensions.
var SupportedMarkdownExtensions = []string{".md", ".markdown", ".mdown", ".mkd"}
