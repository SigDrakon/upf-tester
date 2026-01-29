package pfcp

import (
	"log"
	"os"
	"upftester/internal/util"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
	"gopkg.in/yaml.v3"
)

type ModificationRequestConfig struct {
	FarId         *uint32               `yaml:"farId"`
	ApplyAction   *uint8                `yaml:"applyAction"`
	UpdateFar     *ForwardingParameters `yaml:"forwardingParameters"`
	PfcpSmReqFlag *uint8                `yaml:"pfcpSmReqFlag"`
}

func (cfg *ModificationRequestConfig) Marshal(path string) (message.Message, error) {

	log.Printf("marshal session modification request config from path: %s", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	var ies []*ie.IE

	if cfg.FarId != nil {
		ies = append(ies, ie.NewFARID(*cfg.FarId))
	}

	if cfg.ApplyAction != nil {
		ies = append(ies, ie.NewApplyAction(*cfg.ApplyAction))
	}

	if cfg.UpdateFar != nil {
		fp := ie.NewForwardingParameters(
			ie.NewDestinationInterface(cfg.UpdateFar.DestinationInterface),
			ie.NewTGPPInterfaceType(cfg.UpdateFar.InterfaceType3gpp),
		)

		if cfg.UpdateFar.OuterHeaderCreation != nil {

			cfg.UpdateFar.OuterHeaderCreation.TEID = util.GlobalSeqNumber.Inc()

			fp.Add(ie.NewOuterHeaderCreation(
				cfg.UpdateFar.OuterHeaderCreation.OuterHeaderCreationDescription,
				cfg.UpdateFar.OuterHeaderCreation.TEID,
				cfg.UpdateFar.OuterHeaderCreation.IPv4Address,
				cfg.UpdateFar.OuterHeaderCreation.IPv6Address,
				cfg.UpdateFar.OuterHeaderCreation.PortNumber,
				cfg.UpdateFar.OuterHeaderCreation.CTag,
				cfg.UpdateFar.OuterHeaderCreation.STag,
			))
		}

		ies = append(ies, fp)
	}

	if cfg.PfcpSmReqFlag != nil {
		log.Printf("pfcp sm req flag: 0x%02x", *cfg.PfcpSmReqFlag)
		ies = append(ies, ie.NewPFCPSMReqFlags(*cfg.PfcpSmReqFlag))
	}

	return message.NewSessionModificationRequest(0, 0, 0, util.GlobalSeqNumber.Inc(), 0, ies...), nil
}

func (cfg *ModificationRequestConfig) Unmarshal() {

}
