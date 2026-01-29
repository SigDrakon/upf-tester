package dataplane

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Receiver 数据平面接收器
type Receiver struct {
	listenIP   string
	listenPort int
	teid       uint32
	ueIP       string
	
	conn     *net.UDPConn
	stopChan chan struct{}
	wg       sync.WaitGroup
	
	// 统计信息
	packetsReceived int
	bytesReceived   int64
	mu              sync.Mutex
}

// NewReceiver 创建新的接收器
func NewReceiver(listenIP string, listenPort int, teid uint32, ueIP string) *Receiver {
	return &Receiver{
		listenIP:   listenIP,
		listenPort: listenPort,
		teid:       teid,
		ueIP:       ueIP,
		stopChan:   make(chan struct{}),
	}
}

// Start 启动接收器
func (r *Receiver) Start() error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", r.listenIP, r.listenPort))
	if err != nil {
		return fmt.Errorf("resolve UDP address failed: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("listen UDP failed: %w", err)
	}

	r.conn = conn
	log.Printf("Receiver started on %s:%d, waiting for TEID=%d, UE IP=%s", r.listenIP, r.listenPort, r.teid, r.ueIP)

	r.wg.Add(1)
	go r.receive()

	return nil
}

// receive 接收数据包
func (r *Receiver) receive() {
	defer r.wg.Done()

	buffer := make([]byte, 65535)

	for {
		select {
		case <-r.stopChan:
			return
		default:
			// 设置读取超时，避免阻塞
			r.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			n, remoteAddr, err := r.conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("Read from UDP failed: %v", err)
				continue
			}

			// 解析 GTP 头部
			if n < 8 {
				log.Printf("Packet too short: %d bytes", n)
				continue
			}

			// GTP-U header: flags(1) + msgType(1) + length(2) + TEID(4)
			flags := buffer[0]
			msgType := buffer[1]
			teid := uint32(buffer[4])<<24 | uint32(buffer[5])<<16 | uint32(buffer[6])<<8 | uint32(buffer[7])

			// 检查是否是 GTP-U 数据包 (msgType = 0xFF)
			if msgType != 0xFF {
				continue
			}

			// 检查 TEID 是否匹配
			if teid != r.teid {
				continue
			}

			// 提取内部 IP 包
			gtpHeaderLen := 8
			if flags&0x07 != 0 { // 有扩展头部
				gtpHeaderLen = 12
			}

			if n <= gtpHeaderLen {
				continue
			}

			innerPacket := buffer[gtpHeaderLen:n]

			// 简单验证是否是 IP 包
			if len(innerPacket) < 20 {
				continue
			}

			// 更新统计
			r.mu.Lock()
			r.packetsReceived++
			r.bytesReceived += int64(n)
			r.mu.Unlock()

			if r.packetsReceived%10 == 0 {
				log.Printf("Received %d packets from %s, TEID=%d", r.packetsReceived, remoteAddr, teid)
			}
		}
	}
}

// Stop 停止接收器
func (r *Receiver) Stop() error {
	close(r.stopChan)
	if r.conn != nil {
		r.conn.Close()
	}
	r.wg.Wait()
	
	log.Printf("Receiver stopped. Total received: %d packets, %d bytes", r.packetsReceived, r.bytesReceived)
	return nil
}

// GetStats 获取统计信息
func (r *Receiver) GetStats() (packets int, bytes int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.packetsReceived, r.bytesReceived
}
