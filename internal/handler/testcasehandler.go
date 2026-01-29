package handler

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
	"upftester/encoding"
	"upftester/encoding/pfcp"
	"upftester/internal/config"
	"upftester/internal/dataplane"
	"upftester/internal/network"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
	"gopkg.in/yaml.v3"
)

var GlobalTestCases = make([][]TestCase, 0)

type TestCase struct {
	Type    string
	Action  string
	Path    string
	Config  encoding.MessageConfig
	Message message.Message
}

type TestStep struct {
	Step   int    `yaml:"step"`
	Type   string `yaml:"type"`
	Action string `yaml:"action"`
	Path   string `yaml:"path"`
}

func LoadTestCases(path string, globalTestCases *[][]TestCase) {

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("read test case file failed: %v", err)
		return
	}

	var wrapper struct {
		TestSteps []TestStep `yaml:"testSteps"`
	}
	err = yaml.Unmarshal(data, &wrapper)
	if err != nil {
		log.Printf("unmarshal test case file failed: %v", err)
		return
	}

	sort.Slice(wrapper.TestSteps, func(i, j int) bool {
		return wrapper.TestSteps[i].Step < wrapper.TestSteps[j].Step
	})

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("resolve absolute path failed: %v", err)
	}
	yamlDir := filepath.Dir(absPath)

	for i := range wrapper.TestSteps {
		if wrapper.TestSteps[i].Path != "" && wrapper.TestSteps[i].Type != "sleep" {
			wrapper.TestSteps[i].Path = filepath.Join(yamlDir, "yaml", wrapper.TestSteps[i].Path)
		}
	}

	testCases := make([]TestCase, 0, len(wrapper.TestSteps))
	for _, step := range wrapper.TestSteps {
		var msg message.Message
		var msgConfig encoding.MessageConfig

		switch step.Type {
		case "session_establishment_request":
			msgConfig = new(pfcp.EstablishmentRequestConfig)
			msg, err = msgConfig.Marshal(step.Path)
			if err != nil {
				log.Fatal(err)
				return
			}

		case "session_modification_request":
			msgConfig = new(pfcp.ModificationRequestConfig)
			msg, err = msgConfig.Marshal(step.Path)
			if err != nil {
				log.Fatal(err)
				return
			}

		case "session_deletion_request":
			msgConfig = new(pfcp.DeletionRequestConfig)
			msg, err = msgConfig.Marshal(step.Path)
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		testCases = append(testCases, TestCase{
			Type:    step.Type,
			Action:  step.Action,
			Path:    step.Path,
			Config:  msgConfig,
			Message: msg,
		})
	}

	*globalTestCases = append(*globalTestCases, testCases)
}

func RunTestCases(remoteAddr *net.UDPAddr, conn *network.UDPTransport) {
	var wg sync.WaitGroup
	for i, testCases := range GlobalTestCases {
		wg.Add(1)
		go func(tc []TestCase, index int) {
			defer wg.Done()
			log.Printf("Starting test case set %d", index)
			err := HandleSingleTest(tc, remoteAddr, conn)
			if err != nil {
				log.Printf("Test case set %d failed: %v", index, err)
			} else {
				log.Printf("Test case set %d completed successfully", index)
			}
		}(testCases, i)
	}
	wg.Wait()
}

