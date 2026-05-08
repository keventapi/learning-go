package decode

func DecodeUlawPcm(ulaw []byte) []byte {
	pcm := make([]byte, len(ulaw)*2)
	for i, sample := range ulaw {
		value := DecodeUlawSample(sample)
		pcm[i*2] = byte(value & 0xFF)
		pcm[i*2+1] = byte((value >> 8) & 0xFF)
	}
	return pcm
}

func DecodeUlawSample(sample byte) int16 {
	sample = ^sample

	signal := sample & 0x80
	expoent := (sample >> 4) & 0x07
	mantissa := sample & 0x0F

	Pcm := int16((uint16(mantissa)<<3)+33) << int16(expoent)
	Pcm -= 33

	if signal != 0 {
		return -Pcm
	}
	return Pcm
}
