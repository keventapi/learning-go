package main

import (
	"clientspeaker/ringbuffer"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/gen2brain/malgo"
)

func GetDeviceId(ctx *malgo.AllocatedContext) *malgo.DeviceID {
	devices, _ := ctx.Devices(malgo.Playback)
	for _, info := range devices {
		name := strings.ToLower(info.Name())

		if strings.Contains(name, "cable") && strings.Contains(name, "input") {
			fmt.Println("device para output encontrado: ", name)
			id := info.ID
			return &id
		}
	}
	return nil
}

func main() {
	rb := ringbuffer.New((44100 * 2 * 2) * 2)
	ip := "192.168.1.109:8080"
	addrs, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, addrs)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	go func() {
		for {
			buff := make([]byte, 960*2)
			n, _ := io.ReadFull(conn, buff)
			rb.Write(buff[:n])
		}
	}()

	defer conn.Close()

	AudioChan := make(chan []byte, 100)
	go func() {
		for sample := range AudioChan {
			_, err := conn.Write(sample)
			if err != nil {
				log.Println("couldnt write to the connection, package loss", err)
			}
		}
	}()

	// output handler
	go func() {
		ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
		defer ctx.Uninit()

		deviceConfig := malgo.DefaultDeviceConfig(malgo.Loopback)
		deviceConfig.Capture.Format = malgo.FormatS16
		deviceConfig.Capture.Channels = 2
		deviceConfig.SampleRate = 44100

		onSamples := func(pOutputSample, pInputSamples []byte, frameCount uint32) {
			if len(pInputSamples) > 0 {
				buff := make([]byte, len(pInputSamples))
				copy(buff, pInputSamples)

				select {
				case AudioChan <- buff:
				default:
					log.Println("channel stacked up")
				}

			}
		}

		device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{
			Data: onSamples,
		})
		if err != nil {
			log.Fatalln("Error to create a device: ", err)
			os.Exit(1)
		}

		defer device.Uninit()

		err = device.Start()
		if err != nil {
			log.Fatalln("Error to init a device: ", err)
			os.Exit(1)
		}
		fmt.Scanln()
	}()

	// input handler
	go func() {
		ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
		defer ctx.Uninit()
		id := GetDeviceId(ctx)

		deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
		if id != nil {
			deviceConfig.Playback.DeviceID = id.Pointer()
		} else {
			return
		}
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
			log.Fatalln("Error to create a device: ", err)
			os.Exit(1)
		}

		defer device.Uninit()

		err = device.Start()
		if err != nil {
			log.Fatalln("Error to init a device: ", err)
			os.Exit(1)
		}
		fmt.Scanln()
	}()

	fmt.Println("Áudio inicializado. Pressione ENTER para sair...")
	fmt.Scanln()
}
