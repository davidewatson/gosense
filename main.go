package main

import (
	"log"
	"net/http"

	"experimental/dwat/gosense/pkg/cache"
	"experimental/dwat/gosense/pkg/lmsensors"
	"experimental/dwat/gosense/pkg/lmsensors/classic"
)

const (
	serverAddr      = ":8080" // Address and port for the http server to listen on
	defaultInterval = 60      // Number of seconds between attempts to update the cache
)

func main() {
	startAndRegister("csensors", classic.Update, "/api/sys/sensors")
	startAndRegister("sensors", lmsensors.Update, "/api/sys/sensors2")

	log.Fatal(http.ListenAndServe(serverAddr, nil))
}

// startAndRegister uses a goroutine to start each cache concurrently, only
// registering a handler on success. This helps ensure monitoring starts as
// quickly as possible in an emergency.
func startAndRegister(name string, update cache.Update, pattern string) {
	go func() {
		c := cache.NewCache(name, update, defaultInterval).Start()
		if c != nil {
			http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
				w.Write(c.Get())
			})
		}
	}()
}
