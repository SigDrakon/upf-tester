package network

import (
	"log"
	"net"
	"sync"
)

type Packet struct {
	Data []byte
	Addr *net.UDPAddr
}

type UDPTransport struct {
	conn        *net.UDPConn
	receiveChan chan *Packet
	sendChan    chan *Packet
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func NewUDPTransport(host, port string, queueSize uint16) (*UDPTransport, error) {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &UDPTransport{
		conn:        conn,
		receiveChan: make(chan *Packet, queueSize),
		sendChan:    make(chan *Packet, queueSize),
		stopChan:    make(chan struct{}),
	}, nil
}

func (t *UDPTransport) Send(data []byte, remoteAddr *net.UDPAddr) bool {

	packet := &Packet{
		Data: data,
		Addr: remoteAddr,
	}

	select {
	case t.sendChan <- packet:
		return true
	default:
		log.Println("发送队列已满，丢弃数据包。")
		return false
	}
}

func (t *UDPTransport) Receive() <-chan *Packet {
	return t.receiveChan
}

func (t *UDPTransport) runReceiver() {
	defer t.wg.Done()
	buffer := make([]byte, 1500)

	for {
		select {
		case <-t.stopChan:
			return
		default:
			n, addr, err := t.conn.ReadFromUDP(buffer)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && (opErr.Timeout() || opErr.Op == "read") {
					continue
				}
				if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
					return
				}

				return
			}

			packet := &Packet{
				Data: make([]byte, n),
				Addr: addr,
			}
			copy(packet.Data, buffer[:n])

			select {
			case t.receiveChan <- packet:

			default:
				//log.Println("接收队列已满，丢弃数据包。")
			}
		}
	}
}

func (t *UDPTransport) runSender() {
	defer t.wg.Done()
	for {
		select {
		case <-t.stopChan:
			return
		case packet := <-t.sendChan:
			_, err := t.conn.WriteToUDP(packet.Data, packet.Addr)
			if err != nil {
				//log.Printf("写入 UDP 失败: %v", err)
			} else {

			}
		}
	}
}

func (t *UDPTransport) Start() {
	t.wg.Add(2)
	go t.runReceiver()
	go t.runSender()
}

func (t *UDPTransport) Stop() {
	close(t.stopChan)
	t.conn.Close()
	t.wg.Wait()
	close(t.receiveChan)
	close(t.sendChan)
}
