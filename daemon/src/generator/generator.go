/*
 * Copyright 2022- IBM Inc. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */

package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/containernetworking/cni/pkg/types"
	"google.golang.org/grpc"

	"github.com/foundation-model-stack/multi-nic-cni/daemon/allocator"
	"github.com/foundation-model-stack/multi-nic-cni/daemon/generator/ipam"
	pb "github.com/foundation-model-stack/multi-nic-cni/daemon/generator/proto"
	"github.com/foundation-model-stack/multi-nic-cni/daemon/selector"
)

const (
	DEFAULT_SUBNET = "172.30.0.0/16"
)

// NetConf defines general config for multi-nic-cni
type NetConf struct {
	types.NetConf
	MainPlugin     map[string]interface{} `json:"plugin"`
	Subnet         string                 `json:"subnet"`
	MasterNetAddrs []string               `json:"masterNets"`
	Masters        []string               `json:"masters,omitempty"`
	DeviceIDs      []string               `json:"deviceIDs,omitempty"`
	IsMultiNICIPAM bool                   `json:"multiNICIPAM,omitempty"`
	DaemonIP       string                 `json:"daemonIP"`
	DaemonPort     int                    `json:"daemonPort"`
	Args           struct {
		NicSet selector.NicArgs `json:"cni,omitempty"`
	} `json:"args"`
}

// server is used to implement GeneratorServer.
type ImplementedGeneratorServer struct {
	pb.UnimplementedGeneratorServer
}

func selectNics(configRequest *pb.ConfigRequest, n *NetConf) {
	// select NICs
	nicReq := selector.NICSelectRequest{
		PodName:          configRequest.PodName,
		PodNamespace:     configRequest.PodNamespace,
		HostName:         configRequest.HostName,
		NetAttachDefName: n.Name,
		MasterNetAddrs:   n.MasterNetAddrs,
		NicSet:           n.Args.NicSet,
	}
	selectResp := selector.Select(nicReq)
	n.Masters = selectResp.Masters
	n.DeviceIDs = selectResp.DeviceIDs
}

func getPlugins(n *NetConf) (bool, map[string]interface{}, []interface{}, error) {
	mainPlugin := make(map[string]interface{})
	var chainedPlugins []interface{}
	pluginType := n.MainPlugin["type"]
	simplePluginFlag := pluginType != nil && pluginType.(string) != ""
	if !simplePluginFlag {
		pluginsInterface, chainedPlugin := n.MainPlugin["plugins"]
		plugins := pluginsInterface.([]interface{})
		if !chainedPlugin || len(plugins) == 0 {
			return simplePluginFlag, nil, nil, fmt.Errorf("main plugin has no `type` or `plugin`")
		}
		mainPlugin = plugins[0].(map[string]interface{})
		chainedPlugins = plugins[1:len(plugins)]
	} else {
		mainPlugin = n.MainPlugin
	}
	return simplePluginFlag, mainPlugin, chainedPlugins, nil
}

func loadNetConf(configRequest *pb.ConfigRequest) (*NetConf, bool, map[string]interface{}, []interface{}, error) {
	n := &NetConf{}
	if err := json.Unmarshal(configRequest.Data, n); err != nil {
		return nil, false, nil, nil, err
	}
	selectNics(configRequest, n)
	simplePluginFlag, mainPlugin, chainedPlugins, err := getPlugins(n)
	return n, simplePluginFlag, mainPlugin, chainedPlugins, err
}

