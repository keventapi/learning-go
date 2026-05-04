package main

// just to change something
import (
	"fmt"
	"localspeaker/ringbuffer"
	"log"
	"net"
	"os"

	"github.com/gen2brain/malgo"
)

func main() {
	rb := ringbuffer.New((44100 * 2 * 2) * 5 * 3)

	addrs, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}
	conn, err := net.ListenUDP("udp", addrs)

	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

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
		buff := make([]byte, 1780)
		n, ClientAddrs, err := conn.ReadFromUDP(buff)

		if err != nil {
			println(n, ClientAddrs)
			log.Println(err)
			continue
		}

		rb.Write(buff)
	}
}
