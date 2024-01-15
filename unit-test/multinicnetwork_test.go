/*
 * Copyright 2022- IBM Inc. All rights reserved
 * SPDX-License-Identifier: Apache2.0
 */

package controllers

import (
	"fmt"

	multinicv1 "github.com/foundation-model-stack/multi-nic-cni/api/v1"
	"github.com/foundation-model-stack/multi-nic-cni/controllers"
	"github.com/foundation-model-stack/multi-nic-cni/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//+kubebuilder:scaffold:imports
)

var _ = Describe("Test deploying MultiNicNetwork", func() {
	cniVersion := "0.3.0"
	cniType := "ipvlan"
	mode := "l2"
	mtu := 1500
	cniArgs := make(map[string]string)
	cniArgs["mode"] = mode
	cniArgs["mtu"] = fmt.Sprintf("%d", mtu)

	multinicnetwork := getMultiNicCNINetwork("test-mn", cniVersion, cniType, cniArgs)

	It("successfully create/delete network attachment definition", func() {
		mainPlugin, annotations, err := multinicnetworkReconciler.GetMainPluginConf(multinicnetwork)
		Expect(err).NotTo(HaveOccurred())
		err = multinicnetworkReconciler.NetAttachDefHandler.CreateOrUpdate(multinicnetwork, mainPlugin, annotations)
		Expect(err).NotTo(HaveOccurred())
		err = multinicnetworkReconciler.NetAttachDefHandler.DeleteNets(multinicnetwork)
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("Test definition changes check", func() {
	cniVersion := "0.3.0"
	cniType := "ipvlan"
	mode := "l2"
	mtu := 1500
	cniArgs := make(map[string]string)
	cniArgs["mode"] = mode
	cniArgs["mtu"] = fmt.Sprintf("%d", mtu)

	multinicnetwork := getMultiNicCNINetwork("test-mn", cniVersion, cniType, cniArgs)
	It("detect no change", func() {
		mainPlugin, annotations, err := multinicnetworkReconciler.GetMainPluginConf(multinicnetwork)
		Expect(err).NotTo(HaveOccurred())
		def, err := plugin.NetToDef("", multinicnetwork, mainPlugin, annotations)
		Expect(err).NotTo(HaveOccurred())
		defCopy, err := plugin.NetToDef("", multinicnetwork, mainPlugin, annotations)
		Expect(err).NotTo(HaveOccurred())
		changed := plugin.CheckDefChanged(defCopy, def)
		Expect(changed).To(BeFalse())
	})

	It("detect annotation change", func() {
		mainPlugin, annotations, err := multinicnetworkReconciler.GetMainPluginConf(multinicnetwork)
		Expect(err).NotTo(HaveOccurred())
		def, err := plugin.NetToDef("", multinicnetwork, mainPlugin, annotations)
		Expect(err).NotTo(HaveOccurred())

		newAnnotations := map[string]string{"resource": "new"}
		defWithNewAnnotation, err := plugin.NetToDef("", multinicnetwork, mainPlugin, newAnnotations)
		Expect(err).NotTo(HaveOccurred())
		changed := plugin.CheckDefChanged(defWithNewAnnotation, def)
		Expect(changed).To(BeTrue())
	})

	It("detect config change", func() {
		mainPlugin, annotations, err := multinicnetworkReconciler.GetMainPluginConf(multinicnetwork)
		Expect(err).NotTo(HaveOccurred())
		def, err := plugin.NetToDef("", multinicnetwork, mainPlugin, annotations)
		Expect(err).NotTo(HaveOccurred())

		newCniArgs := make(map[string]string)
		newCniArgs["mode"] = "l3"
		newCniArgs["mtu"] = fmt.Sprintf("%d", mtu)
		changedArgsNetwork := getMultiNicCNINetwork("test-mn", cniVersion, cniType, newCniArgs)
		newMainPlugin, annotations, err := multinicnetworkReconciler.GetMainPluginConf(changedArgsNetwork)
		Expect(err).NotTo(HaveOccurred())
		defWithNewArgs, err := plugin.NetToDef("", changedArgsNetwork, newMainPlugin, annotations)
		Expect(err).NotTo(HaveOccurred())
		changed := plugin.CheckDefChanged(defWithNewArgs, def)
		Expect(changed).To(BeTrue())
	})
})

func getNetStatus(computeResults []multinicv1.NicNetworkResult, discoverStatus multinicv1.DiscoverStatus, netConfigStatus multinicv1.NetConfigStatus, routeStatus multinicv1.RouteStatus) multinicv1.MultiNicNetworkStatus {
	return multinicv1.MultiNicNetworkStatus{
		ComputeResults:  computeResults,
		DiscoverStatus:  discoverStatus,
		NetConfigStatus: netConfigStatus,
		RouteStatus:     routeStatus,
		Message:         "",
		LastSyncTime:    metav1.Now(),
	}
}

func testNewNetStatus(multinicnetwork *multinicv1.MultiNicNetwork, newStatus multinicv1.MultiNicNetworkStatus, expectedChange bool) *multinicv1.MultiNicNetwork {
	if expectedChange {
		updated := controllers.NetStatusUpdated(multinicnetwork, newStatus)
		// check new update
		Expect(updated).To(Equal(expectedChange))
		// update status
		multinicnetwork.Status = newStatus
	}
	updated := controllers.NetStatusUpdated(multinicnetwork, newStatus)
	// expect no update
	Expect(updated).To(BeFalse())
	return multinicnetwork
}

var _ = Describe("Test multinicnetwork status change check", func() {
	cniVersion := "0.3.0"
	cniType := "ipvlan"
	mode := "l3"
	mtu := 1500
	cniArgs := make(map[string]string)
	cniArgs["mode"] = mode
	cniArgs["mtu"] = fmt.Sprintf("%d", mtu)

	It("detect change from no status", func() {
		multinicnetwork := getMultiNicCNINetwork("test-mn", cniVersion, cniType, cniArgs)
		initStatus := getNetStatus([]multinicv1.NicNetworkResult{}, multinicv1.DiscoverStatus{}, multinicv1.WaitForConfig, multinicv1.ApplyingRoute)
		updated := controllers.NetStatusUpdated(multinicnetwork, initStatus)
		Expect(updated).To(BeTrue())
	})

	It("detect change on compute results", func() {
		multinicnetwork := getMultiNicCNINetwork("test-mn", cniVersion, cniType, cniArgs)
		multinicnetwork.Status = getNetStatus([]multinicv1.NicNetworkResult{}, multinicv1.DiscoverStatus{}, multinicv1.WaitForConfig, multinicv1.ApplyingRoute)

		net1 := multinicv1.NicNetworkResult{
			NetAddress: "192.168.0.0/24",
			NumOfHost:  1,
		}
		net2 := multinicv1.NicNetworkResult{
			NetAddress: "192.168.1.0/24",
			NumOfHost:  2,
		}

		computeResults := []multinicv1.NicNetworkResult{net1}
		newStatus := getNetStatus(computeResults, multinicv1.DiscoverStatus{}, multinicv1.WaitForConfig, multinicv1.ApplyingRoute)
		expectedChange := true
		multinicnetwork = testNewNetStatus(multinicnetwork, newStatus, expectedChange)

		// add new compute result
		computeResults = []multinicv1.NicNetworkResult{net1, net2}
		newStatus = getNetStatus(computeResults, multinicv1.DiscoverStatus{}, multinicv1.WaitForConfig, multinicv1.ApplyingRoute)
		expectedChange = true
		multinicnetwork = testNewNetStatus(multinicnetwork, newStatus, expectedChange)

		// change order
		computeResults = []multinicv1.NicNetworkResult{net2, net1}
		newStatus = getNetStatus(computeResults, multinicv1.DiscoverStatus{}, multinicv1.WaitForConfig, multinicv1.ApplyingRoute)
		expectedChange = false
		multinicnetwork = testNewNetStatus(multinicnetwork, newStatus, expectedChange)

		// change values
		net1.NetAddress = "192.168.0.2/24"
		computeResults = []multinicv1.NicNetworkResult{net1, net2}
		newStatus = getNetStatus(computeResults, multinicv1.DiscoverStatus{}, multinicv1.WaitForConfig, multinicv1.ApplyingRoute)
		expectedChange = true
		testNewNetStatus(multinicnetwork, newStatus, expectedChange)
		net1.NumOfHost = 3
		computeResults = []multinicv1.NicNetworkResult{net1, net2}
		newStatus = getNetStatus(computeResults, multinicv1.DiscoverStatus{}, multinicv1.WaitForConfig, multinicv1.ApplyingRoute)
		expectedChange = true
		testNewNetStatus(multinicnetwork, newStatus, expectedChange)
	})

	It("detect change on simple status", func() {
		multinicnetwork := getMultiNicCNINetwork("test-mn", cniVersion, cniType, cniArgs)
		multinicnetwork.Status = getNetStatus([]multinicv1.NicNetworkResult{}, multinicv1.DiscoverStatus{}, multinicv1.WaitForConfig, multinicv1.ApplyingRoute)

		// change discover status
		discoverStatus := multinicv1.DiscoverStatus{
			ExistDaemon: 10,
		}
		newStatus := getNetStatus([]multinicv1.NicNetworkResult{}, discoverStatus, multinicv1.WaitForConfig, multinicv1.ApplyingRoute)
		expectedChange := true
		testNewNetStatus(multinicnetwork, newStatus, expectedChange)

		// change config status
		newStatus = getNetStatus([]multinicv1.NicNetworkResult{}, discoverStatus, multinicv1.ConfigComplete, multinicv1.ApplyingRoute)
		expectedChange = true
		testNewNetStatus(multinicnetwork, newStatus, expectedChange)

		// change route status
		newStatus = getNetStatus([]multinicv1.NicNetworkResult{}, discoverStatus, multinicv1.ConfigComplete, multinicv1.AllRouteApplied)
		expectedChange = true
		testNewNetStatus(multinicnetwork, newStatus, expectedChange)
	})
})
