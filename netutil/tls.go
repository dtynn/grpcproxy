package netutil

import (
	"golang.org/x/net/http2"
)

var NextProtos = []string{http2.NextProtoTLS}
