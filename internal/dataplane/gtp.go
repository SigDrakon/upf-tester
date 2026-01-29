package dataplane

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func buildGTPHeader(tunnelID uint32, payloadLen int) ([]byte, error) {
	var gtpBuf bytes.Buffer

	flags := uint8(0x30)
	if err := binary.Write(&gtpBuf, binary.BigEndian, flags); err != nil {
		return nil, fmt.Errorf("write gtp flag failed: %v", err)
	}

	msgType := uint8(0xFF)
	if err := binary.Write(&gtpBuf, binary.BigEndian, msgType); err != nil {
		return nil, fmt.Errorf("write gtp message type failed: %v", err)
	}

	payloadLen16 := uint16(payloadLen)
	if err := binary.Write(&gtpBuf, binary.BigEndian, payloadLen16); err != nil {
		return nil, fmt.Errorf("write gtp payload length failed: %v", err)
	}

	if err := binary.Write(&gtpBuf, binary.BigEndian, tunnelID); err != nil {
		return nil, fmt.Errorf("write gtp tunnel id failed: %v", err)
	}

	return gtpBuf.Bytes(), nil
}
