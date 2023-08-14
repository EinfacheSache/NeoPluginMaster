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
var PluginVersion *prometheus.GaugeVec
var ServerVersion *prometheus.GaugeVec

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

	PluginVersion = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugin_plugin_versions",
		Help: "show the version of the servers",
	}, []string{"plugin_version"})

	ServerVersion = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugin_server_versions",
		Help: "show the version of the servers",
	}, []string{"server_version"})

	ServerStats = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "server_stats",
		Help: "show the version of the servers",
	}, []string{"backendID", "server_version", "server_name", "java_version", "os_name", "os_arch", "os_version", "plugin_version", "version_status", "update_setting", "neo_protect_plan", "server_plugins", "player_amount", "managed_servers", "core_count", "online_mode", "proxy_protocol"})

	fmt.Println("PrometheusExporter started at 0.0.0.0:8069")
	log.Fatal(http.ListenAndServe(":8069", nil))
}
