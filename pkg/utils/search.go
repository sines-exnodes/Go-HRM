package utils

import "strings"

// EscapeILIKE escapes %, _, and backslash so user input cannot inject ILIKE wildcards.
func EscapeILIKE(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	)
	return r.Replace(s)
}

// BuildILIKEPattern returns "%<escaped>%" suitable for `ILIKE ?`.
func BuildILIKEPattern(s string) string {
	return "%" + EscapeILIKE(s) + "%"
}
