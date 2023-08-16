package api

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	"log"
	"neo-plugin-master/exporter"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type stats struct {
	serverID       string
	backendID      string
	latestPing     int64
	ServerType     string  `json:"serverType"`
	ServerVersion  string  `json:"serverVersion"`
	ServerName     string  `json:"serverName"`
	JavaVersion    string  `json:"javaVersion"`
	OsName         string  `json:"osName"`
	OsArch         string  `json:"osArch"`
	OsVersion      string  `json:"osVersion"`
	PluginVersion  string  `json:"pluginVersion"`
	VersionStatus  string  `json:"versionStatus"`
	UpdateSetting  string  `json:"updateSetting"`
	NeoProtectPlan string  `json:"neoProtectPlan"`
	ServerPlugins  string  `json:"serverPlugins"`
	PlayerAmount   float64 `json:"playerAmount"`
	ManagedServers float64 `json:"managedServers"`
	CoreCount      float64 `json:"coreCount"`
	OnlineMode     bool    `json:"onlineMode"`
	ProxyProtocol  bool    `json:"proxyProtocol"`
}

type ResponseMessage struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

var AmountStats = map[string]float64{}
var AmountStatsMutex = new(sync.RWMutex)

var BackendStats = map[string]stats{}
var BackendStatsMutex = new(sync.RWMutex)

var BackendServerStats = map[string]prometheus.Labels{}
var BackendServerStatsMutex = new(sync.RWMutex)

var limiter = rate.NewLimiter(rate.Every(1*time.Second/30), 30)

func Run() {
	http.HandleFunc("/api/stats/plugin", pluginMetricsFailedHandler)
	fmt.Println("API listener started at 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func pluginMetricsFailedHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		fmt.Println("Method not allowed")
		return
	}

	if !limiter.Allow() {
		message := ResponseMessage{
			Status: "Rate limit exceeded",
			Body:   "You are being rate limited. Please try again later.",
		}

		w.WriteHeader(http.StatusTooManyRequests)
		err := json.NewEncoder(w).Encode(&message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		fmt.Println("rate limit exceeded")
		return
	}

	var statsRequest stats
	statsRequest.backendID = r.Header.Get("backendID")
	if statsRequest.backendID == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Println("request failed: backendID not provided")
		return
	}

	statsRequest.serverID = r.Header.Get("gameshieldID")
	if statsRequest.serverID == "" {
		//coming soon
	}

	err2 := json.NewDecoder(r.Body).Decode(&statsRequest)
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusBadRequest)
		fmt.Println("request failed: formatted error")
		return
	}
	w.WriteHeader(http.StatusOK)

	pluginMetrics(statsRequest)
}

func pluginMetrics(statsRequest stats) {
	fmt.Println("ID (", statsRequest.backendID, ")", " PlayerCount(", statsRequest.PlayerAmount, ")")

	statsRequest.latestPing = time.Now().UnixMilli()

	BackendServerStatsMutex.RLock()
	latestServerStats, ok := BackendServerStats[statsRequest.backendID]
	BackendServerStatsMutex.RUnlock()
	if ok {
		exporter.ServerStats.DeletePartialMatch(latestServerStats)
	}

	BackendStatsMutex.RLock()
	latestStats, ok2 := BackendStats[statsRequest.backendID]
	BackendStatsMutex.RUnlock()
	if ok2 {
		AmountStatsMutex.Lock()
		AmountStats[latestStats.ServerType+"PlayerCount"] -= latestStats.PlayerAmount
		AmountStats[latestStats.ServerType+"ServerCount"] -= 1
		AmountStatsMutex.Unlock()

		delLabel(exporter.PluginVersion, statsRequest.ServerType, "plugin_version", latestStats.PluginVersion)
		delLabel(exporter.ServerVersion, statsRequest.ServerType, "server_version", latestStats.ServerVersion)
		delLabel(exporter.VersionStatus, statsRequest.ServerType, "version_status", latestStats.VersionStatus)
		delLabel(exporter.UpdateSetting, statsRequest.ServerType, "update_setting", latestStats.UpdateSetting)
		delLabel(exporter.NeoProtectPlan, statsRequest.ServerType, "neoprotect_plan", latestStats.NeoProtectPlan)
	}

	AmountStatsMutex.Lock()
	AmountStats[statsRequest.ServerType+"PlayerCount"] += statsRequest.PlayerAmount
	AmountStats[statsRequest.ServerType+"ServerCount"] += 1
	AmountStatsMutex.Unlock()

	AmountStatsMutex.RLock()
	exporter.PlayerAmount.With(prometheus.Labels{"server_type": statsRequest.ServerType}).Set(AmountStats[statsRequest.ServerType+"PlayerCount"])
	exporter.ServerAmount.With(prometheus.Labels{"server_type": statsRequest.ServerType}).Set(AmountStats[statsRequest.ServerType+"ServerCount"])
	AmountStatsMutex.RUnlock()

	addLabel(exporter.ServerVersion, statsRequest.ServerType, "server_version", statsRequest.ServerVersion)
	addLabel(exporter.PluginVersion, statsRequest.ServerType, "plugin_version", statsRequest.PluginVersion)
	addLabel(exporter.VersionStatus, statsRequest.ServerType, "version_status", statsRequest.VersionStatus)
	addLabel(exporter.UpdateSetting, statsRequest.ServerType, "update_setting", statsRequest.UpdateSetting)
	addLabel(exporter.NeoProtectPlan, statsRequest.ServerType, "neoprotect_plan", statsRequest.NeoProtectPlan)

	//AddVersionSpecificStats(statsRequest.ServerType, statsRequest)

	addServerStatsLabel(statsRequest)

	BackendStatsMutex.Lock()
	BackendStats[statsRequest.backendID] = statsRequest
	BackendStatsMutex.Unlock()

	go startTimeout(statsRequest.backendID)
}

