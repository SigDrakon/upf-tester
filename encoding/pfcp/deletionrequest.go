package pfcp

import (
	"log"
	"os"
	"upftester/internal/util"

	"github.com/wmnsk/go-pfcp/message"
	"gopkg.in/yaml.v3"
)

type DeletionRequestConfig struct {
	// Session Deletion Request 通常不需要额外的 IE
	// 只需要在 Header 中携带 SEID
}

func (cfg *DeletionRequestConfig) Marshal(path string) (message.Message, error) {
	// 如果提供了配置文件，尝试加载
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Warning: could not read deletion config file %s: %v, using defaults", path, err)
		} else {
			err = yaml.Unmarshal(data, cfg)
			if err != nil {
				log.Printf("Warning: could not unmarshal deletion config: %v, using defaults", err)
			}
		}
	}

	// Session Deletion Request 通常不包含任何 IE
	// SEID 会在发送时由 handler 设置到 Header 中
	return message.NewSessionDeletionRequest(0, 0, 0, util.GlobalSeqNumber.Inc(), 0), nil
}

func (cfg *DeletionRequestConfig) Unmarshal() {
	// No specific unmarshal logic needed for deletion request
}
