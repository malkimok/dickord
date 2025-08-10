package examples

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"

	"dickord/utils"

	"github.com/gordonklaus/portaudio"
)

func PlayMain() {
	if len(os.Args) < 2 {
		fmt.Println("missing required argument:  input file name")
		return
	}
	fmt.Println("Playing.  Press Ctrl-C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	fileName := os.Args[1]
	f, err := os.Open(fileName)
	utils.Chk(err)
	defer f.Close()

	id, data, err := readChunk(f)
	utils.Chk(err)
	if id.String() != "FORM" {
		fmt.Println("bad file format")
		return
	}
	_, err = data.Read(id[:])
	utils.Chk(err)
	if id.String() != "AIFF" {
		fmt.Println("bad file format")
		return
	}
	var c commonChunk
	var audio io.Reader
	for {
		id, chunk, err := readChunk(data)
		if err == io.EOF {
			break
		}
		utils.Chk(err)
		switch id.String() {
		case "COMM":
			utils.Chk(binary.Read(chunk, binary.BigEndian, &c))
		case "SSND":
			chunk.Seek(8, 1) //ignore offset and block
			audio = chunk
		default:
			fmt.Printf("ignoring unknown chunk '%s'\n", id)
		}
	}

	//assume 44100 sample rate, mono, 32 bit

	portaudio.Initialize()
	defer portaudio.Terminate()
	out := make([]int32, 8192)
	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, len(out), &out)
	utils.Chk(err)
	defer stream.Close()

	utils.Chk(stream.Start())
	defer stream.Stop()
	for remaining := int(c.NumSamples); remaining > 0; remaining -= len(out) {
		if len(out) > remaining {
			out = out[:remaining]
		}
		err := binary.Read(audio, binary.BigEndian, out)
		if err == io.EOF {
			break
		}
		utils.Chk(err)
		utils.Chk(stream.Write())
		select {
		case <-sig:
			return
		default:
		}
	}
}

func readChunk(r readerAtSeeker) (id ID, data *io.SectionReader, err error) {
	_, err = r.Read(id[:])
	if err != nil {
		return
	}
	var n int32
	err = binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return
	}
	off, _ := r.Seek(0, 1)
	data = io.NewSectionReader(r, off, int64(n))
	_, err = r.Seek(int64(n), 1)
	return
}

type readerAtSeeker interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

type ID [4]byte

func (id ID) String() string {
	return string(id[:])
}

type commonChunk struct {
	NumChans      int16
	NumSamples    int32
	BitsPerSample int16
	SampleRate    [10]byte
}
