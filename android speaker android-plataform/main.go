package main

import (
	"fmt"
	"io"
	"localspeaker/ringbuffer"
	"log"
	"net"
	"os"
	"time"

	"github.com/gen2brain/malgo"
)

func handle_connection(conn net.Conn, rb *ringbuffer.Buffer) {
	defer conn.Close()
	buffer := make([]byte, 960*2)
	for {
		conn.SetDeadline(time.Now().Add(40 * time.Millisecond))
		_, err := io.ReadFull(conn, buffer)
		if err != nil {
			continue
		}
		rb.Write(buffer)
	}
}

func main() {
	rb := ringbuffer.New((48000 * 2 * 2) * 5 * 3) // 15s de buffer
	addrs, err := net.ResolveTCPAddr("tcp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}
	listener, err := net.ListenTCP("tcp", addrs)

	if err != nil {
		log.Fatalln(err)
	}

	defer listener.Close()

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		fmt.Printf("Erro ao inicializar contexto: %v\n", err)
		os.Exit(1)
	}
	defer ctx.Uninit()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)

	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 2
	deviceConfig.SampleRate = 44100

	onSamples := func(pOutputSample, pInputSamples []byte, frameCount uint32) {
		packet := make([]byte, len(pOutputSample))
		read_len, packet := rb.Read(pOutputSample, packet)
		copy(pOutputSample, packet)
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

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("erro ao aceitar conexão")
			continue
		}

		go handle_connection(conn, rb)
	}
}
