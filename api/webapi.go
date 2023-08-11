package api

import (
	"NeoPluginMaster/exporter"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"time"
)

type stats struct {
	PlayerAmount   float64 `json:"playerAmount"`
	ManagedServers float64 `json:"managedServers"`
	OnlineMode     int     `json:"onlineMode"`
	ServerVersion  string  `json:"serverVersion"`
	ServerName     string  `json:"serverName"`
	JavaVersion    string  `json:"javaVersion"`
	OsName         string  `json:"osName"`
	OsArch         string  `json:"osArch"`
	OsVersion      string  `json:"osVersion"`
	CoreCount      int     `json:"coreCount"`
	PluginVersion  string  `json:"pluginVersion"`
	latestPing     int64
	backendID      string
}

type ResponseMessage struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}

var BackendStats = map[string]stats{}

var ServerCount = float64(0)
var PlayerCount = float64(0)

func Run() {
	http.Handle("/api/stats/plugin", pluginMetricsWithRateLimit(pluginMetrics))
	fmt.Println("API listener started at 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func pluginMetricsWithRateLimit(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	limiter := rate.NewLimiter(rate.Every(1*time.Second), 25)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
			return
		} else {
			next(w, r)
		}
	})
}

func pluginMetrics(w http.ResponseWriter, r *http.Request) {
	var statsRequest stats
	statsRequest.backendID = r.Header.Get("backendID")
	if statsRequest.backendID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&statsRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	statsRequest.latestPing = time.Now().UnixMilli()

	latestStats, ok := BackendStats[statsRequest.backendID]
	if ok {
		PlayerCount -= latestStats.PlayerAmount
		ServerCount -= 1
	}
	PlayerCount += statsRequest.PlayerAmount
	ServerCount += 1

	BackendStats[statsRequest.backendID] = statsRequest

	exporter.PlayerAmount.Set(PlayerCount)
	exporter.ServerAmount.Set(ServerCount)

	go startTimeout(statsRequest.backendID)
}

func startTimeout(backendID string) {
	// wait 10 seconds until check
	time.Sleep(time.Second * 20)
	// get stats for id
	stats, ok := BackendStats[backendID]
	if !ok {
		fmt.Printf("cant found key in map %s\n", backendID)
		return
	}
	// Check how long since latest ping
	if time.Now().UnixMilli()-stats.latestPing < 1000*20 {
		// Server did not timeout and send ping in latest 40 sec -> dont delete
		return
	}
	// Server most likely was stopped -> time out -> delete id from map
	PlayerCount -= stats.PlayerAmount
	ServerCount -= 1

	delete(BackendStats, backendID)

	exporter.PlayerAmount.Set(PlayerCount)
	exporter.ServerAmount.Set(ServerCount)
}
