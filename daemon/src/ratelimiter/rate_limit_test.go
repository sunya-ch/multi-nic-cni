/*
 * Copyright 2022- IBM Inc. All rights reserved
 * SPDX-License-Identifier: Apache2.0
 */

package ratelimiter

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vishvananda/netlink"
)

func TestRateLmiter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rate Limiter Test Suite")
}

var req = RateLimitRequest{
	srcIPs:    []string{"192.168.0.1", "192.168.1.1"},
	ratePerIP: 2048,
}

func getHostDevice() string {
	hostDeviceName := ""
	links, err := netlink.LinkList()
	Expect(err).To(BeNil())
	for _, link := range links {
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil || len(addrs) == 0 {
			continue
		}
		addr := addrs[0].IPNet
		if addr != nil {
			hostDeviceName = link.Attrs().Name
			break
		}
	}
	Expect(hostDeviceName).NotTo(BeEmpty())
	return hostDeviceName
}

var _ = Describe("Test Rate Limit", func() {
	It("Test rate path", func() {
		hostDevice := getHostDevice()
		req.hostDevice = hostDevice
		err := LimitRate(req)
		Expect(err).To(BeNil())
		defer clear(hostDevice)
		link, err := netlink.LinkByName(req.hostDevice)
		Expect(err).To(BeNil())
		qdiscList, err := netlink.QdiscList(link)
		Expect(err).To(BeNil())
		Expect(len(qdiscList)).To(BeNumerically(">", 0))
		found := false
		for _, qdisc := range qdiscList {
			if qdisc.Type() == "htb" {
				found = true
			}
		}
		Expect(found).To(BeTrue())
	})
})

func clear(hostDevice string) {
	req.hostDevice = hostDevice
	err := RemoveRateLimit(req)
	Expect(err).To(BeNil())
}
