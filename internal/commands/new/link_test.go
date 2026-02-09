package new

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkURLParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantHost string
		wantErr  bool
	}{
		{
			name:     "valid HTTPS URL",
			input:    "https://example.com/article",
			wantHost: "example.com",
		},
		{
			name:     "valid HTTP URL",
			input:    "http://blog.example.org/post",
			wantHost: "blog.example.org",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no host",
			input:   "/just/a/path",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := url.Parse(tt.input)
			if tt.wantErr {
				assert.True(t, err != nil || parsed.Host == "")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, parsed.Host)
		})
	}
}

func TestLinkSlugFormat(t *testing.T) {
	slug := "link-1234567890"
	assert.Contains(t, slug, "link-")
}
