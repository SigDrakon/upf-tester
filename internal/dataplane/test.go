package dataplane

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// DataPlaneTestConfig 数据平面测试配置
type DataPlaneTestConfig struct {
	TestType      string `yaml:"testType"`      // icmp, udp, tcp, throughput
	Duration      int    `yaml:"duration"`      // 测试时长（秒）
	PacketCount   int    `yaml:"packetCount"`   // 发送包数量，0 表示持续发送
	Interval      int    `yaml:"interval"`      // 发送间隔（毫秒）
	PayloadSize   int    `yaml:"payloadSize"`   // 负载大小（字节）
	Bidirectional bool   `yaml:"bidirectional"` // 是否双向测试
	DstIp         string `yaml:"dstIp"`         // 目标 IP 地址 (可选，默认使用 globalConfig.DnIp)
	UDSSocketPath string `yaml:"udsSocketPath"` // Unix Domain Socket 路径 (可选，用于替代 UDP发送)
}

// LoadDataPlaneTestConfig 从文件加载数据平面测试配置
func LoadDataPlaneTestConfig(path string) (*DataPlaneTestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file failed: %w", err)
	}

	var config DataPlaneTestConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	// 设置默认值
	if config.Duration == 0 {
		config.Duration = 10
	}
	if config.Interval == 0 {
		config.Interval = 1000 // 默认 1 秒
	}
	if config.PayloadSize == 0 {
		config.PayloadSize = 64
	}

	return &config, nil
}

// DataPlaneTestResult 数据平面测试结果
type DataPlaneTestResult struct {
	TestType        string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	PacketsSent     int
	PacketsReceived int
	PacketsLost     int
	PacketLossRate  float64
	AvgLatency      time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	Throughput      float64 // Mbps
	Success         bool
	ErrorMessage    string
}

// DataPlaneTest 数据平面测试接口
type DataPlaneTest interface {
	Start() error
	Stop() error
	GetResult() *DataPlaneTestResult
}

// ICMPTest ICMP 测试
type ICMPTest struct {
	config   *DataPlaneTestConfig
	srcIP    string
	dstIP    string
	gnbIP    string
	upfN3IP  string
	teid     uint32
	ueIP     string
	result   *DataPlaneTestResult
	stopChan chan struct{}
	doneChan chan struct{}
}

// NewICMPTest 创建 ICMP 测试
func NewICMPTest(config *DataPlaneTestConfig, gnbIP, upfN3IP string, teid uint32, ueIP, dstIP string) *ICMPTest {
	return &ICMPTest{
		config:   config,
		gnbIP:    gnbIP,
		upfN3IP:  upfN3IP,
		teid:     teid,
		ueIP:     ueIP,
		dstIP:    dstIP,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		result: &DataPlaneTestResult{
			TestType:  "ICMP",
			StartTime: time.Now(),
		},
	}
}

// Start 启动 ICMP 测试
func (t *ICMPTest) Start() error {
	log.Printf("Starting ICMP test: UE IP=%s, TEID=%d, Duration=%ds", t.ueIP, t.teid, t.config.Duration)

	go t.run()

	return nil
}

// run 运行 ICMP 测试
func (t *ICMPTest) run() {
	defer close(t.doneChan)

	srcAddr := t.gnbIP + ":2152"
	dstAddr := t.upfN3IP + ":2152"

	seq := 1
	icmpData := []byte("upf-tester-icmp-payload")

	ticker := time.NewTicker(time.Duration(t.config.Interval) * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(time.Duration(t.config.Duration) * time.Second)

	for {
		select {
		case <-t.stopChan:
			log.Println("ICMP test stopped by user")
			t.result.EndTime = time.Now()
			t.result.Duration = t.result.EndTime.Sub(t.result.StartTime)
			t.calculateResult()
			return

		case <-timeout:
			log.Println("ICMP test completed (timeout)")
			t.result.EndTime = time.Now()
			t.result.Duration = t.result.EndTime.Sub(t.result.StartTime)
			t.calculateResult()
			return

		case <-ticker.C:
			if t.config.PacketCount > 0 && t.result.PacketsSent >= t.config.PacketCount {
				log.Println("ICMP test completed (packet count reached)")
				t.result.EndTime = time.Now()
				t.result.Duration = t.result.EndTime.Sub(t.result.StartTime)
				t.calculateResult()
				return
			}

			// 构造 GTP+IP+ICMP 数据包
			packet, err := BuildGTPIPICMPPacket(t.ueIP, t.dstIP, t.teid, seq, icmpData)
			if err != nil {
				log.Printf("Build GTP+IP+ICMP packet failed: %v", err)
				continue
			}

			log.Printf("entry")
			// 发送数据包
			if t.config.UDSSocketPath != "" {
				log.Printf("Sending UDS packet: %v", packet)
				err = SendUDS(t.config.UDSSocketPath, packet)
				if err != nil {
					log.Printf("Send UDS packet failed: %v", err)
					continue
				}
			} else {
				err = SendUDPWithSrc(srcAddr, dstAddr, packet)
				if err != nil {
					log.Printf("Send UDP packet failed: %v", err)
					continue
				}
			}

			t.result.PacketsSent++
			seq++

			if t.result.PacketsSent%10 == 0 {
				log.Printf("ICMP test progress: sent %d packets", t.result.PacketsSent)
			}
		}
	}
}

// Stop 停止 ICMP 测试
func (t *ICMPTest) Stop() error {
	close(t.stopChan)
	<-t.doneChan
	return nil
}

// GetResult 获取测试结果
func (t *ICMPTest) GetResult() *DataPlaneTestResult {
	return t.result
}

// calculateResult 计算测试结果
func (t *ICMPTest) calculateResult() {
	// TODO: 实现接收端统计，目前只统计发送
	t.result.PacketsReceived = 0 // 需要接收端实现
	t.result.PacketsLost = t.result.PacketsSent - t.result.PacketsReceived

	if t.result.PacketsSent > 0 {
		t.result.PacketLossRate = float64(t.result.PacketsLost) / float64(t.result.PacketsSent) * 100
	}

	t.result.Success = t.result.PacketsSent > 0

	log.Printf("ICMP Test Result: Sent=%d, Received=%d, Lost=%d, Loss Rate=%.2f%%",
		t.result.PacketsSent, t.result.PacketsReceived, t.result.PacketsLost, t.result.PacketLossRate)
}
