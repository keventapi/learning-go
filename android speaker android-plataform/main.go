package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gen2brain/malgo"
)

func playback(buff []byte) {
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
		copy(pOutputSample, buff[:1024])
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

}

func main() {
	addrs, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}
	conn, err := net.ListenUDP("udp", addrs)

	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

	for {
		buff := make([]byte, 1024)
		n, ClientAddrs, err := conn.ReadFromUDP(buff)
		if err != nil {
			log.Println(err)
			continue
		}

		go playback(buff[:n])

		_, err = conn.WriteToUDP([]byte("ola mundo"), ClientAddrs)

		if err != nil {
			log.Println(err)
		}

	}
}