func HandleSingleTest(testCases []TestCase, remoteAddr *net.UDPAddr, conn *network.UDPTransport) error {

	ch := make(chan *PFCPMessage, 5)

	var upfSeid uint64
	var smfSeid uint64
	var sessionCtx *SessionContext

	for _, testcase := range testCases {
		switch testcase.Type {
		case "session_establishment_request":

			msg := testcase.Config.(*pfcp.EstablishmentRequestConfig)
			smfSeid = msg.FSEID.SEID
			GetPFCPDispatcher().Register(smfSeid, ch)

			// 创建会话上下文
			sessionCtx = &SessionContext{
				SEID:  smfSeid,
				State: SessionStateEstablishing,
			}
			if msg.CreatePDRs != nil && len(*msg.CreatePDRs) > 0 {
				for _, pdr := range *msg.CreatePDRs {
					if pdr.PDI.UEAddress != nil {
						sessionCtx.UEIP = pdr.PDI.UEAddress.Ipv4Address
					}
					// Check for Uplink PDR (SourceInterface = Access)
					if pdr.PDI.SourceInterface != nil && *pdr.PDI.SourceInterface == 0 {
						sessionCtx.UplinkPDRID = pdr.PdrId
						log.Printf("Identified Uplink PDR ID: %d", sessionCtx.UplinkPDRID)
					}
				}
			}
			GlobalSessionManager.AddSession(smfSeid, sessionCtx)

			data := make([]byte, (testcase.Message).MarshalLen())
			err := testcase.Message.MarshalTo(data)
			if err != nil {
				log.Println("marshal session establishment request failed:", err)
				return err
			}

			log.Printf("Sending session establishment request, SEID: 0x%016x", smfSeid)
			conn.Send(data, remoteAddr)


		case "session_establishment_response":
			select {
			case <-time.After(time.Second * 5):
				return fmt.Errorf("wait session establishment response timeout")

			case msg := <-ch:

				if msg.MessageType != message.MsgTypeSessionEstablishmentResponse {
					log.Printf("expect session establishment response, but got %v", msg.MessageType)
					return fmt.Errorf("expect session establishment response, but got %v", msg.MessageType)
				}

				resp, err := message.ParseSessionEstablishmentResponse(msg.Payload)
				if err != nil {
					log.Println("session establishment response parse failed:", err)
					return err
				}

				value, err := resp.Cause.ValueAsUint8()
				if err != nil {
					log.Println("session establishment response cause parse failed:", err)
					return err
				}

				if value != ie.CauseRequestAccepted {
					log.Println("session establishment request was rejected")
					return fmt.Errorf("session establishment request was rejected")
				}

				fseid, err := resp.UPFSEID.FSEID()
				if err != nil {
					log.Println("session establishment response fseid parse failed:", err)
					return err
				}

				upfSeid = fseid.SEID
				log.Printf("Session established successfully, SMF SEID: 0x%016x, UPF SEID: 0x%016x", smfSeid, upfSeid)

				// Update session context
				if sessionCtx != nil {
					sessionCtx.UPFSEID = upfSeid
					sessionCtx.State = SessionStateActive

					// Parse Created PDRs to get allocated F-TEID
					// We match the Created PDR with our identified Uplink PDR ID
					for _, item := range resp.CreatedPDR {
						pdrId, err := item.PDRID()
						if err != nil {
							continue
						}
						
						if pdrId == sessionCtx.UplinkPDRID {
							fteid, err := item.FTEID()
							if err == nil {
								sessionCtx.UplinkTEID = fteid.TEID
								log.Printf("Updated Uplink TEID: %d (from PDR ID: %d)", sessionCtx.UplinkTEID, pdrId)
								break
							}
						}
					}
					
					GlobalSessionManager.UpdateSession(smfSeid, sessionCtx)
				}
			}

		case "session_modification_request":
			msg := testcase.Message.(*message.SessionModificationRequest)
			msg.Header.SEID = upfSeid

			if sessionCtx != nil {
				sessionCtx.State = SessionStateModifying
				GlobalSessionManager.UpdateSession(smfSeid, sessionCtx)
			}

			data := make([]byte, (testcase.Message).MarshalLen())
			err := testcase.Message.MarshalTo(data)
			if err != nil {
				log.Println("marshal session modification request failed:", err)
				return err
			}

			log.Printf("Sending session modification request, UPF SEID: 0x%016x", upfSeid)
			conn.Send(data, remoteAddr)

		case "session_modification_response":
			select {
			case <-time.After(time.Second * 5):
				return fmt.Errorf("wait session modification response timeout")

			case msg := <-ch:
				if msg.MessageType != message.MsgTypeSessionModificationResponse {
					log.Printf("expect session modification response, but got %v", msg.MessageType)
					return fmt.Errorf("expect session modification response, but got %v", msg.MessageType)
				}

				resp, err := message.ParseSessionModificationResponse(msg.Payload)
				if err != nil {
					log.Println("session modification response parse failed:", err)
					return err
				}

				value, err := resp.Cause.ValueAsUint8()
				if err != nil {
					log.Println("session modification response cause parse failed:", err)
					return err
				}

				if value != ie.CauseRequestAccepted {
					log.Printf("session modification request was rejected, cause: %d", value)
					return fmt.Errorf("session modification request was rejected")
				}

				log.Printf("Session modified successfully")

				if sessionCtx != nil {
					sessionCtx.State = SessionStateActive
					GlobalSessionManager.UpdateSession(smfSeid, sessionCtx)
				}
			}

		case "session_deletion_request":
			msg := testcase.Message.(*message.SessionDeletionRequest)
			msg.Header.SEID = upfSeid

			if sessionCtx != nil {
				sessionCtx.State = SessionStateDeleting
				GlobalSessionManager.UpdateSession(smfSeid, sessionCtx)
			}

			data := make([]byte, (testcase.Message).MarshalLen())
			err := testcase.Message.MarshalTo(data)
			if err != nil {
				log.Println("marshal session deletion request failed:", err)
				return err
			}

			log.Printf("Sending session deletion request, UPF SEID: 0x%016x", upfSeid)
			conn.Send(data, remoteAddr)

		case "session_deletion_response":
			select {
			case <-time.After(time.Second * 5):
				return fmt.Errorf("wait session deletion response timeout")

			case msg := <-ch:
				if msg.MessageType != message.MsgTypeSessionDeletionResponse {
					log.Printf("expect session deletion response, but got %v", msg.MessageType)
					return fmt.Errorf("expect session deletion response, but got %v", msg.MessageType)
				}

				resp, err := message.ParseSessionDeletionResponse(msg.Payload)
				if err != nil {
					log.Println("session deletion response parse failed:", err)
					return err
				}

				value, err := resp.Cause.ValueAsUint8()
				if err != nil {
					log.Println("session deletion response cause parse failed:", err)
					return err
				}

				if value != ie.CauseRequestAccepted {
					log.Printf("session deletion request was rejected, cause: %d", value)
					return fmt.Errorf("session deletion request was rejected")
				}

				log.Printf("Session deleted successfully, SEID: 0x%016x", smfSeid)

				// 清理会话上下文
				GlobalSessionManager.DeleteSession(smfSeid)
				GetPFCPDispatcher().Unregister(smfSeid)
			}

		case "sleep":
			// 从 path 字段解析睡眠时长（秒）
			duration := 5 // 默认 5 秒
			if testcase.Path != "" {
				var err error
				duration, err = strconv.Atoi(testcase.Path)
				if err != nil {
					log.Printf("parse sleep duration failed: %v, using default 5s", err)
					duration = 5
				}
			}
			log.Printf("Sleeping for %d seconds...", duration)
			time.Sleep(time.Duration(duration) * time.Second)

		case "data_plane_test":
			// 数据平面测试
			if sessionCtx == nil {
				log.Println("No active session for data plane test")
				return fmt.Errorf("no active session for data plane test")
			}

			// 加载数据平面测试配置
			config, err := dataplane.LoadDataPlaneTestConfig(testcase.Path)
			if err != nil {
				log.Printf("Load data plane test config failed: %v", err)
				return err
			}

			// 获取全局配置
			globalConfig, err := getGlobalConfig()
			if err != nil {
				log.Printf("Get global config failed: %v", err)
				return err
			}

			// 确定目标 IP
			dstIp := globalConfig.DataPlane.DnIp
			if config.DstIp != "" {
				dstIp = config.DstIp
			}

			// 根据测试类型创建测试
			switch testcase.Action {
			case "icmp":
				// 创建 ICMP 测试
				icmpTest := dataplane.NewICMPTest(
					config,
					globalConfig.DataPlane.GnbIp,
					globalConfig.DataPlane.N3Ip,
					sessionCtx.UplinkTEID,
					sessionCtx.UEIP,
					dstIp,
				)

				err = icmpTest.Start()
				if err != nil {
					log.Printf("Start ICMP test failed: %v", err)
					return err
				}

				// 等待测试完成
				time.Sleep(time.Duration(config.Duration) * time.Second)
				icmpTest.Stop()

				result := icmpTest.GetResult()
				log.Printf("ICMP Test completed: Sent=%d, Success=%v", result.PacketsSent, result.Success)

			default:
				log.Printf("Unsupported data plane test action: %s", testcase.Action)
				return fmt.Errorf("unsupported data plane test action: %s", testcase.Action)
			}
		}

	}

	return nil
}

// getGlobalConfig 获取全局配置（临时实现，后续需要改进）
func getGlobalConfig() (*config.Config, error) {
	var cfg config.Config
	err := cfg.LoadConfig("../config/config.yaml")
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
