package netutil

import (
	"net/http"
)

type Balancer interface {
	Pick(req *http.Request) http.Handler
}
