package rings

import "sync/atomic"

func NewSequence() (seq *Sequence) {
	seq = &Sequence{
		value:   0,
		padding: [7]uint32{0, 1, 2, 3, 4, 5, 6},
	}
	return
}

type Sequence struct {
	value   uint32
	padding [7]uint32
}

func (seq *Sequence) Next() (n uint32) {
	n = atomic.AddUint32(&seq.value, 1) - 1
	return
}

func (seq *Sequence) Value() (n uint32) {
	n = atomic.LoadUint32(&seq.value) - 1
	return
}
