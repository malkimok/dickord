package hub

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

// func getInputStream(callback func([]float32)) *portaudio.Stream {
// 	h, err := portaudio.DefaultHostApi()
// 	if err != nil {
// 		log.Fatal("Error getting Default Host Api", err)
// 	}
// 	fmt.Printf("Input: %v\n", h.DefaultInputDevice.Name)
// 	p := portaudio.LowLatencyParameters(h.DefaultInputDevice, nil)
// 	p.Input.Channels = 1
// 	stream, err := portaudio.OpenStream(p, callback)
// 	if err != nil {
// 		log.Fatal("Error opening stream", err)
// 	}
// 	return stream
// }

// func getOutputStream(callback func([]float32)) *portaudio.Stream {
// 	h, err := portaudio.DefaultHostApi()
// 	if err != nil {
// 		log.Fatal("Error getting Default Host Api", err)
// 	}
// 	fmt.Printf("Output: %v\n", h.DefaultOutputDevice.Name)
// 	p := portaudio.LowLatencyParameters(nil, h.DefaultOutputDevice)
// 	p.Output.Channels = 1
// 	stream, err := portaudio.OpenStream(p, callback)
// 	if err != nil {
// 		log.Fatal("Error opening stream", err)
// 	}
// 	return stream
// }

// func getInputCallback(conn *websocket.Conn) func([]float32) {
// 	return func(in []float32) {
// 		inBytes := floatToByte(in)
// 		if err := conn.WriteMessage(websocket.BinaryMessage, inBytes); err != nil {
// 			log.Println(err)
// 		}
// 	}
// }

// func getOutputCallback(conn *websocket.Conn) func([]float32) {
// 	return func(out []float32) {
// 		_, p, err := conn.ReadMessage()
// 		if err != nil {
// 			log.Printf("error: can't write from socket to outstream: %v", err)
// 		}
// 		out = byteToFloat(p)
// 	}
// }

func getDuplexStream(callback func([]float32, []float32)) *portaudio.Stream {
	h, err := portaudio.DefaultHostApi()
	if err != nil {
		log.Fatal("Error getting Default Host Api", err)
	}
	fmt.Printf("Output: %v\n", h.DefaultOutputDevice.Name)
	p := portaudio.LowLatencyParameters(h.DefaultInputDevice, h.DefaultOutputDevice)
	p.Input.Channels = 1
	p.Output.Channels = 1
	stream, err := portaudio.OpenStream(p, callback)
	if err != nil {
		log.Fatal("Error opening stream", err)
	}
	return stream
}

func getDuplexCallback(conn *websocket.Conn) func([]float32, []float32) {
	return func(in, out []float32) {
		inBytes := floatToByte(in)
		if err := conn.WriteMessage(websocket.BinaryMessage, inBytes); err != nil {
			log.Println("error: can't write to socket from input stream: %v", err)
		}
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error: can't write from socket to outstream: %v", err)
		}
		out = byteToFloat(p)
	}
}

func floatToByte(fs []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&fs[0])), len(fs)*4)
}

func byteToFloat(bs []byte) []float32 {
	return unsafe.Slice((*float32)(unsafe.Pointer(&bs[0])), len(bs)/4)
}

func RunClient() {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://localhost:1488/ws", http.Header{})
	if err != nil {
		log.Fatal("Error while dial", err)
	}
	defer conn.Close()

	err = portaudio.Initialize()
	if err != nil {
		log.Fatal("Error initializing portaudio in dialer", err)
	}
	defer portaudio.Terminate()

	duplexStream := getDuplexStream(getDuplexCallback(conn))
	defer duplexStream.Close()

	if err = duplexStream.Start(); err != nil {
		log.Fatal("Error starting duplex stream", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	if err = duplexStream.Stop(); err != nil {
		log.Fatal("Error stoping duplex stream", err)
	}
}
