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
	for { // Endlosschleife
		time.Sleep(time.Second * 1)
		fmt.Println("ServerCount ", api.ServerCount)
		fmt.Println("PlayerCount ", api.PlayerCount)
	}
}