func generateConfigBytesArray(n *NetConf, ipResponses map[string]allocator.IPResponse, simplePluginFlag bool, mainPlugin map[string]interface{}, chainedPlugins []interface{}) [][]byte {
	// get device config and apply
	confBytesArray := [][]byte{}
	for index, masterName := range n.Masters {
		copiedMainConfig := make(map[string]interface{})
		for key, value := range mainPlugin {
			copiedMainConfig[key] = value
		}
		if n.IsMultiNICIPAM {
			ipResponse := ipResponses[masterName]
			vlanPodCIDR := fmt.Sprintf("%s/%s", ipResponse.IPAddress, ipResponse.VLANBlockSize)
			staticIPAM := map[string]interface{}{
				"type": "static",
				"addresses": []map[string]string{
					map[string]string{"address": vlanPodCIDR},
				},
			}
			copiedMainConfig["ipam"] = staticIPAM
		} else {
			copiedMainConfig["ipam"] = n.IPAM
		}
		deviceID := n.DeviceIDs[index]
		copiedMainConfig["master"] = masterName
		copiedMainConfig["deviceID"] = deviceID
		var confBytes []byte
		if simplePluginFlag {
			confBytes, _ = json.Marshal(copiedMainConfig)
		} else {
			copiedChainedPlugin := make(map[string]interface{})
			for key, value := range n.MainPlugin {
				switch key {
				case "plugins":
					copiedChainedPlugin[key] = append([]interface{}{copiedMainConfig}, chainedPlugins...)
				default:
					copiedChainedPlugin[key] = value
				}
			}
			confBytes, _ = json.Marshal(copiedChainedPlugin)
		}
		confBytesArray = append(confBytesArray, confBytes)
		log.Println(fmt.Sprintf("copiedMainConfig: %v (%s)", copiedMainConfig, string(confBytes)))
	}
	return confBytesArray
}

// Generate generates config list from the configuration request
func (s *ImplementedGeneratorServer) Generate(ctx context.Context, configRequest *pb.ConfigRequest) (*pb.GenerateResponse, error) {
	n, simplePluginFlag, mainPlugin, chainedPlugins, err := loadNetConf(configRequest)
	if err != nil {
		return &pb.GenerateResponse{Success: false, ConfList: [][]byte{}, Message: err.Error()}, nil
	}
	var ipResponses map[string]allocator.IPResponse
	if n.IsMultiNICIPAM {
		// allocate new IPs
		allocateReq := allocator.IPRequest{
			PodName:          configRequest.PodName,
			PodNamespace:     configRequest.PodNamespace,
			HostName:         configRequest.HostName,
			NetAttachDefName: n.Name,
			InterfaceNames:   n.Masters,
		}
		var ipamInstance ipam.IPAMInterface
		switch n.IPAM.Type {
		case ipam.MULTI_NIC_IPAM_TYPE:
			ipamInstance = ipam.MultiNicIPAM{}
		}
		ipResponses = ipamInstance.AllocateIPs(allocateReq)
	}
	confBytesArray := generateConfigBytesArray(n, ipResponses, simplePluginFlag, mainPlugin, chainedPlugins)
	if len(confBytesArray) == 0 {
		return &pb.GenerateResponse{Success: false, ConfList: confBytesArray, Message: "no config is generated."}, nil
	}
	return &pb.GenerateResponse{Success: true, ConfList: confBytesArray, Message: "succeeded"}, nil
}

func (s *ImplementedGeneratorServer) Cleanup(ctx context.Context, configRequest *pb.ConfigRequest) (*pb.GenerateResponse, error) {
	n, simplePluginFlag, mainPlugin, chainedPlugins, err := loadNetConf(configRequest)
	if err != nil {
		return &pb.GenerateResponse{Success: false, ConfList: [][]byte{}, Message: err.Error()}, nil
	}
	var ipResponses map[string]allocator.IPResponse
	if n.IsMultiNICIPAM {
		// get IP map
		allocateReq := allocator.IPRequest{
			PodName:          configRequest.PodName,
			PodNamespace:     configRequest.PodNamespace,
			HostName:         configRequest.HostName,
			NetAttachDefName: n.Name,
			InterfaceNames:   n.Masters,
		}
		var ipamInstance ipam.IPAMInterface
		switch n.IPAM.Type {
		case ipam.MULTI_NIC_IPAM_TYPE:
			ipamInstance = ipam.MultiNicIPAM{}
		}
		ipResponses = ipamInstance.DeallocateIPs(allocateReq)
	}

	confBytesArray := generateConfigBytesArray(n, ipResponses, simplePluginFlag, mainPlugin, chainedPlugins)
	if len(confBytesArray) == 0 {
		return &pb.GenerateResponse{Success: false, ConfList: confBytesArray, Message: "no config is generated."}, nil
	}
	return &pb.GenerateResponse{Success: true, ConfList: confBytesArray, Message: "succeeded"}, nil
}

func NewGeneratorServer(port int) (*grpc.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, listener, err
	}
	s := grpc.NewServer()
	pb.RegisterGeneratorServer(s, &ImplementedGeneratorServer{})

	return s, listener, nil
}
