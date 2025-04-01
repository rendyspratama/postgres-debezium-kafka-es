package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitPrometheus(port int, path string) error {
	http.Handle(path, promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			panic(fmt.Sprintf("Failed to start metrics server: %v", err))
		}
	}()
	return nil
}
