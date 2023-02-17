package rings

import (
	"bytes"
	"fmt"
	"math"
	"unsafe"
)

func NewImmutable[E Entry](key string, values []E) (r *ImmutableRing[E]) {
	if values == nil || len(values) == 0 {
		panic(fmt.Errorf("new immutable ring failed, cause values is nil or empty"))
		return
	}
	keys := make(map[string]int)
	for _, value := range values {
		n := keys[value.Key()]
		n++
		keys[value.Key()] = n
	}
	err := ""
	for vk, n := range keys {
		if n > 1 {
			err = err + ", " + vk
		}
	}
	if err != "" {
		panic(fmt.Errorf("new immutable ring failed cause [%s] has duplicated", err[2:]))
		return
	}
	valuesLen := uint32(len(values))
	capacity := uint32(0)
	doubleTimes := uint32(1)
	for {
		capacity = uint32(math.Pow(float64(2), float64(doubleTimes)))
		if capacity >= valuesLen {
			break
		}
		doubleTimes++
	}
	filledValues := make([]E, capacity)
	for i := uint32(0); i < capacity; i++ {
		idx := i % valuesLen
		filledValues[i] = values[idx]
	}

	var elements []*element[E] = nil
	align := uint32(unsafe.Alignof(elements))
	mask := capacity - 1
	shift := uint32(math.Log2(float64(capacity)))
	elements = make([]*element[E], capacity*align)
	for i := range elements {
		elements[i] = &element[E]{}
	}

	elementBasePtr := uintptr(unsafe.Pointer(&elements[0]))
	elementMSize := unsafe.Sizeof(elements[0])

	r = &ImmutableRing[E]{
		key:            key,
		sequence:       NewSequence(),
		size:           capacity,
		shift:          shift,
		align:          align,
		mask:           mask,
		elements:       elements,
		elementBasePtr: elementBasePtr,
		elementMSize:   elementMSize,
	}

	for i := uint32(0); i < capacity; i++ {
		r.elementAt(i).value = filledValues[i]
	}

	return
}

type ImmutableRing[E Entry] struct {
	sequence       *Sequence
	elements       []*element[E]
	size           uint32
	shift          uint32
	align          uint32
	mask           uint32
	elementBasePtr uintptr
	elementMSize   uintptr
	key            string
}

func (r *ImmutableRing[E]) Key() (key string) {
	key = r.key
	return
}

func (r *ImmutableRing[E]) Next() (value E) {
	idx := r.sequence.Next() % r.size
	e := r.elementAt(idx)
	value = e.value
	return
}

func (r *ImmutableRing[E]) Head() (value E, has bool) {
	e := r.elementAt(r.sequence.Value())
	value = e.value
	has = true
	return
}

func (r *ImmutableRing[E]) Get(key string) (value E, has bool) {
	for i := uint32(0); i < r.size; i++ {
		e := r.elementAt(i)
		if e.value.Key() == key {
			value = e.value
			has = true
			break
		}
	}
	return
}

func (r *ImmutableRing[E]) String() (value string) {
	p := bytes.NewBufferString("")
	_ = p.WriteByte('[')
	for i := uint32(0); i < r.size; i++ {
		e := r.elementAt(i)
		if i == 0 {
			_, _ = p.WriteString(e.value.Key())
		} else {
			_, _ = p.WriteString(", ")
			_, _ = p.WriteString(e.value.Key())
		}
	}
	_ = p.WriteByte(']')
	value = p.String()
	return
}

func (r *ImmutableRing[E]) elementAt(idx uint32) (e *element[E]) {
	elementPtr := r.elementBasePtr + uintptr(idx&r.mask*r.align)*r.elementMSize
	e = *((*(*element[E]))(unsafe.Pointer(elementPtr)))
	return
}

type element[E Entry] struct {
	value E
}
