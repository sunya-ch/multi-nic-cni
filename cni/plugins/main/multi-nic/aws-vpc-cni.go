/*
 * Copyright 2022- IBM Inc. All rights reserved
 * SPDX-License-Identifier: Apache2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/vishvananda/netlink"
	"log"
	"net"
)

type EnforcingMode string

const (
	EnforcingModeStrict   EnforcingMode = "strict"
	EnforcingModeStandard EnforcingMode = "standard"
)

const (
	// DefaultEnforcingMode is the default enforcing mode if not specified explicitly.
	DefaultEnforcingMode EnforcingMode = EnforcingModeStrict
	// environment variable knob to decide EnforcingMode for SGPP feature.
	envEnforcingMode = "POD_SECURITY_GROUP_ENFORCING_MODE"
)

type AWSCNINetConfig struct {
	types.NetConf
	MainPlugin AWSChainedCNITypeNetConf `json:"plugin"`
}

type AWSChainedCNITypeNetConf struct {
	types.NetConf

	// VethPrefix is the prefix to use when constructing the host-side
	// veth device name. It should be no more than four characters, and
	// defaults to 'eni'.
	VethPrefix string `json:"vethPrefix"`

	// MTU for eth0
	MTU string `json:"mtu"`

	// PodSGEnforcingMode is the enforcing mode for Security groups for pods feature
	PodSGEnforcingMode EnforcingMode `json:"podSGEnforcingMode"`

	// Interface inside container to create
	IfName string `json:"ifName"`

	//MTU for Egress v4 interface
	EgressMTU int `json:"egressMtu"`

	Enabled string `json:"enabled"`

	RandomizeSNAT string `json:"randomizeSNAT"`

	// IP to use as SNAT target
	NodeIP net.IP `json:"nodeIP"`

	PluginLogFile  string `json:"pluginLogFile"`
	PluginLogLevel string `json:"pluginLogLevel"`
	EgressPluginLogFile  string `json:"egressPluginLogFile"`
	EgressPluginLogLevel string `json:"egressPluginLogLevel"`
}

type AWSCNITypeNetConf struct {
	types.NetConf

	// VethPrefix is the prefix to use when constructing the host-side
	// veth device name. It should be no more than four characters, and
	// defaults to 'eni'.
	VethPrefix string `json:"vethPrefix"`

	// MTU for eth0
	MTU string `json:"mtu"`

	// PodSGEnforcingMode is the enforcing mode for Security groups for pods feature
	PodSGEnforcingMode EnforcingMode `json:"podSGEnforcingMode"`

	// ENIPrimaryAddress is ENI Primary address for specifying eni device.
	ENIPrimaryAddress string `json:"primaryAddress"`

	// Mask is CIDR Mask bit of target interface
	Mask int `json:"mask"`

	PluginLogFile string `json:"pluginLogFile"`

	PluginLogLevel string `json:"pluginLogLevel"`
}

type EgressCNITypeNetConf struct {
	types.NetConf

	// Interface inside container to create
	IfName string `json:"ifName"`

	//MTU for Egress v4 interface
	MTU int `json:"mtu"`

	Enabled string `json:"enabled"`

	RandomizeSNAT string `json:"randomizeSNAT"`

	// IP to use as SNAT target
	NodeIP net.IP `json:"nodeIP"`

	PluginLogFile  string `json:"pluginLogFile"`
	PluginLogLevel string `json:"pluginLogLevel"`
}

// PortMapEntry corresponds to a single entry in the port_mappings argument,
// see CONVENTIONS.md
type PortMapEntry struct {
	HostPort      int    `json:"hostPort"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol"`
	HostIP        string `json:"hostIP,omitempty"`
}

type PortMapTypeNetConf struct {
	types.NetConf
	SNAT                 *bool     `json:"snat,omitempty"`
	ConditionsV4         *[]string `json:"conditionsV4"`
	ConditionsV6         *[]string `json:"conditionsV6"`
	MarkMasqBit          *int      `json:"markMasqBit"`
	ExternalSetMarkChain *string   `json:"externalSetMarkChain"`
	RuntimeConfig        struct {
		PortMaps []PortMapEntry `json:"portMappings,omitempty"`
	} `json:"runtimeConfig,omitempty"`

	// These are fields parsed out of the config or the environment;
	// included here for convenience
	ContainerID string    `json:"-"`
	ContIPv4    net.IPNet `json:"-"`
	ContIPv6    net.IPNet `json:"-"`
}

func getHostIP(devName string) (net.IP, int) {
	devLink, err := netlink.LinkByName(devName)
	if err != nil {
		log.Printf("cannot find link %s: %v", devName, err)
		return nil, -1
	}
	addrs, err := netlink.AddrList(devLink, netlink.FAMILY_V4)
	if err != nil || len(addrs) == 0 {
		log.Printf("cannot list address on %s: %v", devName, err)
		return nil, -1
	}
	ip := addrs[0].IPNet.IP
	if ip == nil {
		return nil, -1
	}
	ones, _ := addrs[0].IPNet.Mask.Size()
	return  ip, ones
}

// loadAWSCNIConf unmarshal to AWSCNINetConfig and returns list of AWSCNI configs
func loadAWSCNIConf(bytes []byte, ifName string, n *NetConf, ipConfigs []*current.IPConfig) (string, []map[string][]byte, []string, error) {
	devTypes := []string{"aws-cni", "egress-v4-cni", "portmap"}
	confBytesArray := []map[string][]byte{}
	version := n.CNIVersion

	configInAWSCNI := &AWSCNINetConfig{}
	if err := json.Unmarshal(bytes, configInAWSCNI); err != nil {
		return version, confBytesArray, devTypes, err
	}
	// interfaces are orderly assigned from interface set
	for _, masterName := range n.Masters {
		nodeIP, ones := getHostIP(masterName)
		if nodeIP == nil {
			continue
		}
		// add config
		awsCNIConfig, egressCNIConfig, portMapConfig, err := getAWSChainedCNIConfig(configInAWSCNI.MainPlugin)
		if err != nil {
			return version, confBytesArray, devTypes, err
		}
		if awsCNIConfig.CNIVersion == "" {
			awsCNIConfig.CNIVersion = n.CNIVersion
		}
		version = awsCNIConfig.CNIVersion
		awsCNIConfig.Name = fmt.Sprintf("aws-cni-%s", masterName)
		awsCNIConfig.ENIPrimaryAddress = nodeIP.String()
		awsCNIConfig.Mask = ones
		egressCNIConfig.Name = fmt.Sprintf("egress-v4-cni-%s", masterName)
		egressCNIConfig.NodeIP = nodeIP
		egressCNIConfig.CNIVersion = awsCNIConfig.CNIVersion
		awsCNIConfBytes, err := json.Marshal(awsCNIConfig)
		if err != nil {
			return version, confBytesArray, devTypes, err
		}
		egressCNIConfBytes, err := json.Marshal(egressCNIConfig)
		if err != nil {
			return version, confBytesArray, devTypes, err
		}
		egressCNIConfBytes = injectSingleNicIPAM(egressCNIConfBytes, bytes)
		portMapConfig.Name = fmt.Sprintf("portmap-%s", masterName)
		portMapConfig.CNIVersion = awsCNIConfig.CNIVersion
		portMapConfBytes, err := json.Marshal(portMapConfig)
		if err != nil {
			return version, confBytesArray, devTypes, err
		}
		confBytesMap := map[string][]byte{
			"aws-cni":       awsCNIConfBytes,
			"egress-v4-cni": egressCNIConfBytes,
			"portmap":       portMapConfBytes,
		}
		confBytesArray = append(confBytesArray, confBytesMap)
	}
	return version, confBytesArray, devTypes, nil
}

// getAWSChainedCNIConfig makes a copy of base AWSCNI config
func getAWSChainedCNIConfig(original AWSChainedCNITypeNetConf) (*AWSCNITypeNetConf, *EgressCNITypeNetConf, *PortMapTypeNetConf, error) {
	awsCNIConfig := &AWSCNITypeNetConf{
		NetConf:            original.NetConf,
		VethPrefix:         original.VethPrefix,
		MTU:                original.MTU,
		PodSGEnforcingMode: original.PodSGEnforcingMode,
		PluginLogFile:      original.PluginLogFile,
		PluginLogLevel:     original.PluginLogLevel,
	}
	awsCNIConfig.Type = "aws-cni"
	egressCNIConfig := &EgressCNITypeNetConf{
		NetConf:        original.NetConf,
		MTU:            original.EgressMTU,
		Enabled:        original.Enabled,
		RandomizeSNAT:  original.RandomizeSNAT,
		PluginLogFile:  original.EgressPluginLogFile,
		PluginLogLevel: original.EgressPluginLogLevel,
	}
	egressCNIConfig.Type = "egress-v4-cni"
	trueValue := true
	portMapConfig := &PortMapTypeNetConf{
		NetConf:        original.NetConf,
		SNAT:           &trueValue,
	}
	if portMapConfig.NetConf.Capabilities == nil {
		portMapConfig.NetConf.Capabilities = make(map[string]bool)
		portMapConfig.NetConf.Capabilities["portMappings"] = true
	}
	portMapConfig.Type = "portmap"
	return awsCNIConfig, egressCNIConfig, portMapConfig, nil
}
