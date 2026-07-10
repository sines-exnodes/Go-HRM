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
