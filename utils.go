package main

import (
	"html/template"
	"net/http"
	"sync"
)

var (
	lck  sync.Mutex
	tpls = template.Must(template.ParseGlob("templates/*.html"))
)

func withLock(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		lck.Lock()
		defer lck.Unlock()
		h.ServeHTTP(w, req)
	})
}
