package main

import (
	"log"
	"net/http"
	"time"

	"github.com/ic-n/legacy-endpoint/examples/static-media/media"
)

func main() {
	s := &http.Server{
		Addr:    ":8080",
		Handler: media.Handler(media.Config{
			// [...]
		}),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
