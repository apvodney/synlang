package main

import (
	_ "fmt"
	_ "time"
	"os"
	"math"
	"encoding/binary"
)

var sratehz = 44100
var bdepth = math.MaxInt16
var bufsiz = 1024

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func sin(ic, oc chan []float32)  {
	for i, _ := range buf {
		buf[i] = float32(math.Sin(2.0 * math.Pi * float64(buf[i])))
	}
	return buf
}

func ramp(buf []float32) []float32 {
	l := len(buf)
	for i, _ := range buf {
		buf[i] = float32(i) / float32(l)
	}
	return buf
}

func tooutfmt(out []byte, buf []float32) []byte {
	for i, v := range buf {
		binary.LittleEndian.PutUint16(out[i*2:], uint16(v * float32(bdepth)))
	}
	return out
}

func out(c chan []float32, outf *os.File, obufsiz int) {
	var i []byte
	out := make([]byte, obufsiz)
	tobyte := make([]byte, bufsiz * 2)
	o := out
	for {
		if len(i) == 0 {
			i = tooutfmt(tobyte, <- c)
		}
		if len(o) == 0 {
			o = out
			_, err := outf.Write(out)
			check(err)
		}
		
		cpl := copy(o, i)
		if len(o) == len(i) {
			o = nil
			i = nil
			continue
		}
		if cpl == len(i) {
			o = o[len(i):]
			i = nil
		}
		if cpl == len(o) {
			i = i[len(o):]
			o = nil
		}
	}
}

func main() {
	outf, err := os.OpenFile("/dev/audio", os.O_WRONLY, 0755)
	check(err)
	
	fhunhz := int(math.Round(float64(sratehz)/400))
	sin400 := sin(ramp(make([]float32, fhunhz, fhunhz)))
	
	c := make(chan []float32)
	go out(c, outf, 1024) // Get bufsize from /dev/audioctl
	
	for {
		c <- sin400
	}
	
	_ = outf
	_ = err
	_ = sin400
}
