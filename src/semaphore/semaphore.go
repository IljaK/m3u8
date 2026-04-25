package semaphore

import (
	"sync"
	"time"
)

type Counter struct {
	mx              sync.Mutex
	processingCount int
	accessLimit     int
}

func (p *Counter) StartNext() bool {
	if !p.CanStartNext() {
		return false
	}
	p.mx.Lock()
	defer p.mx.Unlock()
	p.processingCount++
	return true
}

func (p *Counter) CanStartNext() bool {
	p.mx.Lock()
	defer p.mx.Unlock()
	return p.processingCount < p.accessLimit
}

func (p *Counter) WaitAvailable() {
	for true {
		if p.StartNext() {
			break
		}
		time.Sleep(time.Millisecond)
	}
}

func (p *Counter) Complete() {
	p.mx.Lock()
	defer p.mx.Unlock()
	if p.processingCount > 0 {
		p.processingCount--
	}
}

func CreateSemaphore(accessLimit int) *Counter {
	return &Counter{
		mx:              sync.Mutex{},
		processingCount: 0,
		accessLimit:     accessLimit,
	}
}
