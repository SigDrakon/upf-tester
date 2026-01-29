package encoding

import "github.com/wmnsk/go-pfcp/message"

type MessageConfig interface {
	Marshal(path string) (message.Message, error)
	Unmarshal()
}
