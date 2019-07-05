package main

import (
	"os"
	"math"
	"encoding/binary"
	"strconv"
)

// The asteroid to wipe out these globals is still in orbit
var orate = 44100          // output sample rate in hertz
var odepth = math.MaxInt16 // output bit-depth
var irate = orate          // internal sample rate

type Sample float32

type ModIO struct {
	i map[string](Recver),
	o map[string](Sender),
	p map[string]string,
}

type Modentry struct {
	Func        func(ModIO)
	Inputs      []string                // Sample stream inputs
	Parameters  []string                // Static parameters
	Outputs     []string                // Sample stream outputs
}
var Modtbl map[string]Modentry

func init() {
	Modtbl = make(map[string]Modentry)
}

type Sender interface {
	Send(Sample)
}

type Recver interface {
	Recv() Sample
}

func init() {
	Modtbl["constval"] = Modentry{
		Func: constval,
		Outputs: []string{ "out" },
		Parameters: []string{ "val" },
	}
}

func constval(m ModIO) {
	out := m.o["out"]
	val, err := strconv.ParseFloat(m.p["val"], 32)
	check(err)
	for {
		out.Send(Sample(val))
	}
}

func ewrite(outf *os.File, obuf []byte) {
	for {
		_, err := outf.Write(obuf)
		if err == nil {
			break
		}
		if IsMundaneError(err) {
			continue
		}
		check(err)
	}
}

func init() {
	Modtbl["out"] = Modentry{
		// out is a special case, its function doesn't fit the mold
		Inputs: []string{ "in" },
		Parameters: []string{ "stereo" },
	}
}

func out(m ModIO, outf *os.File, obufsiz int) {
	in := m.i["in"]
	obuf := make([]byte, obufsiz)
	to := obuf
	
	bc := make(chan byte)
	sc := make(chan Sample)
	go func() {
		for {
			s := in.Recv()
			sc <- s
			sc <- s
		}
	}()
	go func() {
		bs := make([]byte, 2)
		for {
			s := <- sc
			binary.LittleEndian.PutUint16(bs, uint16(s * Sample(odepth)))
			for _, b := range bs {
				bc <- b
			}
		}
	}()
	for b := range bc {
		if len(to) == 0 {
			to = obuf
			ewrite(outf, obuf)
		}
		to[0] = b
		to = to[1:]
	}
}

func init() {
	Modtbl["add"] = Modentry{
		Func: add,
		Inputs: []string{ "in1", "in2" },
		Outputs: []string{ "out" },
	}
}

func add(m ModIO) {
	in1 := m.i["in1"]
	in2 := m.i["in2"]
	out := m.o["out"]
	for {
		s1, s2 := in1.Recv(), in2.Recv()
		out.Send(s1 + s2)
	}
}

func init() {
	Modtbl["mul"] = Modentry{
		Func: mul,
		Inputs: []string{ "in1", "in2" },
		Outputs: []string{ "out" },
	}
}

func mul(m ModIO) {
	in1 := m.i["in1"]
	in2 := m.i["in2"]
	out := m.o["out"]
	for {
		s1, s2 := in1.Recv(), in2.Recv()
		out.Send(s1 * s2)
	}
}

func init() {
	Modtbl["ramp"] = Modentry{
		Func: ramp,
		Inputs: []string{ "freq" },
		Outputs: []string{ "out" },
	}
}

func ramp(m ModIO) {
	var s, slope float64
	var f, lastf Sample
	ifreq := m.i["freq"]
	out := m.o["out"]
	for {
		if f == 0 {
			slope = 0
		} else if f != lastf {
			slope = 1 / (float64(irate) / float64(f))
		}
		
		s += slope
		s = math.Mod(s, 1)
		
		lastf = f
		f = ifreq.Recv()
		
		out.Send(Sample(s))
	}
}

func init() {
	Modtbl["sinshp"] = Modentry{
		Func: sinshp,
		Inputs: []string{ "in" },
		Outputs: []string{ "out" },
	}
}

func sinshp(m ModIO) {
	in := m.i["in"]
	out := m.o["out"]
	for {
		s := in.Recv()
		out.Send(Sample(math.Sin(2.0 * math.Pi * float64(s))))
	}
}
