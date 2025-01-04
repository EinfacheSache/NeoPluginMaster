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
	identifier     string
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
	VersionError   string  `json:"versionError"`
	UpdateSetting  string  `json:"updateSetting"`
	NeoProtectPlan string  `json:"neoProtectPlan"`
	ServerPlugins  string  `json:"serverPlugins"`
	PlayerAmount   float64 `json:"playerAmount"`
	ManagedServer  float64 `json:"managedServers"`
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

var limiter = rate.NewLimiter(rate.Every(1*time.Second/15), 15)

func Run() {
	http.HandleFunc("/api/stats/plugin", pluginMetricsFailedHandler)
	fmt.Println("API listener started at 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func pluginMetricsFailedHandler(w http.ResponseWriter, r *http.Request) {

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
		fmt.Println("\033[31mrate limit exceeded\033[0m")
		return
	}

	var statsRequest stats

	statsRequest.identifier = r.Header.Get("identifier")
	if statsRequest.identifier == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Println("\033[31mrequest failed: identifier not provided\033[0m")
		return
	}

	if r.Method == http.MethodPost {

		err2 := json.NewDecoder(r.Body).Decode(&statsRequest)
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusBadRequest)
			fmt.Println("\033[31mrequest failed: formatted error\033[0m")
			return
		}

		w.WriteHeader(http.StatusOK)

		pluginMetrics(statsRequest)

		return
	}

	if r.Method == http.MethodDelete {

		BackendStatsMutex.Lock()
		latestStats, ok := BackendStats[statsRequest.identifier]
		if !ok {
			BackendStatsMutex.Unlock()
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusFound)

		fmt.Println("\033[33mThe server went down (", statsRequest.identifier, ")\033[0m")
		delete(BackendStats, statsRequest.identifier)

		BackendStatsMutex.Unlock()

		setOffline(latestStats, statsRequest.identifier)

		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	fmt.Println("\033[31mMethod not allowed\033[0m")
	return
}

