package trtp

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/KirisurfProject/kilog"
)

type JitterBuffer struct {
	cvar   *sync.Cond
	queue  map[uint64]Segment
	expect uint64
	topseq uint64
	hdLim  int
}

// IsAtLeast returns whether the jitter buffer has at least the given
// number of segments buffered. Calling this method always forces a context-switch,
// so it's safe/fair-ish to busy-wait on it. However, it's probably a better
// idea to use WaitAtLeast.
func (jb *JitterBuffer) IsAtLeast(num int) bool {
	jb.cvar.L.Lock()
	defer jb.cvar.L.Unlock()
	runtime.Gosched()
	return jb.topseq-jb.expect >= uint64(num)
}

// WaitAtLeast(num) is equivalent to for !IsAtLeast(num) {}, except it's much
// more efficient and doesn't spin around doing nothing.
func (jb *JitterBuffer) WaitAtLeast(num int) {
	jb.cvar.L.Lock()
	defer jb.cvar.L.Unlock()
	for jb.topseq-jb.expect < uint64(num) {
		jb.cvar.Wait()
	}
}

// Peek peeks at the possible next segment. It behaves like Dequeue,
// except it has no side effects.
func (jb *JitterBuffer) Peek() (seg Segment, err error) {
	jb.cvar.L.Lock()
	defer jb.cvar.L.Unlock()
	seg, ok := jb.queue[jb.expect]
	if !ok {
		err = errors.New("packet loss")
	}
	time.Sleep(0)
	return
}

// Pop dequeues a segment from the jitter buffer. An error
// will be returned if there is packet loss; it's a very bad idea to
// die or otherwise fail at such an error!
func (jb *JitterBuffer) Pop() (seg Segment, err error) {
	jb.cvar.L.Lock()
	defer jb.cvar.L.Unlock()
	defer jb.cvar.Broadcast()
	runtime.Gosched()

	seg, ok := jb.queue[jb.expect]
	// check if need to ketchup
	if len(jb.queue) >= (jb.hdLim*3)/4 {
		if ok {
			delete(jb.queue, jb.expect)
		}
		fmt.Println("Ketchup is so delicious!")
		jb.expect++ // catch up gracefully
		seg, ok = jb.queue[jb.expect]
	}

	if !ok {
		err = errors.New("packet loss")
		fmt.Printf("%v left: %v...%v\n", len(jb.queue), jb.expect, jb.queue)

		if rand.Int()%20 == 0 {
			// point expect at the lowest packet.
			jb.expect = ^uint64(0)
			for i, _ := range jb.queue {
				if i < jb.expect {
					jb.expect = i
				}
			}
		} else {
			jb.expect++ // USUALLY packet loss doesn't reset timing
			fmt.Println()
		}
		return
	}
	delete(jb.queue, jb.expect)
	jb.expect++ // we move ahead
	return
}

// Push enqueues a segment into the jitter buffer. If the buffer
// already contains `maxCount` or more segments, the new segment is
// dropped and an error is returned. It's a bad idea to die on an error
// returned from this function; it's usually safe to just ignore the
// return value.
func (jb *JitterBuffer) Push(seg Segment, maxCount int) (err error) {
	jb.cvar.L.Lock()
	defer jb.cvar.L.Unlock()
	defer jb.cvar.Broadcast()
	jb.hdLim = maxCount

	if len(jb.queue) >= maxCount {
		return errors.New("JitterBuffer overfull")
	}
	// calculate the correct sequence number
	expect8b := jb.expect % 256
	queuepos := uint64(seg.Sequence)

	if queuepos >= expect8b && queuepos-expect8b < 128 {
		// [  0   e   n         0]
		queuepos += jb.expect - expect8b
	} else if queuepos < expect8b && queuepos+256-expect8b < 128 {
		// [   e  0    n           0  ]
		queuepos += 256 + jb.expect - expect8b
	} else {
		// drop packet
		kilog.FineDebug("trtp: JitterBuffer dropped packet (ex=%v, e8=%v, qp=%v)",
			jb.expect, queuepos, queuepos)
		return
	}

	// return if not too late
	if queuepos >= jb.expect {
		// max seqnum
		if queuepos > jb.topseq {
			jb.topseq = queuepos
		}

		defer kilog.FineDebug("trtp: JitterBuffer successfully enqueued packet %v (%v)", queuepos, seg.Sequence)
		jb.queue[queuepos] = seg
	}
	return
}

func NewJitterBuffer() *JitterBuffer {
	return &JitterBuffer{
		sync.NewCond(new(sync.Mutex)),
		make(map[uint64]Segment),
		0,
		0,
		0,
	}
}
