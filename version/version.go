package version

import (
	"fmt"
)

func Version() *version {
	return v
}

var v = &version{
	major: 0,
	minor: 2,
	patch: 2,
}

type version struct {
	major int
	minor int
	patch int
}

func (this *version) String() string {
	return fmt.Sprintf("%d.%d.%d", this.major, this.minor, this.patch)
}
