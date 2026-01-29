package dataplane

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

func checksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(data[i:]))
	}

	if len(data)%2 != 0 {
		sum += uint32(data[len(data)-1]) << 8
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += sum >> 16
	return ^uint16(sum)
}

func buildIPv4Header(srcIP, dstIP net.IP, payloadLen int) ([]byte, error) {
	if srcIP.To4() == nil || dstIP.To4() == nil {
		return nil, fmt.Errorf("仅支持IPv4地址")
	}

	ipHeader := make([]byte, 20)

	ipHeader[0] = 0x45

	ipHeader[1] = 0

	totalLen := uint16(20 + payloadLen)
	binary.BigEndian.PutUint16(ipHeader[2:4], totalLen)

	ipID := uint16(os.Getpid() & 0xffff)
	binary.BigEndian.PutUint16(ipHeader[4:6], ipID)

	ipHeader[6] = 0x40
	ipHeader[7] = 0

	ipHeader[8] = 64

	ipHeader[9] = 1

	ipHeader[10] = 0
	ipHeader[11] = 0

	copy(ipHeader[12:16], srcIP.To4())

	copy(ipHeader[16:20], dstIP.To4())

	cksum := checksum(ipHeader)
	binary.BigEndian.PutUint16(ipHeader[10:12], cksum)

	return ipHeader, nil
}
