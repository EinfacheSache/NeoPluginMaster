package exporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var PlayerAmount *prometheus.GaugeVec
var ServerAmount *prometheus.GaugeVec
var ManageServer *prometheus.GaugeVec
var PluginVersion *prometheus.GaugeVec
var ServerVersion *prometheus.GaugeVec
var VersionStatus *prometheus.GaugeVec
var UpdateSetting *prometheus.GaugeVec
var NeoProtectPlan *prometheus.GaugeVec

var ServerStats *prometheus.CounterVec

func Run() {
	http.Handle("/metrics", promhttp.Handler())

	registerServerSpecificStats()

	fmt.Println("PrometheusExporter started at 0.0.0.0:8069")
	log.Fatal(http.ListenAndServe(":8069", nil))
}

func registerServerSpecificStats() {

	PlayerAmount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugin_player_amount",
		Help: "show the amount of players online",
	}, []string{"server_type"})
	PlayerAmount.With(prometheus.Labels{"server_type": "bungeecord"}).Set(0)
	PlayerAmount.With(prometheus.Labels{"server_type": "velocity"}).Set(0)
	PlayerAmount.With(prometheus.Labels{"server_type": "spigot"}).Set(0)

	ServerAmount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugin_server_amount",
		Help: "show the amount of server online",
	}, []string{"server_type"})
	ServerAmount.With(prometheus.Labels{"server_type": "bungeecord"}).Set(0)
	ServerAmount.With(prometheus.Labels{"server_type": "velocity"}).Set(0)
	ServerAmount.With(prometheus.Labels{"server_type": "spigot"}).Set(0)

	ManageServer = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugin_manage_server_amount",
		Help: "show the amount of manage server online",
	}, []string{"server_type"})
	ManageServer.With(prometheus.Labels{"server_type": "bungeecord"}).Set(0)
	ManageServer.With(prometheus.Labels{"server_type": "velocity"}).Set(0)

	PluginVersion = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugin_plugin_versions",
		Help: "show the version of the servers",
	}, []string{"server_type", "plugin_version"})

	ServerVersion = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugin_server_versions",
		Help: "show the version of the servers",
	}, []string{"server_type", "server_version"})

	VersionStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugin_version_status",
		Help: "show the version status of the servers",
	}, []string{"server_type", "version_status"})

	UpdateSetting = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugin_update_setting",
		Help: "show the update setting of the servers",
	}, []string{"server_type", "update_setting"})

	NeoProtectPlan = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "plugin_neoprotect_plan",
		Help: "show the NeoProtect plan of the servers",
	}, []string{"server_type", "neoprotect_plan"})

	ServerStats = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "server_stats",
		Help: "show the version of the servers",
	}, []string{"serverID", "backendID", "server_type", "server_version", "server_name", "java_version", "os_name", "os_arch", "os_version", "plugin_version", "version_status", "update_setting", "neo_protect_plan", "server_plugins", "player_amount", "managed_servers", "core_count", "online_mode", "proxy_protocol"})

}
