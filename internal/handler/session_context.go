package handler

import (
	"sync"
)

// SessionState 会话状态
type SessionState int

const (
	SessionStateIdle SessionState = iota
	SessionStateEstablishing
	SessionStateActive
	SessionStateModifying
	SessionStateDeleting
	SessionStateDeleted
)

// SessionContext 会话上下文，保存会话相关信息
type SessionContext struct {
	// 信令面标识
	SEID uint64 // SMF 分配的 SEID
	UPFSEID uint64 // UPF 返回的 SEID

	// 数据面标识
	UplinkTEID   uint32 // 上行 TEID (N3 接口)
	UplinkPDRID  uint16 // 上行 PDR ID (用于查找 TEID)
	DownlinkTEID uint32 // 下行 TEID (N3 接口)
	UEIP         string // UE IP 地址

	// 会话状态
	State SessionState

	// 数据平面测试句柄
	DataPlaneTestHandle interface{}

	// 其他信息
	CreatedAt int64
	UpdatedAt int64
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[uint64]*SessionContext
	mu       sync.RWMutex
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[uint64]*SessionContext),
	}
}

// AddSession 添加会话
func (sm *SessionManager) AddSession(seid uint64, ctx *SessionContext) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[seid] = ctx
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(seid uint64) (*SessionContext, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	ctx, ok := sm.sessions[seid]
	return ctx, ok
}

// UpdateSession 更新会话
func (sm *SessionManager) UpdateSession(seid uint64, ctx *SessionContext) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[seid] = ctx
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(seid uint64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, seid)
}

// GetAllSessions 获取所有会话
func (sm *SessionManager) GetAllSessions() []*SessionContext {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	sessions := make([]*SessionContext, 0, len(sm.sessions))
	for _, ctx := range sm.sessions {
		sessions = append(sessions, ctx)
	}
	return sessions
}

// Count 获取会话数量
func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// 全局会话管理器
var GlobalSessionManager = NewSessionManager()
