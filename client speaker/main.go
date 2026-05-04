package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gen2brain/malgo"
)

func main() {
	ip := ":8080"
	ip = "192.168.1.109:8080"
	conn, err := net.Dial("udp", ip)
	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

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

	//var buffer []byte
	//DataLength := 880

	onSamples := func(pOutputSample, pInputSamples []byte, frameCount uint32) {
		if len(pInputSamples) > 0 {
			//buffer = append(buffer, pInputSamples...)
			//if len(buffer) >= DataLength {
			//	packet := buffer[:DataLength]
			//
			//	_, err := conn.Write(packet)
			//	if err != nil {
			//		log.Fatalln(err)
			//	}
			//
			//	buffer = buffer[DataLength:]
			//}

			_, err := conn.Write(pInputSamples)
			if err != nil {
				log.Fatalln(err)
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

	fmt.Println("Áudio inicializado. Pressione ENTER para sair...")
	fmt.Scanln()
}
