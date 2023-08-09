package main

import (
	"NeoPluginMaster/api"
	"NeoPluginMaster/exporter"
)

func main() {
	go exporter.Run()
	api.Run()
}
