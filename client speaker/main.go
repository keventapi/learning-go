package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/gen2brain/malgo"
)

func udp_connection(buff []byte) {
	conn, err := net.Dial("udp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

	_, err = conn.Write(buff)
	if err != nil {
		log.Fatalln(err)
	}

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	n, err := conn.Read(buff)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("recieved: ", buff[:n])
}

func main() {
	// 1. Inicializar o contexto da Malgo
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		fmt.Printf("Erro ao inicializar contexto: %v\n", err)
		os.Exit(1)
	}
	defer ctx.Uninit() // Garante a limpeza dos recursos

	// Em vez de Playback, usamos Capture
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Loopback)

	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 2
	deviceConfig.SampleRate = 44100
	// No Windows, isso ativa a captura do que você está ouvindo
	//deviceConfig.Wasapi.Usage = malgo.Loopback

	// 3. Definir o Callback (onde o áudio é processado)
	//var audioBuffer []byte

	onSamples := func(pOutputSample, pInputSamples []byte, frameCount uint32) {
		// pInputSamples contém o que está sendo capturado agora
		if len(pInputSamples) > 0 {
			go udp_connection(pInputSamples[:1024])
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

	// 5. Iniciar o fluxo de áudio
	err = device.Start()
	if err != nil {
		fmt.Printf("Erro ao iniciar dispositivo: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Áudio inicializado. Pressione ENTER para sair...")
	fmt.Scanln()
}
