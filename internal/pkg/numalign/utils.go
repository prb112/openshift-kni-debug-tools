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
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/openshift-kni/debug-tools/internal/pkg/vfs"

	cpuset "github.com/openshift-kni/debug-tools/pkg/k8s_imported"
)

func splitCPUList(cpuList string) ([]int, error) {
	cpus, err := cpuset.Parse(cpuList)
	if err != nil {
		return nil, err
	}
	return cpus.ToSlice(), nil
}

func GetAllowedCPUList(fs vfs.VFS, statusFile string) ([]int, error) {
	var cpuIDs []int
	var err error
	content, err := fs.ReadFile(statusFile)
	if err != nil {
		return cpuIDs, err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Cpus_allowed_list") {
			pair := strings.SplitN(line, ":", 2)
			return splitCPUList(strings.TrimSpace(pair[1]))
		}
	}
	return cpuIDs, fmt.Errorf("malformed status file: %s", statusFile)
}

func GetCPUToNUMANodeMap(fs vfs.VFS, sysNodeDir string, cpuIDs []int) (map[int]int, error) {
	cpusPerNUMA, err := GetCPUsPerNUMANode(fs, sysNodeDir)
	if err != nil {
		return nil, err
	}
	CPUToNUMANode := MakeCPUsToNUMANodeMap(cpusPerNUMA)

	// filter out only the allowed CPUs
	CPUMap := make(map[int]int)
	for _, cpuID := range cpuIDs {
		_, ok := CPUToNUMANode[cpuID]
		if !ok {
			return nil, fmt.Errorf("CPU %d not found on NUMA map: %v", cpuID, CPUToNUMANode)
		}
		CPUMap[cpuID] = CPUToNUMANode[cpuID]
	}
	return CPUMap, nil
}

func GetPCIDeviceToNumaNodeMap(fs vfs.VFS, sysBusPCIDir string, pciDevs []string) (map[string]int, error) {
	if len(pciDevs) == 0 {
		log.Printf("PCI: devices: none found - SKIP")
		return make(map[string]int), nil
	}
	log.Printf("PCI: devices: detected  %s", strings.Join(pciDevs, " - "))

	NUMAPerDev, err := GetPCIDeviceNUMANode(fs, sysBusPCIDir, pciDevs)
	if err != nil {
		return nil, err
	}
	return NUMAPerDev, nil
}

func GetPCIDeviceNUMANode(fs vfs.VFS, sysPCIDir string, devs []string) (map[string]int, error) {
	NUMAPerDev := make(map[string]int)
	for _, dev := range devs {
		content, err := fs.ReadFile(filepath.Join(sysPCIDir, dev, "numa_node"))
		if err != nil {
			return nil, err
		}
		numacellID, err := strconv.Atoi(strings.TrimSpace(string(content)))
		if err != nil {
			return nil, err
		}
		NUMAPerDev[dev] = numacellID
	}
	return NUMAPerDev, nil
}

func GetCPUsPerNUMANode(fs vfs.VFS, sysfsdir string) (map[int][]int, error) {
	pattern := filepath.Join(sysfsdir, "node*")
	nodes, err := fs.Glob(pattern)
	if err != nil {
		return nil, err
	}
	cpusPerNUMA := make(map[int][]int)
	for _, node := range nodes {
		_, nodeID := filepath.Split(node)
		numacellID, err := strconv.Atoi(strings.TrimSpace(nodeID[4:]))
		if err != nil {
			return nil, err
		}
		content, err := fs.ReadFile(filepath.Join(node, "cpulist"))
		if err != nil {
			return nil, err
		}
		cpuSet, err := splitCPUList(strings.TrimSpace(string(content)))
		if err != nil {
			return nil, err
		}
		cpusPerNUMA[numacellID] = cpuSet
	}
	return cpusPerNUMA, nil
}

func GetPCIDevicesFromEnv(environ []string) []string {
	var pciDevs []string
	for _, envVar := range environ {
		if !strings.HasPrefix(envVar, "PCIDEVICE_") {
			continue
		}
		pair := strings.SplitN(envVar, "=", 2)
		pciDevs = append(pciDevs, pair[1])
	}
	return pciDevs
}

func MakeCPUsToNUMANodeMap(cpusPerNUMA map[int][]int) map[int]int {
	CPUToNUMANode := make(map[int]int)
	for numacellID, cpuSet := range cpusPerNUMA {
		for _, cpu := range cpuSet {
			CPUToNUMANode[cpu] = numacellID
		}
	}
	return CPUToNUMANode
}
