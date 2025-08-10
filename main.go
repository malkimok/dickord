package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"dickord/utils"

	"github.com/gordonklaus/portaudio"
)

func getDevice() (*portaudio.DeviceInfo, error) {
	devices, err := portaudio.Devices()
	utils.Chk(err)
	if len(devices) <= 0 {
		return &portaudio.DeviceInfo{}, errors.New("No available devices")
	}

	for i, d := range devices {
		fmt.Printf(
			"%v. %v (input: %v, out: %v, sr: %v)\n",
			i, d.Name, d.MaxInputChannels, d.MaxOutputChannels, d.DefaultSampleRate,
		)
	}
	fmt.Print("Input device number: ")
	var userInput string
	fmt.Scanln(&userInput)
	userInputInt, err := strconv.Atoi(userInput)
	utils.Chk(err)
	return devices[userInputInt], nil
}

func PenisMain() {
	if len(os.Args) < 2 {
		fmt.Println("missing required argument:  output file name")
		return
	}

	err := portaudio.Initialize()
	utils.Chk(err)
	defer portaudio.Terminate()

	var inDev, outDev *portaudio.DeviceInfo
	fmt.Println("Select input device:")
	inDev, err = getDevice()
	utils.Chk(err)

	fmt.Println()

	fmt.Println("Select output device:")
	outDev, err = getDevice()
	utils.Chk(err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	fileName := os.Args[1]
	if !strings.HasSuffix(fileName, ".aiff") {
		fileName += ".aiff"
	}
	f, err := os.Create(fileName)
	utils.Chk(err)

	// form chunk
	_, err = f.WriteString("FORM")
	utils.Chk(err)
	utils.Chk(binary.Write(f, binary.BigEndian, int32(0))) // total bytes
	_, err = f.WriteString("AIFF")
	utils.Chk(err)

	// common chunk
	_, err = f.WriteString("COMM")
	utils.Chk(err)
	utils.Chk(binary.Write(f, binary.BigEndian, int32(18)))            // size
	utils.Chk(binary.Write(f, binary.BigEndian, int16(1)))             // channels
	utils.Chk(binary.Write(f, binary.BigEndian, int32(0)))             // number of samples
	utils.Chk(binary.Write(f, binary.BigEndian, int16(32)))            // bits per sample
	_, err = f.Write([]byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}) // 80-bit sample rate 44100
	utils.Chk(err)

	// sound chunk
	_, err = f.WriteString("SSND")
	utils.Chk(err)
	utils.Chk(binary.Write(f, binary.BigEndian, int32(0))) // size
	utils.Chk(binary.Write(f, binary.BigEndian, int32(0))) // offset
	utils.Chk(binary.Write(f, binary.BigEndian, int32(0))) // block
	nSamples := 0
	defer func() {
		// fill in missing sizes
		totalBytes := 4 + 8 + 18 + 8 + 8 + 4*nSamples
		_, err = f.Seek(4, 0)
		utils.Chk(err)
		utils.Chk(binary.Write(f, binary.BigEndian, int32(totalBytes)))
		_, err = f.Seek(22, 0)
		utils.Chk(err)
		utils.Chk(binary.Write(f, binary.BigEndian, int32(nSamples)))
		_, err = f.Seek(42, 0)
		utils.Chk(err)
		utils.Chk(binary.Write(f, binary.BigEndian, int32(4*nSamples+8)))
		utils.Chk(f.Close())
	}()

	in := make([]int32, 64)
	out := make([]int32, 64)

	p := portaudio.LowLatencyParameters(inDev, outDev)
	numChannels := min(inDev.MaxInputChannels, outDev.MaxOutputChannels)
	p.Input.Channels = numChannels
	p.Output.Channels = numChannels
	// pIn.SampleRate = inDev.DefaultSampleRate
	p.FramesPerBuffer = len(in)

	stream, err := portaudio.OpenStream(p, in, out)
	utils.Chk(err)
	defer stream.Close()
	utils.Chk(stream.Start())

	fmt.Println("Recording. Press Ctrl-C to stop.")

Penis:
	for {
		utils.Chk(stream.Read())
		copy(out, in)
		utils.Chk(stream.Write())
		// utils.Chk(binary.Write(f, binary.BigEndian, in))
		nSamples += len(in)
		select {
		case <-sig:
			break Penis
		default:
		}
	}
	utils.Chk(stream.Stop())
}

func main() {
	// EchoMain()
	PenisMain()
	// PlayMain()
}
