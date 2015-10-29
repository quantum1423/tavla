package opus

/*
#cgo LDFLAGS: -Bstatic -lopus
#include <opus/opus.h>
*/
import "C"
import "fmt"

type Decoder struct {
	cee   C.OpusDecoder
	pad   [65536]byte // just for safety
	chcnt int
}

// NewDecoder creates a new Opus decoder with the given sample rate
// and number of channels.
func NewDecoder(sampleRate int, chanCount int) (dec *Decoder, err error) {
	dec = new(Decoder)

	// allocate Opus decoder
	ecode := C.opus_decoder_init(&dec.cee, C.opus_int32(sampleRate),
		C.int(chanCount))
	if ecode != C.OPUS_OK {
		dec = nil
		err = ErrUnspecified
	}
	dec.chcnt = chanCount
	return
}

// Decode decodes a single Opus frame to floating point format. Stereo
// output will be interleaved. A null input frame indicates packet loss.
func (dec *Decoder) Decode(frame []byte, pksize int, isFec bool) (pcm []float32, err error) {
	var input *C.uchar
	if frame != nil {
		input = (*C.uchar)(&frame[0])
	}

	isfNum := 0
	if isFec {
		isfNum = 1
	}

	pcm = make([]float32, pksize*dec.chcnt)
	num := C.opus_decode_float(&dec.cee, input, C.opus_int32(len(frame)),
		(*C.float)(&pcm[0]), C.int(pksize), C.int(isfNum))
	if num < 0 {
		pcm = nil
		err = ErrUnspecified
		switch num {
		case C.OPUS_BAD_ARG:
			fmt.Println("OPUS_BAD_ARG")
		case C.OPUS_BUFFER_TOO_SMALL:
			fmt.Println("OPUS_BUFFER_TOO_SMALL")
		case C.OPUS_INVALID_PACKET:
			fmt.Println("OPUS_INVALID_PACKET")
		}
		return
	}
	pcm = pcm[:num*C.int(dec.chcnt)]
	return
}
