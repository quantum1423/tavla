package gopa

// #cgo LDFLAGS: -ljack -lm -lportaudio --lasound -lpthread
// #include <portaudio.h>
import "C"
import (
	"errors"
	"unsafe"
)

const DefaultDevice = 0

type InputOptions struct {
	DeviceID   int
	ChannelCnt int
}

type OutputOptions struct {
	DeviceID   int
	ChannelCnt int
}

type Stream struct {
	cee     unsafe.Pointer
	inopts  *InputOptions
	outopts *OutputOptions
}

func (strm *Stream) WriteSound(sound []float32) error {
	cerr := C.Pa_WriteStream(strm.cee, unsafe.Pointer(&sound[0]),
		C.ulong(len(sound)/strm.outopts.ChannelCnt))
	if cerr != C.paNoError {
		return errors.New(C.GoString(C.Pa_GetErrorText(cerr)))
	}
	return nil
}

func (strm *Stream) ReadSound(sound []float32) error {
	cerr := C.Pa_ReadStream(strm.cee, unsafe.Pointer(&sound[0]),
		C.ulong(len(sound)/strm.inopts.ChannelCnt))
	if cerr != C.paNoError {
		return errors.New(C.GoString(C.Pa_GetErrorText(cerr)))
	}
	return nil
}

func (strm *Stream) Close() {
	C.Pa_StopStream(strm.cee)
	C.Pa_CloseStream(strm.cee)
	strm.cee = nil // we really don't want a wild pointer around
}

func OpenStream(inOpts *InputOptions,
	outOpts *OutputOptions,
	sampleRate float64) (strm *Stream, err error) {
	// initialize return value first
	strm = new(Stream)
	strm.inopts = inOpts
	strm.outopts = outOpts
	// initialize in and out
	var _inopts *C.PaStreamParameters
	var _outopts *C.PaStreamParameters
	if inOpts != nil {
		_inopts = new(C.PaStreamParameters)
		_inopts.device = C.Pa_GetDefaultInputDevice()
		if inOpts.DeviceID != DefaultDevice {
			_inopts.device = C.PaDeviceIndex(inOpts.DeviceID)
		}
		_inopts.channelCount = C.int(inOpts.ChannelCnt)
		_inopts.sampleFormat = C.paFloat32
		_inopts.suggestedLatency = C.Pa_GetDeviceInfo(_inopts.device).defaultHighInputLatency
		_inopts.hostApiSpecificStreamInfo = nil
	}
	if outOpts != nil {
		_outopts = new(C.PaStreamParameters)
		_outopts.device = C.Pa_GetDefaultOutputDevice()
		if outOpts.DeviceID != DefaultDevice {
			_outopts.device = C.PaDeviceIndex(outOpts.DeviceID)
		}
		_outopts.channelCount = C.int(outOpts.ChannelCnt)
		_outopts.sampleFormat = C.paFloat32
		_outopts.suggestedLatency = C.Pa_GetDeviceInfo(_outopts.device).defaultHighOutputLatency
		_outopts.hostApiSpecificStreamInfo = nil
	}
	cerr := C.Pa_OpenStream(
		&strm.cee,
		_inopts,
		_outopts,
		C.double(sampleRate),
		C.paClipOff,
		0,
		nil,
		nil,
	)
	if cerr != C.paNoError {
		err = errors.New(C.GoString(C.Pa_GetErrorText(cerr)))
		return
	}
	cerr = C.Pa_StartStream(strm.cee)
	if cerr != C.paNoError {
		err = errors.New(C.GoString(C.Pa_GetErrorText(cerr)))
		return
	}
	return
}
