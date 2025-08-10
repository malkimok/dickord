package examples

import (
	"fmt"
	"time"

	"dickord/utils"

	"github.com/gordonklaus/portaudio"
)

func EchoMain() {
	err := portaudio.Initialize()
	utils.Chk(err)
	defer portaudio.Terminate()
	e := newEcho(time.Millisecond)
	defer e.Close()
	utils.Chk(e.Start())
	time.Sleep(10 * time.Second)
	utils.Chk(e.Stop())
}

type echo struct {
	*portaudio.Stream
	buffer []float32
	i      int
}

func newEcho(delay time.Duration) *echo {
	h, err := portaudio.DefaultHostApi()
	utils.Chk(err)
	fmt.Printf("Input: %v\n", h.DefaultInputDevice.Name)
	fmt.Printf("Output: %v\n", h.DefaultOutputDevice.Name)
	p := portaudio.LowLatencyParameters(h.DefaultInputDevice, h.DefaultOutputDevice)
	p.Input.Channels = 1
	p.Output.Channels = 1
	e := &echo{buffer: make([]float32, int(p.SampleRate*delay.Seconds()))}
	e.Stream, err = portaudio.OpenStream(p, e.processAudio)
	utils.Chk(err)
	return e
}

func (e *echo) processAudio(in, out []float32) {
	copy(out, in)
	for i := range out {
		out[i] *= 5
	}
}
