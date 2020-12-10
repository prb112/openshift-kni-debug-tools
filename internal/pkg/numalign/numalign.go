/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2020 Red Hat, Inc.
 */

package numalign

import (
	"encoding/json"
	"log"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/openshift-kni/debug-tools/internal/pkg/vfs"
)

const (
	SysDevicesSystemNodeDir = "devices/system/node"
	SysBusPCIDevicesDir     = "bus/pci/devices/"
)

type Resources struct {
	CPUToNUMANode     map[int]int    `json:"cpus"`
	PCIDevsToNUMANode map[string]int `json:"pcidevices"`
}

type Result struct {
	Aligned    bool       `json:"aligned"`
	NUMACellID int        `json:"numacellid"`
	Resources  *Resources `json:"resources",omitempty`
}

func (re Result) JSON() string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.Encode(re)
	return b.String()
}

func (numaRes *Resources) CheckAlignment() Result {
	ret := Result{
		Aligned:    false,
		NUMACellID: -1,
		Resources:  numaRes,
	}
	for _, cpuNode := range numaRes.CPUToNUMANode {
		if ret.NUMACellID == -1 {
			ret.NUMACellID = cpuNode
		} else if ret.NUMACellID != cpuNode {
			return ret
		}
	}
	for _, devNode := range numaRes.PCIDevsToNUMANode {
		if devNode != -1 && ret.NUMACellID != devNode {
			return ret
		}
	}
	ret.Aligned = true
	return ret
}

func (numaRes *Resources) JSON() string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.Encode(numaRes)
	return b.String()
}

func NewResources(fs vfs.VFS, procfsRoot, sysfsRoot string, environ, pids []string) (*Resources, error) {
	var err error

	pciDevs := GetPCIDevicesFromEnv(environ)

	var pidStrings []string
	if len(pids) > 1 {
		pidStrings = append(pidStrings, pids...)
	} else {
		pidStrings = append(pidStrings, "self")
	}

	var refCpuIDs []int
	refCpuIDs, err = GetAllowedCPUList(fs, filepath.Join(procfsRoot, pidStrings[0], "status"))
	if err != nil {
		return nil, err
	}
	log.Printf("CPU: allowed for %q: %v", pidStrings[0], refCpuIDs)

	for _, pidString := range pidStrings[1:] {
		cpuIDs, err := GetAllowedCPUList(fs, filepath.Join(procfsRoot, pidString, "status"))
		if err != nil {
			return nil, err
		}
		log.Printf("CPU: allowed for %q: %v", pidString, cpuIDs)

		if !reflect.DeepEqual(refCpuIDs, cpuIDs) {
			log.Fatalf("CPU: allowed set differs pid %q (%v) pid %q (%v)", pidStrings[0], refCpuIDs, pidString, cpuIDs)
		}
	}

	CPUToNUMANode, err := GetCPUToNUMANodeMap(fs, filepath.Join(sysfsRoot, SysDevicesSystemNodeDir), refCpuIDs)
	if err != nil {
		return nil, err
	}

	NUMAPerDev, err := GetPCIDeviceToNumaNodeMap(fs, filepath.Join(sysfsRoot, SysBusPCIDevicesDir), pciDevs)
	if err != nil {
		return nil, err
	}

	return &Resources{
		CPUToNUMANode:     CPUToNUMANode,
		PCIDevsToNUMANode: NUMAPerDev,
	}, nil

}
