package main

import (
	"net/http"
	"log"
)

func loggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("получен запрос: %s %s", r.Method, r.URL.Path)

		next(w, r)
	}
}