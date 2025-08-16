package main

import (
	"dickord/hub"
	"flag"
	"fmt"
)

var addr = flag.String("addr", ":1488", "http service address")

func main() {
	flag.Parse()
	fmt.Println("Server listening on:", *addr)
	hub.RunHubServer(*addr)
}
