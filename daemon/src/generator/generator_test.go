/*
 * Copyright 2022- IBM Inc. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */

package generator

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/foundation-model-stack/multi-nic-cni/daemon/allocator"
	pb "github.com/foundation-model-stack/multi-nic-cni/daemon/generator/proto"
)

func TestGeneratedConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "config generator")
}

func getMultiNicNetConfig(mainPlugin string) string {
	return fmt.Sprintf(`{
		"cniVersion": "0.3.0",
		"name": "multi-nic-cni-operator-ipvlanl3",
		"type": "multi-nic",
		"ipam": {"hostBlock":8,"interfaceBlock":2,"type":"multi-nic-ipam","vlanMode":"l3"},
		"dns": {},
		"plugin": %s,
		"subnet": "192.168.0.0/16",
		"masterNets":["10.241.132.0/24","10.241.133.0/24"],
		"multiNICIPAM":true,
		"daemonIP": "",
		"daemonPort":11000
		}`, mainPlugin)
}

const (
	dummyMainPlugin    = "ipvlan"
	dummyChainedPlugin = "tuning"
	dummyIPAMPlugin    = "static"
)

var (
	dummyMasters   = []string{"ens4", "ens5"}
	dummyDeviceIDs = []string{"dev1", "dev2"}
	dummyIPs       = []string{"x.x.x.x", "y.y.y.y"}
	dummyVLANBlock = "24"
)

var simpleMultiNicNetConfig string = getMultiNicNetConfig(fmt.Sprintf(`{"cniVersion":"0.3.0","mode":"l3","type":"%s"}`, dummyMainPlugin))
var chainedMultiNicNetConfig string = getMultiNicNetConfig(fmt.Sprintf(`{"cniVersion":"0.3.0", "type": "", "plugins": [{"cniVersion":"0.3.0", "mode":"l3","type":"%s"}, {"cniVersion":"0.3.0", "type":"%s"}]}`, dummyMainPlugin, dummyChainedPlugin))

var _ = Describe("Test GenerateConfigBytesArray", func() {
	It("generate with simple config", func() {
		configRequest := getConfigRequest(simpleMultiNicNetConfig)
		configBytesArray := testGenerateFunc(configRequest)
		Expect(len(configBytesArray)).To(Equal(len(dummyMasters)))
		for index, masterName := range dummyMasters {
			confBytes := configBytesArray[index]
			fmt.Printf("%s\n", string(confBytes))
			ipAddr := fmt.Sprintf("%s/%s", dummyIPs[index], dummyVLANBlock)
			var mainPlugin map[string]interface{}
			err := json.Unmarshal(confBytes, &mainPlugin)
			Expect(err).NotTo(HaveOccurred())
			testMainPlugin(mainPlugin, masterName, ipAddr)
		}
	})
	It("generate with chained plugins", func() {
		configRequest := getConfigRequest(chainedMultiNicNetConfig)
		configBytesArray := testGenerateFunc(configRequest)
		Expect(len(configBytesArray)).To(Equal(len(dummyMasters)))
		for index, masterName := range dummyMasters {
			confBytes := configBytesArray[index]
			fmt.Printf("%s\n", string(confBytes))
			ipAddr := fmt.Sprintf("%s/%s", dummyIPs[index], dummyVLANBlock)
			var chainedConf map[string]interface{}
			err := json.Unmarshal(confBytes, &chainedConf)
			Expect(err).NotTo(HaveOccurred())
			pluginType := chainedConf["type"]
			Expect(pluginType).To(Equal(""))
			pluginInterface, ok := chainedConf["plugins"]
			Expect(ok).To(BeTrue())
			plugins := pluginInterface.([]interface{})
			Expect(len(plugins)).To(Equal(2))
			mainPlugin := plugins[0].(map[string]interface{})
			testMainPlugin(mainPlugin, masterName, ipAddr)
			chainedPlugin := plugins[1].(map[string]interface{})
			testChainedPlugin(chainedPlugin)
		}
	})
})

func testMainPlugin(conf map[string]interface{}, masterName, ipAddr string) {
	confType, ok := conf["type"]
	Expect(ok).To(BeTrue())
	Expect(confType).To(Equal(dummyMainPlugin))
	confMaster, ok := conf["master"]
	Expect(ok).To(BeTrue())
	Expect(confMaster).To(Equal(masterName))
	confIPAM, ok := conf["ipam"]
	Expect(ok).To(BeTrue())
	ipam := confIPAM.(map[string]interface{})
	ipamType, ok := ipam["type"]
	Expect(ok).To(BeTrue())
	Expect(ipamType).To(Equal("static"))
	ipamAddrs, ok := ipam["addresses"]
	Expect(ok).To(BeTrue())
	ipamAddrsList, ok := ipamAddrs.([]interface{})
	Expect(len(ipamAddrsList)).To(Equal(1))
	ipamAddrItem := ipamAddrsList[0].(map[string]interface{})
	ipamAddr, ok := ipamAddrItem["address"]
	Expect(ok).To(BeTrue())
	Expect(ipamAddr).To(Equal(ipAddr))
}

func testChainedPlugin(conf map[string]interface{}) {
	confType, ok := conf["type"]
	Expect(ok).To(BeTrue())
	Expect(confType).To(Equal(dummyChainedPlugin))
}

func getConfigRequest(data string) *pb.ConfigRequest {
	configRequest := &pb.ConfigRequest{
		Data: []byte(data),
	}
	return configRequest
}

func testLoadNetConf(configRequest *pb.ConfigRequest) (*NetConf, bool, map[string]interface{}, []interface{}) {
	n := &NetConf{}
	err := json.Unmarshal(configRequest.Data, n)
	Expect(err).NotTo(HaveOccurred())
	// instead of select call
	n.Masters = dummyMasters
	n.DeviceIDs = dummyDeviceIDs
	simplePluginFlag, mainPlugin, chainedPlugins, err := getPlugins(n)
	fmt.Printf("simplePluginFlag=%v, mainPlugin=%v, chainedPlugins=%v\n", simplePluginFlag, mainPlugin, chainedPlugins)
	Expect(err).NotTo(HaveOccurred())
	return n, simplePluginFlag, mainPlugin, chainedPlugins
}

func testGenerateFunc(configRequest *pb.ConfigRequest) [][]byte {
	n, simplePluginFlag, mainPlugin, chainedPlugins := testLoadNetConf(configRequest)
	ipResponses := make(map[string]allocator.IPResponse)
	for index, masterName := range dummyMasters {
		ipResponses[masterName] = allocator.IPResponse{
			InterfaceName: masterName,
			IPAddress:     dummyIPs[index],
			VLANBlockSize: dummyVLANBlock,
		}
	}
	return generateConfigBytesArray(n, ipResponses, simplePluginFlag, mainPlugin, chainedPlugins)
}
