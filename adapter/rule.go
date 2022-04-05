package adapter

import (
	"github.com/Dreamacro/clash/constant"
)

type RuleType int

const (
	Domain RuleType = iota
	DomainKey
	DomainSuffix
	ScrIp
	SrcIPCIDR
	SrcPort
	DstIp
	DstIPCIDR
	DstPort
	Process
	ProcessPath
	ProcessDir
)

func (rt RuleType) String() string {
	switch rt {
	case Domain:
		return "Domain"
	case DomainKey:
		return "DomainKey"
	case DomainSuffix:
		return "DomainSuffix"
	case ScrIp:
		return "ScrIp"
	case SrcIPCIDR:
		return "SrcIPCIDR"
	case SrcPort:
		return "SrcPort"
	case DstIp:
		return "DstIp"
	case DstIPCIDR:
		return "SstIPCIDR"
	case DstPort:
		return "DstPort"
	case Process:
		return "Process"
	case ProcessPath:
		return "ProcessPath"
	case ProcessDir:
		return "ProcessDir"
	default:
		return "Unknown"
	}
}

type Rule interface {
	Match(metadata *constant.Metadata) bool
	Key() string
	Export() RuleInfo
	AdapterType() AdapterType
	Type() RuleType
}

type RuleInfo struct {
	Rule    string `json:"rule,omitempty" yaml:"rule,omitempty"`
	Payload string `json:"payload,omitempty" yaml:"payload,omitempty"`
	Adapter string `json:"adapter,omitempty" yaml:"adapter,omitempty"`
}
