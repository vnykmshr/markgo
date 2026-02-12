// Package web embeds templates and static assets into the binary.
package web

import "embed"

// Assets contains all embedded web templates and static files.
//
//go:embed all:templates all:static
var Assets embed.FS
