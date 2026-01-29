package util

import "sync/atomic"

var GlobalSeqNumber Uint32

var GlobalTeId Uint32

type Uint32 struct {
	val uint32
}

func (u *Uint32) Load() uint32 {
	return atomic.LoadUint32(&u.val)
}

func (u *Uint32) Inc() uint32 {
	return atomic.AddUint32(&u.val, 1)
}

func (u *Uint32) Dec() uint32 {
	return atomic.AddUint32(&u.val, ^uint32(0))
}

func (u *Uint32) Swap(newVal uint32) uint32 {
	return atomic.SwapUint32(&u.val, newVal)
}

func (u *Uint32) CompareAndSwap(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&u.val, old, new)
}
