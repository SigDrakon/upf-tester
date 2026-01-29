package handler

import (
	"log"
	"sync"
	"upftester/internal/network"

	"github.com/wmnsk/go-pfcp/message"
)

type PFCPMessage struct {
	MessageType uint8
	Sequence    uint32
	SEID        uint64
	Payload     []byte
}

type PFCPDispatcher struct {
	transport *network.UDPTransport

	sessionMap sync.Map
	wg         sync.WaitGroup
	stopChan   chan struct{}
}

var (
	dispatcherInstance *PFCPDispatcher
	once               sync.Once
)

func GetPFCPDispatcher() *PFCPDispatcher {
	return dispatcherInstance
}

func NewPFCPDispatcher(t *network.UDPTransport) *PFCPDispatcher {
	once.Do(func() {
		dispatcherInstance = &PFCPDispatcher{
			transport: t,
			stopChan:  make(chan struct{}),
		}
	})
	return dispatcherInstance
}

func (d *PFCPDispatcher) Register(seid uint64, ch chan *PFCPMessage) {
	d.sessionMap.Store(seid, ch)
}

func (d *PFCPDispatcher) Unregister(seid uint64) {
	d.sessionMap.Delete(seid)
}

func (d *PFCPDispatcher) Start() {
	d.wg.Add(1)
	go d.run()
}

func (d *PFCPDispatcher) Stop() {
	close(d.stopChan)
	d.wg.Wait()
}

func (d *PFCPDispatcher) run() {
	defer d.wg.Done()
	for {
		select {
		case <-d.stopChan:
			return
		case pkt, ok := <-d.transport.Receive():
			if !ok {
				return
			}
			msg := d.decodePFCP(pkt.Data)
			if msg == nil {
				continue
			}
			d.dispatch(msg)
		}
	}
}

func (d *PFCPDispatcher) decodePFCP(data []byte) *PFCPMessage {

	header, err := message.ParseHeader(data)
	if err != nil {
		log.Println("PFCP 解析失败：", err)
		return nil
	}

	if header.Type == message.MsgTypeAssociationSetupResponse ||
		header.Type == message.MsgTypeHeartbeatRequest {
		return &PFCPMessage{
			MessageType: header.Type,
			Sequence:    header.SequenceNumber,
			SEID:        0, // maybe the default value is zero
			Payload:     data,
		}
	} else {
		return &PFCPMessage{
			MessageType: header.Type,
			Sequence:    header.SequenceNumber,
			SEID:        header.SEID,
			Payload:     data,
		}
	}
}

func (d *PFCPDispatcher) dispatch(msg *PFCPMessage) {

	ch, ok := d.sessionMap.Load(msg.SEID)
	if !ok {
		log.Printf("seid=%d not registered", msg.SEID)
		return
	}

	sessionCh, ok := ch.(chan *PFCPMessage)
	if !ok {
		log.Printf("SEID=%d 的 channel 不是一个 chan *PFCPMessage，丢弃消息", msg.SEID)
		return
	}

	select {
	case sessionCh <- msg:

	default:
		log.Printf("SEID=%d channel is full, drop message", msg.SEID)
	}

}