func addLabel(metrics *prometheus.GaugeVec, serverTyp string, key string, value string) {
	if value == "" {
		return
	}

	label := prometheus.Labels{
		"server_type": serverTyp,
		key:           value,
	}

	metrics.With(label).Add(1)
}

func delLabel(metrics *prometheus.GaugeVec, serverTyp string, key string, value string) {
	if value == "" {
		return
	}

	label := prometheus.Labels{
		"server_type": serverTyp,
		key:           value,
	}
	metrics.With(label).Sub(1)
}

func addServerStatsLabel(statsRequest stats) {
	label := prometheus.Labels{
		"serverID":  statsRequest.serverID,
		"backendID": statsRequest.backendID,

		"server_type":      statsRequest.ServerType,
		"server_version":   statsRequest.ServerVersion,
		"server_name":      statsRequest.ServerName,
		"java_version":     statsRequest.JavaVersion,
		"os_name":          statsRequest.OsName,
		"os_arch":          statsRequest.OsArch,
		"os_version":       statsRequest.OsVersion,
		"plugin_version":   statsRequest.PluginVersion,
		"version_status":   statsRequest.VersionStatus,
		"update_setting":   statsRequest.UpdateSetting,
		"neo_protect_plan": statsRequest.NeoProtectPlan,
		"server_plugins":   statsRequest.ServerPlugins,
		"player_amount":    fmt.Sprintf("%f", statsRequest.PlayerAmount),
		"managed_servers":  fmt.Sprintf("%f", statsRequest.ManagedServers),
		"core_count":       fmt.Sprintf("%f", statsRequest.CoreCount),
		"online_mode":      strconv.FormatBool(statsRequest.OnlineMode),
		"proxy_protocol":   strconv.FormatBool(statsRequest.ProxyProtocol),
	}

	BackendServerStatsMutex.Lock()
	BackendServerStats[statsRequest.backendID] = label
	BackendServerStatsMutex.Unlock()

	exporter.ServerStats.With(label).Inc()
}

func startTimeout(backendID string) {

	time.Sleep(time.Second * 30)

	BackendStatsMutex.RLock()
	latestStats, ok := BackendStats[backendID]
	BackendStatsMutex.RUnlock()
	if !ok {
		fmt.Printf("cant found key in map %s\n", backendID)
		return
	}

	if time.Now().UnixMilli()-latestStats.latestPing < 1000*30 {
		// Server did not timeout and send ping in latest 40 sec -> dont delete
		return
	}

	AmountStatsMutex.Lock()
	AmountStats[latestStats.ServerType+"PlayerCount"] -= latestStats.PlayerAmount
	AmountStats[latestStats.ServerType+"ServerCount"] -= 1
	AmountStatsMutex.Unlock()

	AmountStatsMutex.RLock()
	exporter.PlayerAmount.With(prometheus.Labels{"server_type": latestStats.ServerType}).Set(AmountStats[latestStats.ServerType+"PlayerCount"])
	exporter.ServerAmount.With(prometheus.Labels{"server_type": latestStats.ServerType}).Set(AmountStats[latestStats.ServerType+"ServerCount"])
	AmountStatsMutex.RUnlock()

	delLabel(exporter.PluginVersion, latestStats.ServerType, "plugin_version", latestStats.PluginVersion)
	delLabel(exporter.ServerVersion, latestStats.ServerType, "server_version", latestStats.ServerVersion)
	delLabel(exporter.VersionStatus, latestStats.ServerType, "version_status", latestStats.VersionStatus)
	delLabel(exporter.UpdateSetting, latestStats.ServerType, "update_setting", latestStats.UpdateSetting)
	delLabel(exporter.NeoProtectPlan, latestStats.ServerType, "neoprotect_plan", latestStats.NeoProtectPlan)

	BackendStatsMutex.Lock()
	delete(BackendStats, backendID)
	BackendStatsMutex.Unlock()

	BackendServerStatsMutex.RLock()
	latestServerStats, ok2 := BackendServerStats[latestStats.backendID]
	BackendServerStatsMutex.RUnlock()
	if ok2 {
		exporter.ServerStats.Delete(latestServerStats)
		BackendStatsMutex.Lock()
		delete(BackendStats, backendID)
		BackendStatsMutex.Unlock()
	}
}
