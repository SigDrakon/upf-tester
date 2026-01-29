package util

import "sync/atomic"

// GlobalTeid 全局 TEID 分配器
var GlobalTeid AtomicCounter32

// TEID 管理器，用于分配和回收 TEID
type TeidManager struct {
	current uint32
}

// NewTeidManager 创建新的 TEID 管理器
func NewTeidManager(start uint32) *TeidManager {
	return &TeidManager{
		current: start,
	}
}

// Allocate 分配一个新的 TEID
func (t *TeidManager) Allocate() uint32 {
	return atomic.AddUint32(&t.current, 1)
}

// AtomicCounter 原子计数器，用于 TEID 分配
type AtomicCounter32 struct {
	counter uint32
}

// Inc 递增并返回新值
func (a *AtomicCounter32) Inc() uint32 {
	return atomic.AddUint32(&a.counter, 1)
}

// Get 获取当前值
func (a *AtomicCounter32) Get() uint32 {
	return atomic.LoadUint32(&a.counter)
}

// Set 设置值
func (a *AtomicCounter32) Set(val uint32) {
	atomic.StoreUint32(&a.counter, val)
}
