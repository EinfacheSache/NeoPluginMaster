package api

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	"log"
	"neo-plugin-master/exporter"
	"net/http"
	"sync"
	"time"
)

type stats struct {
	PlayerAmount   float64 `json:"playerAmount"`
	ManagedServers float64 `json:"managedServers"`
	OnlineMode     bool    `json:"onlineMode"`
	ServerVersion  string  `json:"serverVersion"`
	ServerName     string  `json:"serverName"`
	JavaVersion    string  `json:"javaVersion"`
	OsName         string  `json:"osName"`
	OsArch         string  `json:"osArch"`
	OsVersion      string  `json:"osVersion"`
	CoreCount      float64 `json:"coreCount"`
	PluginVersion  string  `json:"pluginVersion"`
	latestPing     int64
	backendID      string
}

type ResponseMessage struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

var BackendStats = map[string]stats{}
var BackendStatsMutex = new(sync.RWMutex)

var ServerCount = float64(0)
var PlayerCount = float64(0)

var limiter = rate.NewLimiter(rate.Every(1*time.Second/30), 1)

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
			Body:   "You are being rate limitted. Please try again later.",
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

	BackendStatsMutex.RLock()
	latestStats, ok := BackendStats[statsRequest.backendID]
	BackendStatsMutex.RUnlock()
	if ok {
		PlayerCount -= latestStats.PlayerAmount
		ServerCount -= 1
		labels := prometheus.Labels{
			"server_version": latestStats.ServerVersion,
		}
		exporter.PluginVersion.DeletePartialMatch(labels)
	}

	PlayerCount += statsRequest.PlayerAmount
	ServerCount += 1
	labels := prometheus.Labels{
		"server_version": statsRequest.ServerVersion,
	}

	exporter.PluginVersion.With(labels).Inc()
	exporter.PlayerAmount.Set(PlayerCount)
	exporter.ServerAmount.Set(ServerCount)

	BackendStatsMutex.Lock()
	BackendStats[statsRequest.backendID] = statsRequest
	BackendStatsMutex.Unlock()

	go startTimeout(statsRequest.backendID)
}

func startTimeout(backendID string) {

	time.Sleep(time.Second * 30)

	BackendStatsMutex.RLock()
	lastStats, ok := BackendStats[backendID]
	BackendStatsMutex.RUnlock()
	if !ok {
		fmt.Printf("cant found key in map %s\n", backendID)
		return
	}

	if time.Now().UnixMilli()-lastStats.latestPing < 1000*30 {
		// Server did not timeout and send ping in latest 40 sec -> dont delete
		return
	}

	PlayerCount -= lastStats.PlayerAmount
	ServerCount -= 1
	labels := prometheus.Labels{
		"server_version": lastStats.ServerVersion,
	}

	exporter.PlayerAmount.Set(PlayerCount)
	exporter.ServerAmount.Set(ServerCount)
	exporter.PluginVersion.DeletePartialMatch(labels)

	BackendStatsMutex.Lock()
	delete(BackendStats, backendID)
	BackendStatsMutex.Unlock()
}
