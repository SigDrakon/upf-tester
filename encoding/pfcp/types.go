package pfcp

type NodeId struct {
	Ipv4 string `yaml:"ipv4"`
	Ipv6 string `yaml:"ipv6"`
}

type FSEID struct {
	SEID        uint64 `yaml:"seid"`
	Ipv4Address string `yaml:"ipv4Address"`
	Ipv6Address string `yaml:"ipv6Address"`
}

type FTEID struct {
	Flag     uint8 `yaml:"flag"`
	ChooseId uint8 `yaml:"chooseId"`
}

type UEAddress struct {
	Flag        uint8  `yaml:"flag"`
	Ipv4Address string `yaml:"ipv4Address"`
	Ipv6Address string `yaml:"ipv6Address"`
}

type PDI struct {
	SourceInterface   *uint8     `yaml:"sourceInterface"`
	FTEID             *FTEID     `yaml:"fteid"`
	UEAddress         *UEAddress `yaml:"ueAddress"`
	SDFFilter         *string    `yaml:"sdfFilter"`
	InterfaceType3gpp *uint8     `yaml:"interfaceType3gpp"`
}

type OuterHeaderRemoval struct {
	Desc uint8 `yaml:"desc"`
	Ext  uint8 `yaml:"ext"`
}

type PDR struct {
	PdrId              uint16              `yaml:"pdrId"`
	Precedence         uint32              `yaml:"precedence"`
	PDI                PDI                 `yaml:"pdi"`
	OuterHeaderRemoval *OuterHeaderRemoval `yaml:"outerHeaderRemoval"`
	FarId              *uint32             `yaml:"farId"`
	UrrId              *uint32             `yaml:"urrId"`
	QerIds             *[]uint32           `yaml:"qerIds"`
}

type OuterHeaderCreation struct {
	OuterHeaderCreationDescription uint16 `yaml:"outerHeaderCreationDescription"`
	TEID                           uint32 `yaml:"teid"`
	IPv4Address                    string `yaml:"ipv4Address"`
	IPv6Address                    string `yaml:"ipv6Address"`
	PortNumber                     uint16 `yaml:"portNumber"`
	CTag                           uint32 `yaml:"cTag"`
	STag                           uint32 `yaml:"sTag"`
}

type ForwardingParameters struct {
	DestinationInterface uint8                `yaml:"destinationInterface"`
	OuterHeaderCreation  *OuterHeaderCreation `yaml:"outerHeaderCreation"`
	InterfaceType3gpp    uint8                `yaml:"interfaceType3gpp"`
}

type FAR struct {
	FarId                uint32                `yaml:"farId"`
	ApplyAction          uint8                 `yaml:"applyAction"`
	ForwardingParameters *ForwardingParameters `yaml:"forwardingParameters"`
}

type URR struct {
	UrrId           uint32           `yaml:"urrId"`
	MeasureMethod   *MeasureMethod   `yaml:"measureMethod"`
	ReportTriggers  *ReportTriggers  `yaml:"reportTriggers"`
	VolumeThreshold *VolumeThreshold `yaml:"volumeThreshold"`
	VolumeQuota     *VolumeQuota     `yaml:"volumeQuota"`
	TimeThreshold   *TimeThreshold   `yaml:"timeThreshold"`
	TimeQuota       *TimeQuota       `yaml:"timeQuota"`
}

type MeasureMethod struct {
	Event    int `yaml:"event"`
	Volum    int `yaml:"volum"`
	Duration int `yaml:"duration"`
}

type ReportTriggers struct {
	Octet1 uint8 `yaml:"octet1"`
	Octet2 uint8 `yaml:"octet2"`
	Octet3 uint8 `yaml:"octet3"`
}

type VolumeThreshold struct {
	Flag  uint8  `yaml:"flag"`
	Tovol uint64 `yaml:"tovol"`
	Ulvol uint64 `yaml:"ulvol"`
	Dlvol uint64 `yaml:"dlvol"`
}
type VolumeQuota struct {
	Flag  uint8  `yaml:"flag"`
	Tovol uint64 `yaml:"tovol"`
	Ulvol uint64 `yaml:"ulvol"`
	Dlvol uint64 `yaml:"dlvol"`
}
type TimeThreshold struct {
	Duration int `yaml:"duration"`
}
type TimeQuota struct {
	Duration int `yaml:"duration"`
}

type GateStatus struct {
	UL uint8 `yaml:"ul"`
	DL uint8 `yaml:"dl"`
}

type MBR struct {
	UL *uint64 `yaml:"ul"`
	DL *uint64 `yaml:"dl"`
}

type QER struct {
	QerId      *uint32     `yaml:"qerId"`
	GateStatus *GateStatus `yaml:"gateStatus"`
	MBR        *MBR        `yaml:"mbr"`
}

type UserID struct {
	Flag   uint8  `yaml:"flag"`
	IMSI   string `yaml:"imsi"`
	IMEI   string `yaml:"imei"`
	MSISDN string `yaml:"msisdn"`
}
