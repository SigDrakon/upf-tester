package dataplane

import (
	"fmt"
	"net"
)

// func UplinkIcmpSender(teid uint32, ueIpAddr *common.UEAddress) {

// srcAddr := config.GetGnbIp() + ":2152"
// dstAddr := config.GetUpfN3Ip() + ":2152"

// icmpSeq := 1
// icmpData := []byte("hello world")

// for {
// 	packet, err := BuildGTPIPICMPPacket(ueIpAddr.Ipv4Address, config.GetUpfDnIp(), teid, icmpSeq, icmpData)
// 	if err != nil {
// 		log.Println("构造GTP+IP+ICMP数据包失败:", err)
// 		return
// 	}

// 	err = SendUDPWithSrc(srcAddr, dstAddr, packet)
// 	if err != nil {
// 		log.Println("发送UDP数据包失败:", err)
// 		return
// 	}

// 	icmpSeq++

// 	time.Sleep(1 * time.Second)
// }
// }

func SendUDPWithSrc(srcAddrStr, dstAddrStr string, data []byte) error {

	srcAddr, err := net.ResolveUDPAddr("udp", srcAddrStr)
	if err != nil {
		return fmt.Errorf("解析本地源地址失败: %v（格式需为 IP:Port，如 192.168.1.10:8888）", err)
	}

	dstAddr, err := net.ResolveUDPAddr("udp", dstAddrStr)
	if err != nil {
		return fmt.Errorf("解析远程目标地址失败: %v（格式需为 IP:Port，如 10.0.0.1:2152）", err)
	}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		return fmt.Errorf("创建UDP连接失败: %v（可能原因：源端口被占用、无权限使用该源IP）", err)
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("发送UDP数据失败: %v", err)
	}

	return nil
}

func BuildIPICMPPacket(srcIP, dstIP string, seq int, icmpData []byte) ([]byte, error) {

	src := net.ParseIP(srcIP)
	dst := net.ParseIP(dstIP)
	if src == nil || dst == nil {
		return nil, fmt.Errorf("无效的IP地址")
	}

	icmpPacket, err := buildICMPMessage(seq, icmpData)
	if err != nil {
		return nil, fmt.Errorf("构造ICMP消息失败: %v", err)
	}

	ipHeader, err := buildIPv4Header(src, dst, len(icmpPacket))
	if err != nil {
		return nil, fmt.Errorf("构造IP头部失败: %v", err)
	}

	fullPacket := append(ipHeader, icmpPacket...)

	return fullPacket, nil
}

func BuildGTPIPICMPPacket(srcIP, dstIP string, tunnelID uint32, seq int, icmpData []byte) ([]byte, error) {

	ipIcmpPacket, err := BuildIPICMPPacket(srcIP, dstIP, seq, icmpData)
	if err != nil {
		return nil, fmt.Errorf("构造IP+ICMP数据包失败: %v", err)
	}

	gtpHeader, err := buildGTPHeader(tunnelID, len(ipIcmpPacket))
	if err != nil {
		return nil, fmt.Errorf("构造GTP头部失败: %v", err)
	}

	gtpFullPacket := append(gtpHeader, ipIcmpPacket...)

	return gtpFullPacket, nil
}

func SendUDS(socketPath string, data []byte) error {
	conn, err := net.Dial("unixgram", socketPath)
	if err != nil {
		return fmt.Errorf("connect directly to unix domain socket failed: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("write to unix domain socket failed: %v", err)
	}

	return nil
}
