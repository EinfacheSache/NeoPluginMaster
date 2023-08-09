package api

import (
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"time"
)

type stats struct {
	BackendID      string  `json:"backendID"`
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
	limiter := rate.NewLimiter(2, 1)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			message := ResponseMessage{
				Status: "Rate limit exceeded",
				Body:   "You are being rate limitted. Please try again later.",
			}

			w.WriteHeader(http.StatusTooManyRequests)
			err := json.NewEncoder(w).Encode(&message)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		} else {
			next(w, r)
		}
	})
}

func pluginMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var statsRequest stats
	err := json.NewDecoder(r.Body).Decode(&statsRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	// Set latestping ...
	statsRequest.latestPing = time.Now().UnixMilli()

	latestStats, ok := BackendStats[statsRequest.BackendID]
	if ok {
		PlayerCount -= latestStats.PlayerAmount
		ServerCount -= latestStats.ManagedServers
	}
	PlayerCount += statsRequest.PlayerAmount
	ServerCount += latestStats.ManagedServers

	BackendStats[statsRequest.BackendID] = statsRequest

	go startTimeout(statsRequest.BackendID)

}

func startTimeout(backendID string) {
	// wait 60 seconds until check
	time.Sleep(time.Second * 30)
	// get stats for id
	stats, ok := BackendStats[backendID]
	if !ok {
		// Key isn't in map anymore
		return
	}
	// Check how long since latest ping
	if time.Now().UnixMilli()-stats.latestPing < 1000*40 {
		// Server did not timeout and send ping in latest 40 sec -> dont delete
		return
	}
	// Server most likely was stopped -> time out -> delete id from map
	PlayerCount -= stats.PlayerAmount
	ServerCount -= stats.ManagedServers
	delete(BackendStats, backendID)
}
