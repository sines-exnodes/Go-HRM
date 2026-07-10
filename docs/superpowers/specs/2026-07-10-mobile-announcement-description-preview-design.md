# Mobile Announcement Description Preview

## Problem

`GET /api/v1/mobile/announcements` returns `MobileAnnouncementBrief` items,
which currently omit the announcement description. The mobile home screen needs
a short description preview, but announcement descriptions can contain rich-text
HTML and can be too long for a compact card.

## API Contract

Add a required `description` string to `MobileAnnouncementBrief`. The field is
returned by both endpoints that use this projection:

- `GET /api/v1/mobile/announcements`
- `GET /api/v1/mobile/announcements/list`

The detail endpoint remains unchanged and continues to return the full stored
description.

Example preview:

```json
{
  "id": "2f82d850-e221-4bd0-a6f1-efdb06339ca4",
  "title": "Office maintenance",
  "description": "The office will close at 5 PM on Friday & reopen Monday.",
  "has_viewed": false
}
```

## Preview Rules

The preview is derived from the stored description when the response is built:

1. Replace HTML tags with spaces.
2. Decode HTML entities such as `&amp;` and `&nbsp;`.
3. Collapse consecutive whitespace and trim leading/trailing whitespace.
4. Return at most 100 Unicode characters.
5. When truncation is required, return the first 99 characters followed by the
   single-character ellipsis `…`, keeping the total at 100 characters.

An empty or HTML-only description produces an empty string. Stored announcement
content is not modified.

## Implementation

Generalize the existing push-notification plain-text conversion into a shared,
private announcement preview helper that accepts a character limit. Keep the
push notification behavior at 128 characters and use a 100-character limit for
`MobileAnnouncementBrief`.

This keeps HTML handling consistent across mobile previews and push
notifications while avoiding handler-specific transformation logic.

## Verification

Add regression coverage proving that the mobile projection:

- includes `description`;
- strips tags, decodes entities, and normalizes whitespace;
- truncates by Unicode characters rather than bytes;
- preserves descriptions that are exactly 100 characters;
- appends `…` only when the normalized text exceeds 100 characters.

Run focused announcement service tests, then `go test ./...` and `go vet ./...`.

## Non-Goals

- No database migration or stored-content rewrite.
- No change to announcement visibility, ordering, or pagination.
- No preference for the optional `summary` field over `description`.
- No change to the full mobile announcement detail response.
