package opus

/*
#cgo LDFLAGS: -Bstatic -lopus
#include <opus/opus.h>

static int _opusSetBitrate(OpusEncoder *st, opus_int32 x) {
	return opus_encoder_ctl(st, OPUS_SET_BITRATE(x));
}

static int _opusSetLoss(OpusEncoder *st, opus_int32 x) {
	return opus_encoder_ctl(st, OPUS_SET_PACKET_LOSS_PERC(x));
}

static int _opusSetInbandFec(OpusEncoder *st, opus_int32 x) {
	return opus_encoder_ctl(st, OPUS_SET_INBAND_FEC(x));
}
*/
import "C"
import "errors"

type Encoder struct {
	cee   C.OpusEncoder
	pad   [65536]byte // just for safety
	chcnt int
}

var (
	OPUS_APPLICATION_VOIP  = C.OPUS_APPLICATION_VOIP
	OPUS_APPLICATION_AUDIO = C.OPUS_APPLICATION_AUDIO
)

var ErrUnspecified = errors.New("Unspecified error")

// NewEncoder creates a new Opus encoder with the given sample rate,
// channel count, and application type.
func NewEncoder(sampleRate int,
	chanCount int,
	application int) (enc *Encoder, err error) {
	enc = new(Encoder)

	// allocate Opus encoder
	ecode := C.opus_encoder_init(&enc.cee, C.opus_int32(sampleRate),
		C.int(chanCount), C.int(application))
	if ecode != C.OPUS_OK {
		err = ErrUnspecified
		return
	}
	enc.chcnt = chanCount
	return
}

// Encode encodes an Opus frame from floating-point input, limited
// to the maximum frame size (in bytes) given by outLimit. Stereo
// input should be interleaved.
func (enc *Encoder) Encode(pcm []float32, outLimit int) (frame []byte, err error) {
	frame = make([]byte, outLimit)
	ecode := C.opus_encode_float(&enc.cee, (*C.float)(&pcm[0]),
		C.int(len(pcm)/enc.chcnt), (*C.uchar)(&frame[0]), C.opus_int32(outLimit))
	if ecode < 0 {
		frame = nil
		err = ErrUnspecified
		return
	}
	frame = frame[:ecode]
	return
}

/* CTL function wrappers */

func (enc *Encoder) SetBitrate(bps int) (err error) {
	ecode := C._opusSetBitrate(&enc.cee, C.opus_int32(bps))
	if ecode != C.OPUS_OK {
		err = ErrUnspecified
	}
	return
}

func (enc *Encoder) SetExpectedPacketLoss(percent int) (err error) {
	ecode := C._opusSetLoss(&enc.cee, C.opus_int32(percent))
	if ecode != C.OPUS_OK {
		err = ErrUnspecified
	}
	if percent != 0 {
		C._opusSetInbandFec(&enc.cee, 1)
	}
	return
}
