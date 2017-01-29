package netutil

import (
	"net/http"
)

type Balancer interface {
	Pick() http.Handler
}
