/*
 * Copyright 2022- IBM Inc. All rights reserved
 * SPDX-License-Identifier: Apache2.0
 */

package main

import (
	"context"
	"github.com/containernetworking/cni/pkg/invoke"
	"github.com/containernetworking/cni/pkg/skel"
	"os"
	"path/filepath"

	"github.com/containernetworking/cni/pkg/types"
)

var defaultExec = &invoke.DefaultExec{
	RawExec: &invoke.RawExec{Stderr: os.Stderr},
}

func execPlugin(version string, plugin string, command string, confBytes []byte, args *skel.CmdArgs, ifName string, isAdd bool) (types.Result, error) {
	cniPath := os.Getenv("CNI_PATH")
	singleNicArgs := &invoke.Args{
		Command:       command,
		ContainerID:   args.ContainerID,
		NetNS:         args.Netns,
		IfName:        ifName,
		PluginArgsStr: args.Args,
		Path:          cniPath,
	}
	paths := filepath.SplitList(cniPath)
	pluginPath, err := defaultExec.FindInPath(plugin, paths)
	if err != nil {
		return nil, err
	}

	if isAdd {
		r, err := invoke.ExecPluginWithResult(context.TODO(), pluginPath, confBytes, singleNicArgs, defaultExec)
		if err != nil {
			return nil, err
		}
		return newResultFromResult(version, r)
	} else {
		err = invoke.ExecPluginWithoutResult(context.TODO(), pluginPath, confBytes, singleNicArgs, defaultExec)
		return nil, err
	}
}
