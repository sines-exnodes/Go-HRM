# Mobile Announcement Description Preview Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Return a plain-text announcement `description` preview of at most 100 Unicode characters from the mobile home and mobile list APIs.

**Architecture:** Extract the existing push-notification sanitizer into a private, limit-aware announcement preview helper. Use that helper with the existing 128-character push limit and a new 100-character mobile limit, then expose the derived value through the shared `MobileAnnouncementBrief` DTO.

**Tech Stack:** Go 1.24, Gin handlers, GORM repositories, standard-library `html`, `regexp`, and `strings`, `testify`, Swag v1.16.4.

## Global Constraints

- Mobile preview descriptions are plain text derived at response time; stored content is unchanged.
- Normalize tags, HTML entities, and whitespace before truncating.
- A truncated mobile preview contains 99 content characters plus `…`, for a total of 100 Unicode characters.
- Both `GET /api/v1/mobile/announcements` and `GET /api/v1/mobile/announcements/list` use the same preview field.
- The mobile detail response, visibility, ordering, and pagination remain unchanged.
- Keep push-notification previews at their existing 128-character limit.

---

### Task 1: Shared Plain-Text Preview Helper

**Files:**
- Create: `internal/services/announcement_preview.go`
- Create: `internal/services/announcement_preview_test.go`
- Modify: `internal/services/announcement_notifier.go:3-29,52`

**Interfaces:**
- Consumes: rich-text announcement description strings and a positive Unicode-character limit.
- Produces: `plainTextPreview(htmlContent string, maxRunes int) string`, a private helper shared by push and mobile projections.

- [ ] **Step 1: Write the failing helper test**

```go
package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
```

- [ ] **Step 2: Run the helper test and verify RED**

Run: `go test ./internal/services -run TestPlainTextPreview -count=1`

Expected: build failure containing `undefined: plainTextPreview`.

- [ ] **Step 3: Implement the shared helper**

Create `internal/services/announcement_preview.go`:

```go
package services

import (
	"html"
	"regexp"
	"strings"
)

var reHTMLTag = regexp.MustCompile(`<[^>]+>`)

func plainTextPreview(htmlContent string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}

	plainText := reHTMLTag.ReplaceAllString(htmlContent, " ")
	plainText = html.UnescapeString(plainText)
	plainText = strings.Join(strings.Fields(plainText), " ")
	runes := []rune(plainText)
	if len(runes) <= maxRunes {
		return plainText
	}
	return string(runes[:maxRunes-1]) + "…"
}
```

Remove the `html`, `regexp`, `strings`, and `unicode/utf8` imports, regex declarations, and `pushBody` helper from `internal/services/announcement_notifier.go`. Change the push request body to:

```go
Body: plainTextPreview(description, 128),
```

- [ ] **Step 4: Run the helper test and announcement tests**

Run: `go test ./internal/services -run TestPlainTextPreview -count=1`

Run: `go test ./internal/services -run TestAnnouncement -count=1`

Expected: PASS.

- [ ] **Step 5: Commit the helper extraction**

```bash
git add internal/services/announcement_preview.go internal/services/announcement_preview_test.go internal/services/announcement_notifier.go
git commit -m "refactor: share announcement preview sanitizer"
```

### Task 2: Mobile Description Preview Response

**Files:**
- Modify: `internal/dto/announcement.go:140-153`
- Modify: `internal/services/announcement_service.go:735-825`
- Modify: `internal/services/announcement_service_test.go:3-12,568-592`

**Interfaces:**
- Consumes: `plainTextPreview(htmlContent string, maxRunes int) string` from Task 1 and `models.Announcement.Description`.
- Produces: required `MobileAnnouncementBrief.Description string` serialized as JSON field `description`.

- [ ] **Step 1: Write the failing mobile response test**

Add `encoding/json` and `strings` to the imports in `internal/services/announcement_service_test.go`, then add:

