package proxy

import (
	"strings"
)

func str2NonEmptySlice(s, sep string) []string {
	pieces := strings.Split(s, sep)

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
