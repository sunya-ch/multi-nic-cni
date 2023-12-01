/*
 * Copyright 2022- IBM Inc. All rights reserved
 * SPDX-License-Identifier: Apache-2.0
 */

package ipam

import (
	"github.com/foundation-model-stack/multi-nic-cni/daemon/allocator"
)

type IPAMInterface interface {
	AllocateIPs(req allocator.IPRequest) map[string]allocator.IPResponse
	DeallocateIPs(req allocator.IPRequest) map[string]allocator.IPResponse
}
