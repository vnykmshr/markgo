package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupportedMarkdownExtensions(t *testing.T) {
	assert.Contains(t, SupportedMarkdownExtensions, ".md")
	assert.Contains(t, SupportedMarkdownExtensions, ".markdown")
	assert.Len(t, SupportedMarkdownExtensions, 4)
}

func TestApplicationMetadata(t *testing.T) {
	assert.Equal(t, "MarkGo", AppName)
	assert.Regexp(t, `^v\d+\.\d+\.\d+$`, AppVersion)
}
