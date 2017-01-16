package proxy

import (
	"log"
	"net/http"
)

func newProxyHandler(apps []*App) http.Handler {
	f := func(rw http.ResponseWriter, req *http.Request) {
		for _, app := range apps {
			proxy, ok := app.Match(req)
			if ok {
				proxy.handler.ServeHTTP(rw, req)
				return
			}
		}

		log.Printf("[NOT FOUND][%s] %s%s", req.Method, req.Host, req.RequestURI)
	}

	return http.HandlerFunc(f)
}
