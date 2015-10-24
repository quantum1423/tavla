package gopa

// #cgo LDFLAGS: -Wl,-Bstatic -lportaudio -Wl,-Bdynamic -lasound -lrt -lpthread -lm -ljack
// #include <portaudio.h>
import "C"
import (
	"math"
	"runtime"
	"unsafe"
)

func genSine(hz int) []float32 {
	toret := make([]float32, 4800)
	frq := float64(hz) / 48000.0
	for i := 0; i < 4800; i++ {
		toret[i] = float32(math.Sin(2.0 * math.Pi * frq * float64(i)))
	}
	return toret
}

func lolwut() {
	runtime.LockOSThread()
	var stream unsafe.Pointer
	err := C.Pa_OpenDefaultStream(&stream,
		0,           // no input
		1,           // mono
		C.paFloat32, // 32-bit floating point
		48000,
		C.paFramesPerBufferUnspecified,
		nil,
		nil)
	if err != C.paNoError {
		panic(C.GoString(C.Pa_GetErrorText(err)))
	}
	err = C.Pa_StartStream(stream)
	if err != C.paNoError {
		panic(C.GoString(C.Pa_GetErrorText(err)))
	}
	for {
		err := C.Pa_WriteStream(stream, unsafe.Pointer(&genSine(440)[0]), 48000)
		if err != C.paNoError {
			panic(C.GoString(C.Pa_GetErrorText(err)))
		}
	}
}

func init() {

}
