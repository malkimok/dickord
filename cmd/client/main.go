package main

import (
	"dickord/hub_client"
	"flag"
)

var addr = flag.String("addr", "ws://localhost:1488/ws", "server websocket addres")

func main() {
	flag.Parse()
	hub_client.RunClient(*addr)
}
