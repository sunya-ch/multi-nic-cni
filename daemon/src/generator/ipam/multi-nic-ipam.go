/*
 * Copyright 2022- IBM Inc. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */

package ipam

import (
	"github.com/foundation-model-stack/multi-nic-cni/daemon/allocator"
)

const (
	MULTI_NIC_IPAM_TYPE = "multi-nic-ipam"
)

type MultiNicIPAM struct {
}

func (ip MultiNicIPAM) AllocateIPs(req allocator.IPRequest) map[string]allocator.IPResponse {
	responseMap := make(map[string]allocator.IPResponse)
	responseList := allocator.AllocateIP(req)
	for _, response := range responseList {
		responseMap[response.InterfaceName] = response
	}
	return responseMap
}

func (ip MultiNicIPAM) DeallocateIPs(req allocator.IPRequest) map[string]allocator.IPResponse {
	responseMap := make(map[string]allocator.IPResponse)
	responseList := allocator.DeallocateIP(req)
	for _, response := range responseList {
		responseMap[response.InterfaceName] = response
	}
	return responseMap
}
