package utils

import "testing"

func TestEscapeILIKE(t *testing.T) {
	cases := map[string]string{
		"":          "",
		"john":      "john",
		"100%":      "100\\%",
		"foo_bar":   "foo\\_bar",
		"a\\b":      "a\\\\b",
		"50%_off\\": "50\\%\\_off\\\\",
	}
	for in, want := range cases {
		if got := EscapeILIKE(in); got != want {
			t.Errorf("EscapeILIKE(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildILIKEPattern(t *testing.T) {
	if got := BuildILIKEPattern("a%b"); got != "%a\\%b%" {
		t.Errorf("got %q", got)
	}
}
