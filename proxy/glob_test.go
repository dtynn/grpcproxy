package proxy

import (
	"testing"

	"github.com/gobwas/glob"
)

func TestGlob(t *testing.T) {
	testcases := []struct {
		Str      string
		Pattern  string
		Expected bool
	}{
		{
			"",
			"*",
			true,
		},
	}

	for i, c := range testcases {
		pattern := glob.MustCompile(c.Pattern)
		got := pattern.Match(c.Str)
		if got != c.Expected {
			t.Errorf("#%d matching %q for pattern %q: expected %v, got %v", i, c.Str, c.Pattern, c.Expected, got)
		}
	}
}
