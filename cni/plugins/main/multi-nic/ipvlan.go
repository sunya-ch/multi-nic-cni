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
)

// IPVLANNetConfig defines ipvlan net config
// Master string `json:"master"`
// Mode   string `json:"mode"`
// MTU    int    `json:"mtu"`
type IPVLANNetConfig struct {
	types.NetConf
	MainPlugin IPVLANTypeNetConf `json:"plugin"`
}

type IPVLANTypeNetConf struct {
	types.NetConf
	Master string `json:"master"`
	Mode   string `json:"mode"`
	MTU    int    `json:"mtu"`
}

// loadIPVANConf unmarshal to IPVLANNetConfig and returns list of IPVLAN configs
func loadIPVANConf(bytes []byte, ifName string, n *NetConf, ipConfigs []*current.IPConfig) ([]map[string][]byte, []string, error) {
	devTypes := []string{"ipvlan"}
	confBytesArray := []map[string][]byte{}

	configInIPVLAN := &IPVLANNetConfig{}
	if err := json.Unmarshal(bytes, configInIPVLAN); err != nil {
		return confBytesArray, devTypes, err
	}

	// interfaces are orderly assigned from interface set
	for index, masterName := range n.Masters {
		// add config
		singleConfig, err := copyIPVLANConfig(configInIPVLAN.MainPlugin)
		if err != nil {
			return confBytesArray, devTypes, err
		}
		if singleConfig.CNIVersion == "" {
			singleConfig.CNIVersion = n.CNIVersion
		}
		singleConfig.Name = fmt.Sprintf("%s-%d", ifName, index)
		singleConfig.Master = masterName
		confBytes, err := json.Marshal(singleConfig)
		if err != nil {
			return confBytesArray, devTypes, err
		}
		if n.IsMultiNICIPAM {
			// multi-NIC IPAM config
			if index < len(ipConfigs) {
				confBytes = injectMultiNicIPAM(confBytes, ipConfigs, index)
				confBytesMap := map[string][]byte{
					"ipvlan": confBytes,
				}
				confBytesArray = append(confBytesArray, confBytesMap)
			}
		} else {
			confBytes = injectSingleNicIPAM(confBytes, bytes)
			confBytesMap := map[string][]byte{
				"ipvlan": confBytes,
			}
			confBytesArray = append(confBytesArray, confBytesMap)
		}
	}
	return confBytesArray, devTypes, nil
}

// copyIPVLANConfig makes a copy of base IPVLAN config
func copyIPVLANConfig(original IPVLANTypeNetConf) (*IPVLANTypeNetConf, error) {
	copiedObject := &IPVLANTypeNetConf{}
	byteObject, err := json.Marshal(original)
	if err != nil {
		return copiedObject, err
	}
	err = json.Unmarshal(byteObject, copiedObject)
	if err != nil {
		return copiedObject, err
	}
	return copiedObject, nil
}
