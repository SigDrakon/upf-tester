package handler

import (
	"log"
	"net"
	"os"
	"time"
	"upftester/internal/network"
	"upftester/internal/util"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

var StartTime time.Time

func SendPFCPAssociationRequest(localN4Ip, upfN4Ip string, udpTransport *network.UDPTransport) error {

	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(upfN4Ip, "8805"))
	if err != nil {
		log.Println("解析 UDP 地址失敗:", err)
		return err
	}

	StartTime = time.Now()

	msg := message.NewAssociationSetupRequest(util.GlobalSeqNumber.Inc(),
		ie.NewNodeID(localN4Ip, "", ""),
		ie.NewRecoveryTimeStamp(StartTime),
	)

	data, err := msg.Marshal()
	if err != nil {
		log.Println("消息編碼失敗:", err)
		return err
	}

	ch := make(chan *PFCPMessage)
	GetPFCPDispatcher().Register(0, ch)
	go func() {
		for {
			select {
			case msg := <-ch:
				if msg.MessageType == message.MsgTypeAssociationSetupResponse {
					resp, err := message.ParseAssociationSetupResponse(msg.Payload)
					if err != nil {
						log.Println("消息解析失敗:", err)
						os.Exit(1)
					}

					value, err := resp.Cause.ValueAsUint8()
					if err != nil {
						log.Println("消息解析失敗:", err)
						os.Exit(1)
					}

					if value != ie.CauseRequestAccepted {
						log.Println("association create success")
						os.Exit(1)
					}
				}

				if msg.MessageType == message.MsgTypeHeartbeatRequest {
					msg := message.NewHeartbeatResponse(msg.Sequence,
						ie.NewRecoveryTimeStamp(StartTime),
					)
					data, err := msg.Marshal()
					if err != nil {
						log.Println("消息編碼失敗:", err)
						return
					}
					udpTransport.Send(data, addr)
				}
			}
		}
	}()

	udpTransport.Send(data, addr)
	return nil
}
