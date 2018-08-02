package main

import (
	"log"
	"net/http"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf(
			"proto=%s method=%s remote-addr=%s\n",
			r.Proto,
			r.Method,
			r.RemoteAddr,
		)

		next.ServeHTTP(w, r)
	})
}
