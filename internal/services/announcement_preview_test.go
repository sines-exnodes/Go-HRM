package services

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exnodes/hrm-api/internal/models"
)

func TestPlainTextPreview(t *testing.T) {
	exactly100 := strings.Repeat("界", 100)
	over100 := strings.Repeat("界", 101)

	tests := []struct {
		name     string
		input    string
		limit    int
		expected string
	}{
		{
			name:     "normalizes rich text",
			input:    "<p>Hello&nbsp;<strong>team</strong></p>\n next",
			limit:    100,
			expected: "Hello team next",
		},
		{
			name:     "returns empty for HTML-only content",
			input:    "<p><br></p>",
			limit:    100,
			expected: "",
		},
		{
			name:     "preserves exact limit",
			input:    exactly100,
			limit:    100,
			expected: exactly100,
		},
		{
			name:     "truncates Unicode by rune",
			input:    over100,
			limit:    100,
			expected: strings.Repeat("界", 99) + "…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, plainTextPreview(tt.input, tt.limit))
		})
	}
}

func TestToMobileBriefIncludesPlainTextDescriptionPreview(t *testing.T) {
	description := "<p>Hello&nbsp;<strong>team</strong></p><p>" + strings.Repeat("界", 100) + "</p>"
	item := (&AnnouncementService{}).toMobileBrief(&models.Announcement{
		Description: description,
	}, false)

	raw, err := json.Marshal(item)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(raw, &payload))
	assert.Equal(t, "Hello team "+strings.Repeat("界", 88)+"…", payload["description"])
}
