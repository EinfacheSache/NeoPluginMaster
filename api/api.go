package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type response struct {
	IP string `json:"ip"`
}

func Run() {
	http.HandleFunc("/api/stats/plugin", pluginMetrics)
	fmt.Println("API listener started at 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func pluginMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var joinResponse response
	err := json.NewDecoder(r.Body).Decode(&joinResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
