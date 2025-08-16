package hub_client

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
		log.Fatalln("error: can't get default host portaudio api:", err)
	}
	fmt.Printf("Output: %v\n", h.DefaultOutputDevice.Name)
	fmt.Printf("Input: %v\n", h.DefaultInputDevice.Name)
	p := portaudio.LowLatencyParameters(h.DefaultInputDevice, h.DefaultOutputDevice)
	p.Input.Channels = 1
	p.Output.Channels = 1
	stream, err := portaudio.OpenStream(p, callback)
	if err != nil {
		log.Fatalln("error: can't open stream:", err)
	}
	return stream
}

func readPump(conn *websocket.Conn, readCh chan []byte) {
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: websocket upexpectedly closed: %v", err)
			} else {
				log.Printf("error: when reading message from client, disconnecting: %v", err)
			}
			break
		}
		readCh <- message
	}
}

func getDuplexCallback(conn *websocket.Conn, readCh chan []byte) func([]float32, []float32) {
	return func(in, out []float32) {
		inBytes := floatToByte(in)
		if err := conn.WriteMessage(websocket.BinaryMessage, inBytes); err != nil {
			log.Println("error: can't write to socket from input stream:", err)
		}
		log.Printf("debug: %v bytes sent to server:\n", len(inBytes))

		// TODO: add support for more than 2 clients
		// (intput and output sizes are not 1 to 1 when more than 2 people on the call)
		select {
		case p := <-readCh:
			copy(out, byteToFloat(p))
			log.Printf("debug: %v bytes writen to out\n", len(p))
		default:
			log.Println("debug: no data for read")
		}
	}
}

func floatToByte(fs []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&fs[0])), len(fs)*4)
}

func byteToFloat(bs []byte) []float32 {
	return unsafe.Slice((*float32)(unsafe.Pointer(&bs[0])), len(bs)/4)
}

func RunClient() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, os.Kill)

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://localhost:1488/ws", http.Header{})
	if err != nil {
		log.Fatalln("error: can't dial a server:", err)
	}
	log.Println("info: client started on", conn.LocalAddr().String())
	defer conn.Close()

	err = portaudio.Initialize()
	if err != nil {
		log.Fatal("error: can't initialize portaudio in dialer: ", err)
	}
	defer portaudio.Terminate()

	readCh := make(chan []byte, 256)
	go readPump(conn, readCh)

	duplexStream := getDuplexStream(getDuplexCallback(conn, readCh))
	defer duplexStream.Close()

	if err = duplexStream.Start(); err != nil {
		log.Fatal("error: can't start duplex stream: ", err)
	}

	<-sig

	if err = duplexStream.Stop(); err != nil {
		log.Fatal("error: can't stop duplex stream: ", err)
	}
}
