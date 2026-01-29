package util

import "sync/atomic"

var GlobalSeid Uint64

type Uint64 struct {
	val uint64
}

func (u *Uint64) Load() uint64 {
	return atomic.LoadUint64(&u.val)
}

func (u *Uint64) Inc() uint64 {
	return atomic.AddUint64(&u.val, 1)
}

func (u *Uint64) Dec() uint64 {
	return atomic.AddUint64(&u.val, ^uint64(0))
}

func (u *Uint64) Swap(newVal uint64) uint64 {
	return atomic.SwapUint64(&u.val, newVal)
}

func (u *Uint64) CompareAndSwap(old, new uint64) bool {
	return atomic.CompareAndSwapUint64(&u.val, old, new)
}
