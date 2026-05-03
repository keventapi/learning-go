package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gen2brain/malgo"
)

func udp_connection(buff []byte) {
	conn, err := net.Dial("udp", "192.168.1.109:8080")
	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

	_, err = conn.Write(buff)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		fmt.Printf("Erro ao inicializar contexto: %v\n", err)
		os.Exit(1)
	}
	defer ctx.Uninit() // Garante a limpeza dos recursos

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Loopback)

	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = 44100

	onSamples := func(pOutputSample, pInputSamples []byte, frameCount uint32) {
		// pInputSamples contém o que está sendo capturado agora
		if len(pInputSamples) > 0 {
			go udp_connection(pInputSamples)
		}
	}

	// 4. Inicializar o dispositivo
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

	fmt.Println("Áudio inicializado. Pressione ENTER para sair...")
	fmt.Scanln()
}
