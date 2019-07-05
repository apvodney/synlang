package main

import (
	"sync"
)

type BufPipe struct {
	ends     int           // total number of ends
	complock *sync.Mutex   // lock for complete
	complete int           // number of completed ends
	swaplock *sync.RWMutex // other half of lock given to swapcond
	swapcond *sync.Cond    // buffers have been swapped condition
	sendbuf  []Sample      // 
	recvbuf  []Sample      // 
}

type BPEnd struct {
	*BufPipe
	i int
}

func NewBufPipe(bufsiz int) *BufPipe {
	b := &BufPipe{
		complock: new(sync.Mutex),
		swaplock: new(sync.RWMutex),
		sendbuf:  make([]Sample, bufsiz),
		recvbuf:  make([]Sample, bufsiz),
	}
	b.swapcond = sync.NewCond(b.swaplock.RLocker())
	return b
}

func (b *BufPipe) newEnd() *BPEnd {
	b.ends++
	b.swapcond.L.Lock()
	return &BPEnd{
		BufPipe: b,
	}
}

func (b *BufPipe) NewSender() *BPEnd {
	return b.newEnd()
}

func (b *BufPipe) NewRecver() *BPEnd {
	return b.newEnd()
}

func (e *BPEnd) swap() {
	e.i = 0
	e.complock.Lock()
	e.complete++
	if e.ends == e.complete {
		go func() {
			e.recvbuf, e.sendbuf = e.sendbuf, e.recvbuf
			e.complete = 0
			
			e.swaplock.Lock()  // Make sure all ends are Wait-ing
			e.swapcond.Broadcast()
			e.swaplock.Unlock()
			
			return
		}()
	}
	e.complock.Unlock()
	
	e.swapcond.Wait()
}

func (e *BPEnd) Send(s Sample) {
	if e.i >= len(e.sendbuf) {
		e.swap()
	}
	e.sendbuf[e.i] = s
	e.i++
	return
}

func (e *BPEnd) Recv() (s Sample) {
	if e.i >= len(e.recvbuf) {
		e.swap()
	}
	s = e.recvbuf[e.i]
	e.i++
	return
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type UnbufPipe struct {
	recvers [](chan Sample)
}

type UBPEnd struct {
	*UnbufPipe
	id int      // index of chan to recv from
}

func NewUnbufPipe() *UnbufPipe {
	return new(UnbufPipe)
}

func (p *UnbufPipe) NewRecver() *UBPEnd {
	p.recvers = append(p.recvers, make(chan Sample, 1)) // I lied, there's a little buffer
	return &UBPEnd{
		UnbufPipe: p,
		id: len(p.recvers) - 1,
	}
}

func (p *UnbufPipe) NewSender() *UBPEnd {
	return &UBPEnd{
		UnbufPipe: p,
		id: -1,
	}
}

func (e *UBPEnd) Send(s Sample) {
	for _, c := range e.recvers {
		c <- s
	}
}

func (e *UBPEnd) Recv() Sample {
	return <- e.recvers[e.id]
}
