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

func handle_connection(conn net.Conn, rb_output *ringbuffer.Buffer, rb_input *ringbuffer.Buffer) {
	defer conn.Close()
	buffer_output := make([]byte, 960*2)
	for {
		conn.SetDeadline(time.Now().Add(40 * time.Millisecond))
		_, err := io.ReadFull(conn, buffer_output)
		if err != nil {
			continue
		}
		rb_output.Write(buffer_output)

		packet := make([]byte, 960*2)
		read_len, packet := rb_output.Read(packet, packet)
		if read_len > 0 {
			conn.Write(packet[:read_len])
		}
	}
}

func main() {
	rb_output := ringbuffer.New((48000 * 2 * 2) * 5 * 3) // 15s de buffer
	rb_input := ringbuffer.New((44100 * 2 * 2) * 2)      // 2s

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

	// output
	go func() {
		deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
		deviceConfig.Playback.Format = malgo.FormatS16
		deviceConfig.Playback.Channels = 2
		deviceConfig.SampleRate = 44100

		onSamples := func(pOutputSample, pInputSamples []byte, frameCount uint32) {
			packet := make([]byte, len(pOutputSample))
			read_len, packet := rb_output.Read(pOutputSample, packet)
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
		fmt.Scanln()
	}()

	// input
	go func() {
		deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
		deviceConfig.Capture.Format = malgo.FormatS16
		deviceConfig.Capture.Channels = 2
		deviceConfig.SampleRate = 44100

		onSamples := func(pOutputSample, pInputSamples []byte, frameCount uint32) {
			rb_input.Write(pOutputSample)
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
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("erro ao aceitar conexão")
			continue
		}

		go handle_connection(conn, rb_output, rb_input)
	}
}
