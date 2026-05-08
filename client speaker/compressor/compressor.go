package compressor

func BytesToInt16(b []byte) []int16 {
	nSamples := len(b) / 2
	samples := make([]int16, nSamples)

	for i := 0; i < nSamples; i++ {
		low := uint16(b[i*2])
		high := uint16(b[i*2+1])
		samples[i] = int16(low | (high << 8))
	}
	return samples
}

func PcmToUlaw(pcmsample []byte) []byte {
	pcmsample16 := BytesToInt16(pcmsample)
	out := make([]byte, len(pcmsample16))
	for i, sample := range pcmsample16 {
		out[i] = EncodeUlawpample(sample)
	}
	return out
}

func EncodeUlawpample(sample int16) byte {
	const bias = 33
	const clip = 32635

	var sign uint16
	if sample < 0 {
		sample = -sample
		sign = 0x80
	} else {
		sign = 0x00
	}
	if sample > clip {
		sample = clip
	}
	sample += bias

	expoent := uint16(7)
	for expmask := uint16(0x4000); (uint16(sample)&expmask) == 0 && expoent > 0; expmask >>= 1 {
		expoent--
	}

	mantissa := (uint16(sample) >> (expoent + 3)) & 0x0F

	return byte(^(sign | (expoent << 4) | mantissa))
}
