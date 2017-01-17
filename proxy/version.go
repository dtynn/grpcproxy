package proxy

import (
	"fmt"
)

var version = Version{
	Majar: 0,
	Minor: 1,
	Patch: 6,
}

type Version struct {
	Majar, Minor, Patch int
}

func (this *Version) String() string {
	return fmt.Sprintf("ver %d.%d.%d", this.Majar, this.Minor, this.Patch)
}
