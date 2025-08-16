package socketwire

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":1488", "http service address")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func getDefaultOutStream(b *[]float32) *portaudio.Stream {
	h, err := portaudio.DefaultHostApi()
	if err != nil {
		log.Fatal("Error using DefaultHostApi: ", err)
	}
	fmt.Printf("Output: %v\n", h.DefaultOutputDevice.Name)
	p := portaudio.LowLatencyParameters(nil, h.DefaultOutputDevice)
	p.Output.Channels = 1
	stream, err := portaudio.OpenStream(p, b)
	if err != nil {
		log.Fatal("Error opening stream: ", err)
	}
	return stream
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error upgrading http connection: ", err)
	}
	defer conn.Close()

	var write_buff []float32
	stream := getDefaultOutStream(&write_buff)
	stream.Start()
	defer stream.Stop()
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Fatal("Error reading message: ", err)
		}
		write_buff = byteToFloat(p)
		stream.Write()
	}
}

func RunListener() {
	flag.Parse()
	if err := portaudio.Initialize(); err != nil {
		log.Fatal("Error initializing portaudio: ", err)
	}
	defer portaudio.Terminate()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
