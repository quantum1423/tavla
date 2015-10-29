package trtp

import (
	"bytes"
	"fmt"

	"github.com/lunixbochs/struc"
)

// Segment represents a TRTP packet on the wire.
type Segment struct {
	Sequence uint8
	Length   int `struc:"uint8,sizeof=Payload"`
	Payload  []byte
}

func (seg *Segment) ToBytes() []byte {
	buf := new(bytes.Buffer)
	err := struc.Pack(buf, seg)
	if err != nil {
		panic(err.Error())
	}
	return buf.Bytes()
}

func (seg Segment) String() string {
	return ""
}

func (seg *Segment) FromBytes(b []byte) error {
	return struc.Unpack(bytes.NewBuffer(b), seg)
}

func init() {
	test := Segment{123, 0, []byte("HELLO WORLD")}
	fmt.Println(test.ToBytes())
}
