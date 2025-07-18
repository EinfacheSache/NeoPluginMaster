package main

import (
	"fmt"
	"plugin-master/api"
	"plugin-master/exporter"
	"time"
)

func main() {
	go exporter.Run()
	go webapi.Run()
	writer()
}

func writer() {

	time.Sleep(time.Millisecond * 500)

	for {
		webapi.AmountStatsMutex.RLock()

		var bungeecordPlayerCount = webapi.AmountStats["bungeecordPlayerCount"]
		var velocityPlayerCount = webapi.AmountStats["velocityPlayerCount"]
		var spigotPlayerCount = webapi.AmountStats["spigotPlayerCount"]
		var restPlayerCount = webapi.AmountStats["PlayerCount"]

		webapi.AmountStatsMutex.RUnlock()

		fmt.Println(
			"PlayerCount (", bungeecordPlayerCount+velocityPlayerCount+spigotPlayerCount+restPlayerCount, ") [ "+
				"Bungee(", bungeecordPlayerCount, ") "+
				"Velocity(", velocityPlayerCount, ") "+
				"Spigot(", spigotPlayerCount, ") "+
				"Rest(", restPlayerCount, ") ]")

		time.Sleep(time.Second * 5)
	}
}
