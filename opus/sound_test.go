package opus

import (
	"fmt"
	"math/rand"
	"tavla/gopa"
	"testing"
	"time"
)

func TestSound(t *testing.T) {
	gopa.Initialize()
	defer gopa.Release()
	inOpts := &gopa.InputOptions{gopa.DefaultDevice, 2}
	outOpts := &gopa.OutputOptions{gopa.DefaultDevice, 2}
	// mic and speaker
	mic, err := gopa.OpenStream(inOpts, nil, 48000)
	if err != nil {
		panic(err.Error())
	}
	spkr, err := gopa.OpenStream(nil, outOpts, 48000)
	if err != nil {
		panic(err.Error())
	}

	ch := make(chan []byte, 5)
	// decoder thread
	go func() {
		time.Sleep(time.Second * 2)
		dec, err := NewDecoder(48000, 2)
		if err != nil {
			panic(err.Error())
		}
		for {
			var frm []byte
			select {
			case x := <-ch:
				frm = x
			default:
				fmt.Println("Loss!")
			}
			toplay, err := dec.Decode(frm, 960)
			if err != nil {
				panic(err.Error())
			}
			spkr = spkr
			toplay = toplay
			err = spkr.WriteSound(toplay)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}()

	// encoder thread
	enc, err := NewEncoder(48000, 2, OPUS_APPLICATION_VOIP)
	enc.SetExpectedPacketLoss(20)
	err = enc.SetBitrate(32000)
	if err != nil {
		panic(err.Error())
	}
	buff := make([]float32, 960*2)
	for {
		err := mic.ReadSound(buff)
		if err != nil {
			panic(err.Error())
		}

		for i, _ := range buff {
			buff[i] *= 12.0
		}

		frm, err := enc.Encode(buff, 512)
		if err != nil {
			panic(err.Error())
		}
		// horrible packet loss
		if rand.Int()%100 >= 7 {
			select {
			case ch <- frm:
			default:
			}
		} else {
			select {
			case ch <- nil:
			default:
			}
		}
	}
}
