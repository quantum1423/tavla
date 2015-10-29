package main

import (
	"fmt"
	"net"
	"os"
	"tavla/gopa"
	"tavla/opus"
	"tavla/trtp"
	"time"
	"unsafe"
)

func server() {
	dr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:33333")
	srv, err := net.ListenUDP("udp", dr)
	if err != nil {
		panic(err.Error())
	}
	buff := make([]byte, 512)
	for {
		num, addr, err := srv.ReadFromUDP(buff)
		if err != nil {
			panic(err.Error())
		}
		srv.WriteToUDP(buff[:num], addr)
	}
}

func client() {
	file, err := os.Open("song.wav")
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	gopa.Initialize()
	defer gopa.Release()
	outOpts := &gopa.OutputOptions{gopa.DefaultDevice, 2}
	spkr, err := gopa.OpenStream(nil, outOpts, 48000)
	if err != nil {
		panic(err.Error())
	}

	const RD_FACTOR = 2

	btrt := 48000 / RD_FACTOR

	const FRAME_SIZE = 20

	jbf := trtp.NewJitterBuffer()
	// decoder thread
	go func() {
		time.Sleep(time.Second)
		dec, err := opus.NewDecoder(48000, 2)
		if err != nil {
			panic(err.Error())
		}
		for {
			var x []byte
			stretchP := !jbf.IsAtLeast(4)
			seg, err := jbf.Pop()
			x = seg.Payload

			var toplay []float32
			nxt, err := jbf.Peek()
			if err == nil && x == nil {
				fmt.Printf("But don't worry! We have the next one for FEC!\n")
				toplay, err = dec.Decode(nxt.Payload, FRAME_SIZE*48, true)
			} else {
				toplay, err = dec.Decode(x, FRAME_SIZE*48, false)
			}
			if err != nil {
				panic(err.Error())
			}
			if !stretchP {
				err = spkr.WriteSound(toplay)
				if err != nil {
					fmt.Println(err.Error())
				}
			} else {
				longer := make([]float32, len(toplay)*2)
				for i := 0; i < len(longer); i++ {
					longer[i] = toplay[i%len(toplay)]
				}
				/*for i := 0; i < len(toplay); i++ {
					//longer[len(toplay)+i] = toplay[len(toplay)-1-i]
					longer[len(toplay)+i] = toplay[i]
				}*/
				err = spkr.WriteSound(longer)
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println("**************** PLAYED")
			}
		}
	}()
	dr, err := net.ListenUDP("udp4", nil)
	if err != nil {
		panic(err.Error())
	}
	go func() {
		// reader thread
		for {
			buff := make([]byte, 512)
			num, err := dr.Read(buff)
			if err != nil {
				panic(err.Error())
			}
			var seg trtp.Segment
			err = seg.FromBytes(buff[:num])
			if err != nil {
				panic(err.Error())
			}
			e := jbf.Push(seg, 16)
			if e != nil {
				fmt.Println(e)
			}
		}
	}()

	addr, _ := net.ResolveUDPAddr("udp4", "106.185.44.173:33333")

	// uploader thread
	enc, err := opus.NewEncoder(48000, 2, opus.OPUS_APPLICATION_VOIP)
	enc.SetExpectedPacketLoss(20)
	err = enc.SetBitrate(btrt)
	if err != nil {
		panic(err.Error())
	}
	nfo, _ := file.Stat()
	bts := make([]byte, FRAME_SIZE*96*2)
	i16s := (*[FRAME_SIZE * 96]int16)(unsafe.Pointer(&bts[0]))[:]
	ticker := time.NewTicker(time.Millisecond * (FRAME_SIZE))
	defer ticker.Stop()

	var xaxa uint8

	for i := 0; int64(i) < nfo.Size(); i += (FRAME_SIZE * 96 * 2) {
		// get bytes
		_, err := file.Read(bts)
		if err != nil {
			panic(err.Error())
		}
		// convert to float
		samples := make([]float32, FRAME_SIZE*96)
		for k, v := range i16s {
			samples[k] += float32(v) / 32767.0
		}
		// encode to opus
		frm, err := enc.Encode(samples, 240)
		if err != nil {
			panic(err.Error())
		}
		seg := trtp.Segment{xaxa, 0, frm}

		// extreme resilience mode
		for i := 0; i < RD_FACTOR; i++ {
			_, err = dr.WriteToUDP(seg.ToBytes(), addr)
			if err != nil {
				//panic(err.Error())
			}
		}

		// wait for ticker
		<-ticker.C
		enc.SetBitrate(btrt)
		if xaxa%10 == 0 {
			fmt.Printf("Real bitrate: %v bps\n", RD_FACTOR*(len(frm)+28+2+16)*8*1000/FRAME_SIZE)
		}
		xaxa++
	}
}

func main() {
	if os.Args[1] == "server" {
		server()
	} else {
		client()
	}
}
