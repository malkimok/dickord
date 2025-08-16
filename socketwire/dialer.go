package socketwire

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"time"
	"unsafe"

	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

func getRemoteStream(callback func([]float32)) *portaudio.Stream {
	h, err := portaudio.DefaultHostApi()
	if err != nil {
		log.Fatal("Error getting Default Host Api", err)
	}
	fmt.Printf("Input: %v\n", h.DefaultInputDevice.Name)
	p := portaudio.LowLatencyParameters(h.DefaultInputDevice, nil)
	p.Input.Channels = 1
	s, err := portaudio.OpenStream(p, callback)
	if err != nil {
		log.Fatal("Error opening stream", err)
	}
	return s
}

func getAudioWriter(conn *websocket.Conn) func([]float32) {
	return func(in []float32) {
		w, err := conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			log.Fatal(err)
		}
		binary.Write(w, binary.BigEndian, in)
		if err := w.Close(); err != nil {
			log.Fatal(err)
		}
	}
}

func getUnsafeAudioWriter(conn *websocket.Conn) func([]float32) {
	return func(in []float32) {
		inBytes := floatToByte(in)
		if err := conn.WriteMessage(websocket.BinaryMessage, inBytes); err != nil {
			log.Println(err)
		}
	}
}

func floatToByte(fs []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&fs[0])), len(fs)*4)
}

func byteToFloat(bs []byte) []float32 {
	return unsafe.Slice((*float32)(unsafe.Pointer(&bs[0])), len(bs)/4)
}

func RunDialer() {
	err := portaudio.Initialize()
	if err != nil {
		log.Fatal("Error initializing portaudio in dialer", err)
	}
	defer portaudio.Terminate()

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://localhost:1488/ws", http.Header{})
	if err != nil {
		log.Fatal("Error while dial", err)
	}
	defer conn.Close()

	stream := getRemoteStream(getUnsafeAudioWriter(conn))
	defer stream.Close()
	if err := stream.Start(); err != nil {
		log.Fatal("Error starting dialer stream")
	}
	time.Sleep(100 * time.Second)
	if err := stream.Stop(); err != nil {
		log.Fatal("Error stoping dialer stream", err)
	}
}
