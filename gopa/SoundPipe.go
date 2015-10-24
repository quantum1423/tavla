package gopa

import "sync"

type SoundPipe struct {
	bfr []float32
	*sync.Cond
}

func (lol *SoundPipe) Read(p []float32) (int, error) {
	// wait until bfr is not empty
	lol.L.Lock()
	defer lol.L.Unlock()
	for len(lol.bfr) < len(p) {
		lol.Wait()
	}
	// now there is something
	n := copy(p, lol.bfr)
	lol.bfr = lol.bfr[n:]
	return n, nil
}

func (lol *SoundPipe) Write(p []float32) (int, error) {
	lol.L.Lock()
	defer lol.L.Unlock()
	xaxa := make([]float32, len(p))
	copy(xaxa, p)
	lol.bfr = append(lol.bfr, xaxa...)
	lol.Broadcast()
	return len(p), nil
}

func (lol *SoundPipe) Close() error { return nil } // TODO make this not noop

func NewSoundPipe() *SoundPipe {
	return &SoundPipe{nil, sync.NewCond(new(sync.Mutex))}
}
