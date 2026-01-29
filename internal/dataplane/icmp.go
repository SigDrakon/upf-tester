package dataplane

import (
	"os"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func buildICMPMessage(seq int, data []byte) ([]byte, error) {

	pid := os.Getpid() & 0xffff

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   pid,
			Seq:  seq,
			Data: data,
		},
	}

	return msg.Marshal(nil)
}