func pluginMetrics(statsRequest stats) {
	fmt.Println("ID (", statsRequest.identifier, ")", " PlayerCount(", statsRequest.PlayerAmount, ")", "ServerType(", statsRequest.ServerType, ")")

	statsRequest.latestPing = time.Now().UnixMilli()

	BackendServerStatsMutex.RLock()
	latestServerStats, ok := BackendServerStats[statsRequest.identifier]
	BackendServerStatsMutex.RUnlock()

	if ok {
		exporter.ServerStats.DeletePartialMatch(latestServerStats)
	}

	BackendStatsMutex.Lock()
	latestStats, ok := BackendStats[statsRequest.identifier]
	BackendStats[statsRequest.identifier] = statsRequest
	BackendStatsMutex.Unlock()

	if ok {
		AmountStatsMutex.Lock()
		AmountStats[latestStats.ServerType+"PlayerCount"] -= latestStats.PlayerAmount
		AmountStats[latestStats.ServerType+"ServerCount"] -= 1
		AmountStats[latestStats.ServerType+"ManageServer"] -= latestStats.ManagedServer
		AmountStatsMutex.Unlock()

		delLabel(exporter.PluginVersion, latestStats.ServerType, "plugin_version", latestStats.PluginVersion)
		delLabel(exporter.ServerVersion, latestStats.ServerType, "server_version", latestStats.ServerVersion)
		delLabel(exporter.VersionStatus, latestStats.ServerType, "version_status", latestStats.VersionStatus)
		delLabel(exporter.VersionError, latestStats.ServerType, "version_error", latestStats.VersionError)
		delLabel(exporter.UpdateSetting, latestStats.ServerType, "update_setting", latestStats.UpdateSetting)
		delLabel(exporter.NeoProtectPlan, latestStats.ServerType, "neoprotect_plan", latestStats.NeoProtectPlan)

		delLabel(exporter.ServerName, latestStats.ServerType, "server_name", latestStats.ServerName)
		delLabel(exporter.JavaVersion, latestStats.ServerType, "java_version", latestStats.JavaVersion)
		delLabel(exporter.OsName, latestStats.ServerType, "os_name", latestStats.OsName)
		delLabel(exporter.OsArch, latestStats.ServerType, "os_arch", latestStats.OsArch)
		delLabel(exporter.OsVersion, latestStats.ServerType, "os_version", latestStats.OsVersion)
		delLabel(exporter.CoreCount, latestStats.ServerType, "core_count", strconv.FormatFloat(latestStats.CoreCount, 'f', 0, 64))
		delLabel(exporter.OnlineMode, latestStats.ServerType, "online_mode", strconv.FormatBool(latestStats.OnlineMode))
		delLabel(exporter.ProxyProtocol, latestStats.ServerType, "proxy_protocol", strconv.FormatBool(latestStats.ProxyProtocol))
	}

	AmountStatsMutex.Lock()
	AmountStats[statsRequest.ServerType+"PlayerCount"] += statsRequest.PlayerAmount
	AmountStats[statsRequest.ServerType+"ServerCount"] += 1
	AmountStats[statsRequest.ServerType+"ManageServer"] += statsRequest.ManagedServer
	AmountStatsMutex.Unlock()

	AmountStatsMutex.RLock()
	exporter.PlayerAmount.With(prometheus.Labels{"server_type": statsRequest.ServerType}).Set(AmountStats[statsRequest.ServerType+"PlayerCount"])
	exporter.ServerAmount.With(prometheus.Labels{"server_type": statsRequest.ServerType}).Set(AmountStats[statsRequest.ServerType+"ServerCount"])
	exporter.ManageServer.With(prometheus.Labels{"server_type": statsRequest.ServerType}).Set(AmountStats[statsRequest.ServerType+"ManageServer"])
	exporter.PlayerAmount.With(prometheus.Labels{"server_type": latestStats.ServerType}).Set(AmountStats[latestStats.ServerType+"PlayerCount"])
	exporter.ServerAmount.With(prometheus.Labels{"server_type": latestStats.ServerType}).Set(AmountStats[latestStats.ServerType+"ServerCount"])
	exporter.ManageServer.With(prometheus.Labels{"server_type": latestStats.ServerType}).Set(AmountStats[latestStats.ServerType+"ManageServer"])
	AmountStatsMutex.RUnlock()

	addLabel(exporter.ServerVersion, statsRequest.ServerType, "server_version", statsRequest.ServerVersion)
	addLabel(exporter.PluginVersion, statsRequest.ServerType, "plugin_version", statsRequest.PluginVersion)
	addLabel(exporter.VersionStatus, statsRequest.ServerType, "version_status", statsRequest.VersionStatus)
	addLabel(exporter.VersionError, statsRequest.ServerType, "version_error", statsRequest.VersionError)
	addLabel(exporter.UpdateSetting, statsRequest.ServerType, "update_setting", statsRequest.UpdateSetting)
	addLabel(exporter.NeoProtectPlan, statsRequest.ServerType, "neoprotect_plan", statsRequest.NeoProtectPlan)

	addLabel(exporter.ServerName, statsRequest.ServerType, "server_name", statsRequest.ServerName)
	addLabel(exporter.JavaVersion, statsRequest.ServerType, "java_version", statsRequest.JavaVersion)
	addLabel(exporter.OsName, statsRequest.ServerType, "os_name", statsRequest.OsName)
	addLabel(exporter.OsArch, statsRequest.ServerType, "os_arch", statsRequest.OsArch)
	addLabel(exporter.OsVersion, statsRequest.ServerType, "os_version", statsRequest.OsVersion)
	addLabel(exporter.CoreCount, statsRequest.ServerType, "core_count", strconv.FormatFloat(statsRequest.CoreCount, 'f', 0, 64))
	addLabel(exporter.OnlineMode, statsRequest.ServerType, "online_mode", strconv.FormatBool(statsRequest.OnlineMode))
	addLabel(exporter.ProxyProtocol, statsRequest.ServerType, "proxy_protocol", strconv.FormatBool(statsRequest.ProxyProtocol))

	addServerStatsLabel(statsRequest)

	go startTimeout(statsRequest.identifier)
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
		"identifier": statsRequest.identifier,
		"serverID":   statsRequest.serverID,
		"backendID":  statsRequest.backendID,

		"server_type":      statsRequest.ServerType,
		"server_version":   statsRequest.ServerVersion,
		"server_name":      statsRequest.ServerName,
		"java_version":     statsRequest.JavaVersion,
		"os_name":          statsRequest.OsName,
		"os_arch":          statsRequest.OsArch,
		"os_version":       statsRequest.OsVersion,
		"plugin_version":   statsRequest.PluginVersion,
		"version_status":   statsRequest.VersionStatus,
		"version_error":    statsRequest.VersionError,
		"update_setting":   statsRequest.UpdateSetting,
		"neo_protect_plan": statsRequest.NeoProtectPlan,
		"server_plugins":   statsRequest.ServerPlugins,
		"player_amount":    fmt.Sprintf("%f", statsRequest.PlayerAmount),
		"managed_servers":  fmt.Sprintf("%f", statsRequest.ManagedServer),
		"core_count":       fmt.Sprintf("%f", statsRequest.CoreCount),
		"online_mode":      strconv.FormatBool(statsRequest.OnlineMode),
		"proxy_protocol":   strconv.FormatBool(statsRequest.ProxyProtocol),
	}

	BackendServerStatsMutex.Lock()
	BackendServerStats[statsRequest.identifier] = label
	BackendServerStatsMutex.Unlock()

	exporter.ServerStats.With(label).Inc()
}

