package exporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func Run() {
	http.Handle("/metrics", promhttp.Handler())

	/*SessionCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "player_location_cords",
		Help: "show player join locations",
	}, []string{"country", "latitude", "longitude"})

	*/

	fmt.Println("PrometheusExporter started at 0.0.0.0:8069")
	log.Fatal(http.ListenAndServe(":8069", nil))
}
