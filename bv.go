package main

type BitVector struct {
	len  int
	bits []uint8
}

func NewBitVector(len int) *BitVector {
	bv := new(BitVector)
	bv.len = len
	if len%8 != 0 {
		len = len/8 + 1
	} else {
		len /= 8
	}
	bv.bits = make([]uint8, len)
	return bv
}

func (bv *BitVector) Set(i int) {
	if bv != nil && i < bv.len {
		bv.bits[i/8] |= 1 << (i % 8)
	}
}

func (bv *BitVector) Get(i int) bool {
	if bv == nil || i >= bv.len {
		return false
	}
	return bv.bits[i/8]&1<<(i%8) != 0
}

func (bv *BitVector) Clear(i int) {
	if bv != nil && i < bv.len {
		bv.bits[i/8] &= ^(1 << (i % 8))
	}
}

func (bv *BitVector) Toggle(i int) {
	if bv != nil && i < bv.len {
		bv.bits[i/8] ^= 1 << (i % 8)
	}
}

func (bv *BitVector) Len() int {
	return bv.len
}

func NewBitVectorFromBools(opt ...bool) *BitVector {
	bv := NewBitVector(len(opt))
	for i, v := range opt {
		if v {
			bv.Set(i)
		}
	}
	return bv
}
