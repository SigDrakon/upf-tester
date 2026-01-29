package pfcp

import (
	"log"
	"net"
	"os"
	"time"
	"upftester/internal/util"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
	"gopkg.in/yaml.v3"
)

type EstablishmentRequestConfig struct {
	NodeId     *NodeId `yaml:"nodeId"`
	FSEID      *FSEID  `yaml:"fseid"`
	CreatePDRs *[]PDR  `yaml:"createPdrs"`
	CreateFARs *[]FAR  `yaml:"createFars"`
	CreateURRs *[]URR  `yaml:"createUrrs"`
	CreateQERs *[]QER  `yaml:"createQers"`
	PDNType    *uint8  `yaml:"pdnType"`
	ApnDnn     string  `yaml:"apnDnn"`
	UserID     *UserID `yaml:"userId"`
}

func (cfg *EstablishmentRequestConfig) Marshal(path string) (message.Message, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading file %s: %v", path, err)
		return nil, err
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		log.Printf("Error unmarshalling file %s: %v", path, err)
		return nil, err
	}

	var ies []*ie.IE

	if cfg.NodeId != nil {
		ies = append(ies, ie.NewNodeID(cfg.NodeId.Ipv4, cfg.NodeId.Ipv6, ""))
	}

	if cfg.FSEID != nil {
		cfg.FSEID.SEID = util.GlobalSeid.Inc()
		cfg.FSEID.Ipv4Address = cfg.NodeId.Ipv4
		cfg.FSEID.Ipv6Address = cfg.NodeId.Ipv6
		ies = append(ies, ie.NewFSEID(cfg.FSEID.SEID, net.ParseIP(cfg.FSEID.Ipv4Address), net.ParseIP(cfg.FSEID.Ipv6Address)))
	}

	if cfg.CreatePDRs != nil {

		for _, pdr := range *cfg.CreatePDRs {

			var pdiChildren []*ie.IE

			if pdr.PDI.SourceInterface != nil {
				pdiChildren = append(pdiChildren,
					ie.NewSourceInterface(*pdr.PDI.SourceInterface))
			}

			if pdr.PDI.FTEID != nil {
				pdiChildren = append(pdiChildren,
					ie.NewFTEID(pdr.PDI.FTEID.Flag, 0, nil, nil, pdr.PDI.FTEID.ChooseId))
			}

			if pdr.PDI.UEAddress != nil {
				pdiChildren = append(pdiChildren,
					ie.NewUEIPAddress(pdr.PDI.UEAddress.Flag, pdr.PDI.UEAddress.Ipv4Address, pdr.PDI.UEAddress.Ipv6Address, 0, 0))
			}

			if pdr.PDI.SDFFilter != nil {
				pdiChildren = append(pdiChildren,
					ie.NewSDFFilter(*pdr.PDI.SDFFilter, "", "", "", 0))
			}

			if pdr.PDI.InterfaceType3gpp != nil {
				pdiChildren = append(pdiChildren,
					ie.NewTGPPInterfaceType(*pdr.PDI.InterfaceType3gpp))
			}

			pdi := ie.NewPDI(pdiChildren...)

			var pdrChildren []*ie.IE

			pdrChildren = append(pdrChildren, ie.NewPDRID(pdr.PdrId))

			pdrChildren = append(pdrChildren, ie.NewPrecedence(pdr.Precedence))

			pdrChildren = append(pdrChildren, pdi)

			if pdr.OuterHeaderRemoval != nil {
				pdrChildren = append(pdrChildren,
					ie.NewOuterHeaderRemoval(pdr.OuterHeaderRemoval.Desc, pdr.OuterHeaderRemoval.Ext))
			}

			if pdr.FarId != nil {
				pdrChildren = append(pdrChildren, ie.NewFARID(*pdr.FarId))
			}

			if pdr.UrrId != nil {
				pdrChildren = append(pdrChildren, ie.NewURRID(*pdr.UrrId))
			}

			if pdr.QerIds != nil {
				for _, qid := range *pdr.QerIds {
					pdrChildren = append(pdrChildren, ie.NewQERID(qid))
				}
			}

			ies = append(ies, ie.NewCreatePDR(pdrChildren...))
		}

	}

	if cfg.CreateFARs != nil {
		for _, far := range *cfg.CreateFARs {

			farId := ie.NewFARID(far.FarId)

			apply := ie.NewApplyAction(far.ApplyAction)

			fp := ie.NewForwardingParameters(
				ie.NewDestinationInterface(far.ForwardingParameters.DestinationInterface),
				ie.NewTGPPInterfaceType(far.ForwardingParameters.InterfaceType3gpp),
			)

			if far.ForwardingParameters.OuterHeaderCreation != nil {

				far.ForwardingParameters.OuterHeaderCreation.TEID = util.GlobalSeqNumber.Inc()

				fp.Add(ie.NewOuterHeaderCreation(
					far.ForwardingParameters.OuterHeaderCreation.OuterHeaderCreationDescription,
					far.ForwardingParameters.OuterHeaderCreation.TEID,
					far.ForwardingParameters.OuterHeaderCreation.IPv4Address,
					far.ForwardingParameters.OuterHeaderCreation.IPv6Address,
					far.ForwardingParameters.OuterHeaderCreation.PortNumber,
					far.ForwardingParameters.OuterHeaderCreation.CTag,
					far.ForwardingParameters.OuterHeaderCreation.STag,
				))
			}

			ies = append(ies, ie.NewCreateFAR(farId, apply, fp))
		}
	}

	if cfg.CreateURRs != nil {
		for _, urr := range *cfg.CreateURRs {

			var urrChildren []*ie.IE

			urrChildren = append(urrChildren, ie.NewURRID(urr.UrrId))

			if urr.MeasureMethod != nil {
				urrChildren = append(urrChildren,
					ie.NewMeasurementMethod(urr.MeasureMethod.Event, urr.MeasureMethod.Volum, urr.MeasureMethod.Duration))
			}

			if urr.ReportTriggers != nil {
				urrChildren = append(urrChildren,
					ie.NewReportingTriggers(urr.ReportTriggers.Octet1, urr.ReportTriggers.Octet2, urr.ReportTriggers.Octet3))
			}

			if urr.VolumeThreshold != nil {
				urrChildren = append(urrChildren,
					ie.NewVolumeThreshold(urr.VolumeThreshold.Flag, urr.VolumeThreshold.Tovol, urr.VolumeThreshold.Ulvol, urr.VolumeThreshold.Dlvol))
			}

			if urr.VolumeQuota != nil {
				urrChildren = append(urrChildren,
					ie.NewVolumeQuota(urr.VolumeQuota.Flag, urr.VolumeQuota.Tovol, urr.VolumeQuota.Ulvol, urr.VolumeThreshold.Dlvol))
			}

			if urr.TimeThreshold != nil {
				urrChildren = append(urrChildren,
					ie.NewTimeThreshold(time.Duration(urr.TimeThreshold.Duration)*time.Second))
			}

			if urr.TimeQuota != nil {
				urrChildren = append(urrChildren,
					ie.NewTimeQuota(time.Duration(urr.TimeQuota.Duration)*time.Second))
			}

			ies = append(ies,
				ie.NewCreateURR(urrChildren...))
		}
	}

	if cfg.CreateQERs != nil {
		for _, q := range *cfg.CreateQERs {
			qerID := ie.NewQERID(*q.QerId)
			gate := ie.NewGateStatus((*q.GateStatus).UL, (*q.GateStatus).DL)
			mbr := ie.NewMBR(*q.MBR.UL, *q.MBR.DL)
			createQER := ie.NewCreateQER(qerID, gate, mbr)
			ies = append(ies, createQER)
		}
	}

	if cfg.PDNType != nil {
		ies = append(ies, ie.NewPDNType(*cfg.PDNType))
	}

	if cfg.ApnDnn != "" {
		ies = append(ies, ie.NewAPNDNN(cfg.ApnDnn))
	}

	if cfg.UserID != nil {
		ies = append(ies, ie.NewUserID(cfg.UserID.Flag, cfg.UserID.IMSI, "", "", ""))
	}

	return message.NewSessionEstablishmentRequest(0, 0, 0, util.GlobalSeqNumber.Inc(), 0, ies...), nil
}

func (cfg *EstablishmentRequestConfig) Unmarshal() {

}
