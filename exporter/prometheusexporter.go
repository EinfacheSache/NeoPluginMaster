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
var ServerVersion *prometheus.CounterVec
var ServerStats *prometheus.CounterVec

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
		Name: "plugin_plugin_versions",
		Help: "show the version of the servers",
	}, []string{"plugin_version"})

	ServerVersion = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "plugin_server_versions",
		Help: "show the version of the servers",
	}, []string{"server_version"})

	ServerStats = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "server_stats",
		Help: "show the version of the servers",
	}, []string{"backendID", "online_mode", "player_amount", "managed_servers", "core_count", "server_version", "server_name", "java_version", "os_name", "os_arch", "os_version", "plugin_version"})

	fmt.Println("PrometheusExporter started at 0.0.0.0:8069")
	log.Fatal(http.ListenAndServe(":8069", nil))
}
