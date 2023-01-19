/*
 * Copyright 2022- IBM Inc. All rights reserved
 * SPDX-License-Identifier: Apache2.0
 */

package controllers

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	multinicv1 "github.com/foundation-model-stack/multi-nic-cni/api/v1"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	RouteMessage map[multinicv1.RouteStatus]string = map[multinicv1.RouteStatus]string{
		multinicv1.SomeRouteFailed: "some route cannot be applied, need attention",
		multinicv1.RouteUnknown:    "some daemon cannot be connected",
		multinicv1.ApplyingRoute:   "waiting for route update",
		multinicv1.AllRouteApplied: "",
	}
)

// MultiNicNetworkHandler handles MultiNicNetwork object
// - update MultiNicNetwork status according to CIDR results
type MultiNicNetworkHandler struct {
	client.Client
	Log logr.Logger
}

func (h *MultiNicNetworkHandler) GetNetwork(name string) (*multinicv1.MultiNicNetwork, error) {
	instance := &multinicv1.MultiNicNetwork{}
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: metav1.NamespaceAll,
	}
	err := h.Client.Get(context.TODO(), namespacedName, instance)
	return instance, err
}

func (h *MultiNicNetworkHandler) SyncAllStatus(name string, spec multinicv1.CIDRSpec, routeStatus multinicv1.RouteStatus, daemonSize, infoAvailableSize int, cidrChange bool) error {
	instance, err := h.GetNetwork(name)
	if err != nil {
		return err
	}
	discoverStatus := instance.Status.DiscoverStatus
	netConfigStatus := instance.Status.NetConfigStatus
	message := instance.Status.Message
	if routeStatus == multinicv1.SomeRouteFailed || routeStatus == multinicv1.ApplyingRoute {
		netConfigStatus = multinicv1.WaitForConfig
	} else if routeStatus == multinicv1.AllRouteApplied {
		netConfigStatus = multinicv1.ConfigComplete
	}

	if routeErrMsg, found := RouteMessage[routeStatus]; found {
		message = routeErrMsg
	}

	discoverStatus = multinicv1.DiscoverStatus{
		ExistDaemon:            daemonSize,
		InterfaceInfoAvailable: infoAvailableSize,
		CIDRProcessedHost:      discoverStatus.CIDRProcessedHost,
	}

	err = h.updateStatus(instance, spec, routeStatus, discoverStatus, netConfigStatus, message, cidrChange)
	return err
}

func (h *MultiNicNetworkHandler) updateStatus(instance *multinicv1.MultiNicNetwork, spec multinicv1.CIDRSpec, status multinicv1.RouteStatus, discoverStatus multinicv1.DiscoverStatus, netConfigStatus multinicv1.NetConfigStatus, message string, cidrChange bool) error {
	results := []multinicv1.NicNetworkResult{}

	if cidrChange {
		maxNumOfHost := 0
		for _, entry := range spec.CIDRs {
			numOfHost := len(entry.Hosts)
			result := multinicv1.NicNetworkResult{
				NetAddress: entry.NetAddress,
				NumOfHost:  numOfHost,
			}
			if numOfHost > maxNumOfHost {
				maxNumOfHost = numOfHost
			}
			results = append(results, result)
		}
		discoverStatus.CIDRProcessedHost = maxNumOfHost
	}

	instance.Status = multinicv1.MultiNicNetworkStatus{
		ComputeResults:  results,
		LastSyncTime:    metav1.Now(),
		DiscoverStatus:  discoverStatus,
		NetConfigStatus: netConfigStatus,
		Message:         message,
		RouteStatus:     status,
	}
	err := h.Client.Status().Update(context.Background(), instance)
	if err != nil {
		h.Log.V(2).Info(fmt.Sprintf("Failed to update %s status: %v", instance.Name, err))
	}
	return err
}

func (h *MultiNicNetworkHandler) UpdateNetConfigStatus(instance *multinicv1.MultiNicNetwork, netConfigStatus multinicv1.NetConfigStatus, message string) error {
	if message != "" {
		instance.Status.Message = message
	}
	if instance.Status.ComputeResults == nil {
		instance.Status.ComputeResults = []multinicv1.NicNetworkResult{}
	}
	instance.Status.LastSyncTime = metav1.Now()
	instance.Status.NetConfigStatus = netConfigStatus
	err := h.Client.Status().Update(context.Background(), instance)
	if err != nil {
		h.Log.V(2).Info(fmt.Sprintf("Failed to update %s network status: %v", instance.Name, err))
	}
	return err
}
