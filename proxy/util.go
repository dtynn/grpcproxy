package proxy

import (
	"strings"
)

func str2NonEmptySlice(s, sep string) []string {
	return nonEmptySlice(strings.Split(s, sep))
}

func nonEmptySlice(pieces []string) []string {
	i := 0
	for i < len(pieces) {
		part := strings.TrimSpace(pieces[i])
		if part == "" {
			pieces = append(pieces[:i], pieces[i+1:]...)
			continue
		}

		pieces[i] = part
		i += 1
	}
	return pieces
}
