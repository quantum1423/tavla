package gopa

import (
	"fmt"
	"runtime"
	"testing"
)

/*
func TestLolWut(t *testing.T) {
	lolwut()
}
*/

func TestEcho(t *testing.T) {
	Initialize()
	defer Release()

	inOpts := &InputOptions{DefaultDevice, 2}
	outOpts := &OutputOptions{DefaultDevice, 2}
	mic, err := OpenStream(inOpts, nil, 48000)
	if err != nil {
		panic(err.Error())
	}
	defer mic.Close()
	spkr, err := OpenStream(nil, outOpts, 48000)
	defer spkr.Close()

	ch := make(chan []float32, 10)

	go func() {
		runtime.LockOSThread()
		buff := make([]float32, 960*2)
		for {
			select {
			case lol := <-ch:
				fmt.Printf("playback %v\n", len(ch))
				copy(buff, lol)
			default:
			}

			spkr.WriteSound(buff)
		}
	}()

	runtime.LockOSThread()
	for {
		my := make([]float32, 960*2)
		err = mic.ReadSound(my)
		if err != nil {
			panic(err.Error())
		}
		for i, _ := range my {
			my[i] *= 4.0
		}
		select {
		case ch <- my:
		default:
		}
	}
}
