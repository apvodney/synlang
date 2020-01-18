package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

var Log *log.Logger

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	Log = log.New(os.Stderr, "", 0)
}

func main() {

	intc := make(chan os.Signal, 1)
	signal.Notify(intc, os.Interrupt)
	go func() {
		_ = <-intc
		os.Exit(1)
	}()

	// Right now, we're just testing the pipes.
	// This test is from when there was a race bug.
	p := NewBufPipe(1024)
	s := p.NewSender()

	for i := 0; i < 10; i++ {
		r := p.NewRecver()
		go func() {
			for {
				r.Recv()
			}
		}()
	}

	var i int
	var v Sample = 13.123
	go func() {
		for i = 0; i < 20000000; i++ {
			s.Send(v)
		}
	}()
	for {
		time.Sleep(5 * time.Second)
		fmt.Println(i)
	}

	// 	var outf *os.File
	//	var err error
	//	if len(os.Args) >= 2 {
	//		switch f := os.Args[1]; f {
	//		case "-":
	//			outf = os.Stdout
	//		default:
	//			outf, err = os.OpenFile(f, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	//		}
	//	} else {
	//		outf, err = os.OpenFile("/dev/audio", os.O_WRONLY, 0755)
	//	}
	//	check(err)
	//
	//	_ = outf
	//	outp := New
	//	go out(Modin{"in": outc}, nil, outf, 1024) // Get bufsize from /dev/audioctl
	//
	//	add1 := make(chan modbuffer)
	//	add2 := make(chan modbuffer)
	//
	//	cr1 := make(chan modbuffer)
	//	cm := make(chan modbuffer)
	//	rm := make(chan modbuffer)
	//	go constval(Modio{"out": cr1}, Modparams{"val": "20"})
	//	go constval(Modio{"out": cm}, Modparams{"val": "1000"})
	//	go ramp(Modio{"freq": cr1, "out": rm}, nil)
	//	go mul(Modio{"in1": rm, "in2": cm, "out": add1}, nil)
	//
	//	cr := make(chan modbuffer)
	//	rs := make(chan modbuffer)
	//	go constval(Modio{"out": add2}, Modparams{"val": "440"})
	//	go add(Modio{"in1": add1, "in2": add2, "out": cr}, nil)
	//	go ramp(Modio{"freq": cr, "out": rs}, nil)
	//	go sinshp(Modio{"in": rs, "out": outc}, nil)

	//	select {}
}
