package exporter

import (
	"NeoPluginMaster/api"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var PlayerAmount prometheus.Gauge
var ServerAmount prometheus.Gauge

func Run() {
	http.Handle("/metrics", promhttp.Handler())
	PlayerAmount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "plugin_player_amount",
		Help: "show the amount of players online",
	})
	PlayerAmount.Set(api.PlayerCount)

	ServerAmount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "plugin_server_amount",
		Help: "show the amount of server online",
	})
	ServerAmount.Set(api.ServerCount)

	/*
		playerCounter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "plugin_online_player",
			Help: "show the amount of players online",
		}, []string{"online_player"})

	*/

	fmt.Println("PrometheusExporter started at 0.0.0.0:8069")
	log.Fatal(http.ListenAndServe(":8069", nil))
}
