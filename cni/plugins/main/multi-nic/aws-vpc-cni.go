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

func getHostIP(devName string) net.IP {
	devLink, err := netlink.LinkByName(devName)
	if err != nil {
		log.Printf("cannot find link %s: %v", devName, err)
		return nil
	}
	addrs, err := netlink.AddrList(devLink, netlink.FAMILY_V4)
	if err != nil || len(addrs) == 0 {
		log.Printf("cannot list address on %s: %v", devName, err)
		return nil
	}
	return addrs[0].IPNet.IP
}

// loadAWSCNIConf unmarshal to AWSCNINetConfig and returns list of AWSCNI configs
func loadAWSCNIConf(bytes []byte, ifName string, n *NetConf, ipConfigs []*current.IPConfig) ([]map[string][]byte, []string, error) {
	devTypes := []string{"aws-cni", "egress-v4-cni"}
	confBytesArray := []map[string][]byte{}

	configInAWSCNI := &AWSCNINetConfig{}
	if err := json.Unmarshal(bytes, configInAWSCNI); err != nil {
		fmt.Println(err)
		return confBytesArray, devTypes, err
	}
	fmt.Println(n.Masters)
	// interfaces are orderly assigned from interface set
	for _, masterName := range n.Masters {
		nodeIP := getHostIP(masterName)
		if nodeIP == nil {
			fmt.Println("cannot get nodeIP")
			continue
		}
		fmt.Println(nodeIP.To4())
		// add config
		awsCNIConfig, egressCNIConfig, err := getAWSChainedCNIConfig(configInAWSCNI.MainPlugin)
		if err != nil {
			fmt.Println(err)
			return confBytesArray, devTypes, err
		}
		if awsCNIConfig.CNIVersion == "" {
			awsCNIConfig.CNIVersion = n.CNIVersion
		}
		awsCNIConfig.Name = fmt.Sprintf("aws-cni-%s", masterName)
		egressCNIConfig.Name = fmt.Sprintf("egress-v4-cni-%s", masterName)
		egressCNIConfig.NodeIP = nodeIP
		awsCNIConfBytes, err := json.Marshal(awsCNIConfig)
		if err != nil {
			fmt.Println(err)
			return confBytesArray, devTypes, err
		}
		egressCNIConfBytes, err := json.Marshal(egressCNIConfig)
		if err != nil {
			fmt.Println(err)
			return confBytesArray, devTypes, err
		}
		egressCNIConfBytes = injectSingleNicIPAM(egressCNIConfBytes, bytes)
		confBytesMap := map[string][]byte{
			"aws-cni":       awsCNIConfBytes,
			"egress-v4-cni": egressCNIConfBytes,
		}
		confBytesArray = append(confBytesArray, confBytesMap)

	}
	return confBytesArray, devTypes, nil
}

// getAWSChainedCNIConfig makes a copy of base AWSCNI config
func getAWSChainedCNIConfig(original AWSChainedCNITypeNetConf) (*AWSCNITypeNetConf, *EgressCNITypeNetConf, error) {
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
		PluginLogFile:  original.PluginLogFile,
		PluginLogLevel: original.PluginLogLevel,
	}
	egressCNIConfig.Type = "egress-v4-cni"
	return awsCNIConfig, egressCNIConfig, nil
}