func startTimeout(identifier string) {

	time.Sleep(time.Second * 10)

	BackendStatsMutex.Lock()
	latestStats, ok := BackendStats[identifier]
	if time.Now().UnixMilli()-latestStats.latestPing < 1000*10 || !ok {
		BackendStatsMutex.Unlock()
		return
	} else {
		fmt.Println("\033[33mThe server has timed out (", identifier, ") PlayerCount(", latestStats.PlayerAmount, ")\033[0m")
		delete(BackendStats, identifier)
	}
	BackendStatsMutex.Unlock()

	setOffline(latestStats, identifier)
}

func setOffline(latestStats stats, identifier string) {

	AmountStatsMutex.Lock()
	AmountStats[latestStats.ServerType+"PlayerCount"] -= latestStats.PlayerAmount
	AmountStats[latestStats.ServerType+"ServerCount"] -= 1
	AmountStats[latestStats.ServerType+"ManageServer"] -= latestStats.ManagedServer
	AmountStatsMutex.Unlock()

	AmountStatsMutex.RLock()
	exporter.PlayerAmount.With(prometheus.Labels{"server_type": latestStats.ServerType}).Set(AmountStats[latestStats.ServerType+"PlayerCount"])
	exporter.ServerAmount.With(prometheus.Labels{"server_type": latestStats.ServerType}).Set(AmountStats[latestStats.ServerType+"ServerCount"])
	exporter.ManageServer.With(prometheus.Labels{"server_type": latestStats.ServerType}).Set(AmountStats[latestStats.ServerType+"ManageServer"])
	AmountStatsMutex.RUnlock()

	delLabel(exporter.PluginVersion, latestStats.ServerType, "plugin_version", latestStats.PluginVersion)
	delLabel(exporter.ServerVersion, latestStats.ServerType, "server_version", latestStats.ServerVersion)
	delLabel(exporter.VersionStatus, latestStats.ServerType, "version_status", latestStats.VersionStatus)
	delLabel(exporter.VersionError, latestStats.ServerType, "version_error", latestStats.VersionError)
	delLabel(exporter.UpdateSetting, latestStats.ServerType, "update_setting", latestStats.UpdateSetting)
	delLabel(exporter.NeoProtectPlan, latestStats.ServerType, "neoprotect_plan", latestStats.NeoProtectPlan)

	delLabel(exporter.ServerName, latestStats.ServerType, "server_name", latestStats.ServerName)
	delLabel(exporter.JavaVersion, latestStats.ServerType, "java_version", latestStats.JavaVersion)
	delLabel(exporter.OsName, latestStats.ServerType, "os_name", latestStats.OsName)
	delLabel(exporter.OsArch, latestStats.ServerType, "os_arch", latestStats.OsArch)
	delLabel(exporter.OsVersion, latestStats.ServerType, "os_version", latestStats.OsVersion)
	delLabel(exporter.CoreCount, latestStats.ServerType, "core_count", strconv.FormatFloat(latestStats.CoreCount, 'f', 0, 64))
	delLabel(exporter.OnlineMode, latestStats.ServerType, "online_mode", strconv.FormatBool(latestStats.OnlineMode))
	delLabel(exporter.ProxyProtocol, latestStats.ServerType, "proxy_protocol", strconv.FormatBool(latestStats.ProxyProtocol))

	BackendServerStatsMutex.Lock()
	latestServerStats := BackendServerStats[identifier]
	delete(BackendServerStats, identifier)
	BackendServerStatsMutex.Unlock()

	exporter.ServerStats.Delete(latestServerStats)
}
