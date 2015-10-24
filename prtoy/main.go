package main

import (
	"fmt"
	"net"
	"os"
	"tavla/gopa"
	"tavla/opus"
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
	ch := make(chan []byte, 5)
	// decoder thread
	go func() {
		dec, err := opus.NewDecoder(48000, 2)
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
			select {
			case ch <- buff[:num]:
			default:
			}
		}
	}()

	addr, _ := net.ResolveUDPAddr("udp4", "oyashio.ithisa.net:33333")

	// uploader thread
	enc, err := opus.NewEncoder(48000, 2, opus.OPUS_APPLICATION_AUDIO)
	enc.SetExpectedPacketLoss(10)
	err = enc.SetBitrate(64000)
	if err != nil {
		panic(err.Error())
	}
	nfo, _ := file.Stat()
	bts := make([]byte, 1920*2)
	i16s := (*[1920]int16)(unsafe.Pointer(&bts[0]))[:]
	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()
	for i := 0; int64(i) < nfo.Size(); i += (1920 * 2) {
		fmt.Println(i)
		// get bytes
		_, err := file.Read(bts)
		if err != nil {
			panic(err.Error())
		}
		// convert to float
		samples := make([]float32, 1920)
		for k, v := range i16s {
			samples[k] = float32(v) / 32767.0
		}
		// encode to opus
		frm, err := enc.Encode(samples, 512)
		if err != nil {
			panic(err.Error())
		}
		_, err = dr.WriteToUDP(frm, addr)
		if err != nil {
			panic(err.Error())
		}
		// wait for ticker
		<-ticker.C
	}
}

func main() {
	if os.Args[1] == "server" {
		server()
	} else {
		client()
	}
}
