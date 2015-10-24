package gopa

// #cgo LDFLAGS: -Wl,-Bstatic -lportaudio -Wl,-Bdynamic -lasound -lrt -lpthread -lm -ljack
// #include <portaudio.h>
import "C"

func Initialize() {
	err := C.Pa_Initialize()
	if err != C.paNoError {
		panic(C.Pa_GetErrorText(err))
	}
}

func Release() {
	C.Pa_Terminate()
}
