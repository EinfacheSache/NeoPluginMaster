package main

import (
	"fmt"
	"neo-plugin-master/api"
	"neo-plugin-master/exporter"
	"time"
)

func main() {
	go exporter.Run()
	go api.Run()
	reader()
}

func reader() {
	for {
		api.AmountStatsMutex.RLock()
		fmt.Println("PlayerCount ", api.AmountStats["PlayerCount"]+api.AmountStats["bungeecordPlayerCount"]+api.AmountStats["velocityPlayerCount"]+api.AmountStats["spigotPlayerCount"], "[Bungee(", api.AmountStats["bungeecordPlayerCount"], ") Velocity(", api.AmountStats["velocityPlayerCount"], ") Spigot(", api.AmountStats["spigotPlayerCount"], ") Rest(", api.AmountStats["PlayerCount"], "]")
		api.AmountStatsMutex.RUnlock()
		time.Sleep(time.Second * 5)
	}
}
