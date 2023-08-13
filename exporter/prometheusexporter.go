package exporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var PlayerAmount prometheus.Gauge
var ServerAmount prometheus.Gauge
var PluginVersion *prometheus.CounterVec

func Run() {
	http.Handle("/metrics", promhttp.Handler())
	PlayerAmount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "plugin_player_amount",
		Help: "show the amount of players online",
	})
	PlayerAmount.Set(0)

	ServerAmount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "plugin_server_amount",
		Help: "show the amount of server online",
	})
	ServerAmount.Set(0)

	PluginVersion = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "plugin_server_versions",
		Help: "show the version of the servers",
	}, []string{"server_version"})

	fmt.Println("PrometheusExporter started at 0.0.0.0:8069")
	log.Fatal(http.ListenAndServe(":8069", nil))
}