```go
func TestAnnouncement_MobileBrief_IncludesPlainTextDescriptionPreview(t *testing.T) {
	skipIfNoDB(t)
	truncateAll(t)
	ctx := context.Background()

	svc, _ := newAnnouncementSvc(t)
	author, _ := makeEmpUser(t, "mobile-preview-author@example.com", "Author")
	reader, _ := makeEmpUser(t, "mobile-preview-reader@example.com", "Reader")
	pub := models.AnnouncementStatusPublished
	description := "<p>Hello&nbsp;<strong>team</strong></p><p>" + strings.Repeat("界", 100) + "</p>"

	_, err := svc.Create(ctx, author.ID, dto.AnnouncementCreate{
		Title:       "preview",
		Description: description,
		Status:      &pub,
	})
	require.NoError(t, err)

	items, err := svc.MobileBrief(ctx, reader.ID)
	require.NoError(t, err)
	require.Len(t, items, 1)

	raw, err := json.Marshal(items[0])
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(raw, &payload))
	assert.Equal(t, "Hello team "+strings.Repeat("界", 88)+"…", payload["description"])
}
```

- [ ] **Step 2: Run the mobile response test and verify RED**

Run: `go test ./internal/services -run TestAnnouncement_MobileBrief_IncludesPlainTextDescriptionPreview -count=1`

Expected: FAIL because `payload["description"]` is `nil`.

- [ ] **Step 3: Add the DTO field and service projection**

Update the DTO comment and add the required field:

```go
// MobileAnnouncementBrief is the compact projection used by mobile lists and
// the home-screen widget. Description is a plain-text, 100-character preview.
type MobileAnnouncementBrief struct {
	ID            uuid.UUID                 `json:"id"`
	Title         string                    `json:"title"`
	Description   string                    `json:"description"`
	Summary       *string                   `json:"summary,omitempty"`
	CoverImageURL *string                   `json:"cover_image_url,omitempty"`
	Status        models.AnnouncementStatus `json:"status"`
	Pinned        bool                      `json:"pinned"`
	PublishedAt   *time.Time                `json:"published_at,omitempty"`
	Labels        []AnnouncementLabelBrief  `json:"labels"`
	HasViewed     bool                      `json:"has_viewed"`
}
```

At the mobile-specific service section, define:

```go
const mobileDescriptionPreviewLimit = 100
```

Populate the field in `toMobileBrief`:

```go
Description: plainTextPreview(a.Description, mobileDescriptionPreviewLimit),
```

- [ ] **Step 4: Run focused service tests and verify GREEN**

Run: `go test ./internal/services -run TestAnnouncement_Mobile -count=1`

Run: `go test ./internal/services -run TestPlainTextPreview -count=1`

Expected: PASS.

- [ ] **Step 5: Commit the mobile response change**

```bash
git add internal/dto/announcement.go internal/services/announcement_service.go internal/services/announcement_service_test.go
git commit -m "fix: add description preview to mobile announcements"
```

### Task 3: API Documentation and Full Verification

**Files:**
- Modify: `internal/handlers/announcement_handler.go:260-288`
- Regenerate: `docs/swagger/docs.go`
- Regenerate: `docs/swagger/swagger.json`
- Regenerate: `docs/swagger/swagger.yaml`

**Interfaces:**
- Consumes: the `MobileAnnouncementBrief.description` contract from Task 2.
- Produces: Swagger descriptions that state the mobile preview is plain text and limited to 100 characters.

- [ ] **Step 1: Update the handler documentation**

Use these descriptions for `MobileBrief`:

```go
// @Description  Returns the latest 5 published announcements visible to the
// @Description  current user. Each item includes a plain-text description preview
// @Description  limited to 100 Unicode characters. Unpaginated.
```

Use these descriptions for `MobileList`:

```go
// @Description  Always visibility-filtered to published + audience match.
// @Description  Each item includes a plain-text description preview limited to
// @Description  100 Unicode characters; fetch full content via MobileGet.
```

- [ ] **Step 2: Regenerate Swagger artifacts**

Run: `swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal`

Expected: command exits 0 and updates the three generated Swagger files.

- [ ] **Step 3: Format and inspect the change**

Run: `gofmt -w internal/dto/announcement.go internal/handlers/announcement_handler.go internal/services/announcement_notifier.go internal/services/announcement_preview.go internal/services/announcement_preview_test.go internal/services/announcement_service.go internal/services/announcement_service_test.go`

Run: `git diff --check`

Expected: both commands exit 0 with no formatting errors.

- [ ] **Step 4: Run full verification**

Run: `go test ./...`

Expected: PASS for every package.

Run: `go vet ./...`

Expected: exit 0 with no diagnostics.

- [ ] **Step 5: Commit documentation and generated artifacts**

```bash
git add internal/handlers/announcement_handler.go docs/swagger/docs.go docs/swagger/swagger.json docs/swagger/swagger.yaml
git commit -m "docs: document mobile announcement description preview"
```
