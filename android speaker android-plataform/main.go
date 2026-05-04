package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gen2brain/malgo"
)

type RingBuffer struct {
	data []byte
	head int
	tail int
	size int
}

func playback(rb *RingBuffer) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		fmt.Printf("Erro ao inicializar contexto: %v\n", err)
		os.Exit(1)
	}
	defer ctx.Uninit()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)

	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = 44100

	onSamples := func(pOutputSample, pInputSamples []byte, frameCount uint32) {
		read_len := rb.Read(pOutputSample)
		if read_len < len(pOutputSample) {
			for i := read_len; i < len(pOutputSample); i++ {
				pOutputSample[i] = 0
			}
		}
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: onSamples,
	})
	if err != nil {
		fmt.Printf("Erro ao inicializar dispositivo: %v\n", err)
		os.Exit(1)
	}
	defer device.Uninit()

	err = device.Start()
	if err != nil {
		fmt.Printf("Erro ao iniciar dispositivo: %v\n", err)
		os.Exit(1)
	}
	fmt.Scanln()
}

func (rb *RingBuffer) Read(p []byte) int {
	n := 0
	for i := 0; i < len(p); i++ {
		if rb.tail == rb.head {
			break
		}

		p[i] = rb.data[rb.tail]

		rb.data[rb.tail] = 0

		rb.tail = (rb.tail + 1) % rb.size
		n++
	}
	return n
}

func (rb *RingBuffer) write(data []byte) int {
	n := 0
	for i := 0; i < len(data); i++ {
		if ((rb.head + 1) % rb.size) == rb.tail {
			break
		}
		rb.data[rb.head] = data[i]
		n++
		rb.head = (rb.head + 1) % rb.size
	}
	return n
}

func main() {
	var rb RingBuffer
	rb.head = 0
	rb.tail = 0
	rb.size = 44100 * 2 * 2
	rb.data = make([]byte, rb.size)

	addrs, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}
	conn, err := net.ListenUDP("udp", addrs)

	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

	go playback(&rb)

	for {
		buff := make([]byte, 1780)
		n, ClientAddrs, err := conn.ReadFromUDP(buff)

		if err != nil {
			println(n, ClientAddrs)
			log.Println(err)
			continue
		}

		rb.write(buff)

	}
}
