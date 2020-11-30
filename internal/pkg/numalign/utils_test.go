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
	"reflect"
	"testing"

	"github.com/openshift-kni/debug-tools/internal/pkg/vfs"
)

func TestGetPCIDevicesFromEnv(t *testing.T) {
	type testCase struct {
		name     string
		env      []string
		expected []string
	}

	testCases := []testCase{
		{
			name: "empty",
		},
		{
			name: "not matching",
			env:  []string{"FOO", "BAR=BAZ"},
		},
		{
			name: "malformed",
			env:  []string{"PCIDEVICEMALFORMEDNAME"},
		},
		{
			name:     "single device",
			env:      []string{"PCIDEVICE_IO_OPENSHIFT_KNI_CARD=0000:00:1f.0"},
			expected: []string{"0000:00:1f.0"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			devs := GetPCIDevicesFromEnv(tc.env)
			if !reflect.DeepEqual(devs, tc.expected) {
				t.Errorf("got %#v expected %#v", devs, tc.expected)
			}
		})
	}

}

func TestGetPCIDeviceToNumaNodeMap(t *testing.T) {
	type testCase struct {
		name        string
		content     map[string]vfs.ReadFileResult
		pciDevs     []string
		expectedMap map[string]int
	}

	testCases := []testCase{
		{
			name: "no device",
			content: map[string]vfs.ReadFileResult{
				"/sys/bus/pci/devices/0000:00:1f.0/numa_node": vfs.ReadFileResult{
					Data: []byte("0"),
				},
			},
			pciDevs:     []string{},
			expectedMap: map[string]int{},
		},
		{
			name: "single device",
			content: map[string]vfs.ReadFileResult{
				"/sys/bus/pci/devices/0000:00:1f.0/numa_node": vfs.ReadFileResult{
					Data: []byte("0"),
				},
			},
			pciDevs: []string{"0000:00:1f.0"},
			expectedMap: map[string]int{
				"0000:00:1f.0": 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := vfs.FakeFS{
				FileContents: tc.content,
			}
			devMap, err := GetPCIDeviceToNumaNodeMap(fs, "/sys/bus/pci/devices", tc.pciDevs)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(devMap, tc.expectedMap) {
				t.Errorf("got %#v expected %#v", devMap, tc.expectedMap)
			}
		})
	}
}

func TestGetCPUsPerNUMANode(t *testing.T) {
	type testCase struct {
		name         string
		glob         map[string]vfs.GlobResult
		content      map[string]vfs.ReadFileResult
		expectedCPUs map[int][]int
	}

	testCases := []testCase{
		{
			name: "single node",
			glob: map[string]vfs.GlobResult{
				"/sys/devices/system/node/node*": vfs.GlobResult{
					Matches: []string{
						"/sys/devices/system/node/node0",
					},
				},
			},
			content: map[string]vfs.ReadFileResult{
				"/sys/devices/system/node/node0/cpulist": vfs.ReadFileResult{
					Data: []byte("0-3"),
				},
			},
			expectedCPUs: map[int][]int{
				0: []int{0, 1, 2, 3},
			},
		},
		{
			name: "dual node",
			glob: map[string]vfs.GlobResult{
				"/sys/devices/system/node/node*": vfs.GlobResult{
					Matches: []string{
						"/sys/devices/system/node/node0",
						"/sys/devices/system/node/node1",
					},
				},
			},
			content: map[string]vfs.ReadFileResult{
				"/sys/devices/system/node/node0/cpulist": vfs.ReadFileResult{
					Data: []byte("0-3"),
				},
				"/sys/devices/system/node/node1/cpulist": vfs.ReadFileResult{
					Data: []byte("4-7"),
				},
			},
			expectedCPUs: map[int][]int{
				0: []int{0, 1, 2, 3},
				1: []int{4, 5, 6, 7},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := vfs.FakeFS{
				FileContents: tc.content,
				GlobResults:  tc.glob,
			}
			cpusPerNuma, err := GetCPUsPerNUMANode(fs, "/sys/devices/system/node")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(cpusPerNuma, tc.expectedCPUs) {
				t.Errorf("got %#v expected %#v", cpusPerNuma, tc.expectedCPUs)
			}
		})
	}
}

func TestGetCPUToNUMANodeMap(t *testing.T) {
	type testCase struct {
		name        string
		glob        map[string]vfs.GlobResult
		content     map[string]vfs.ReadFileResult
		allowedCPUs []int
		expectedMap map[int]int
	}

	testCases := []testCase{
		{
			name: "node aligned",
			glob: map[string]vfs.GlobResult{
				"/sys/devices/system/node/node*": vfs.GlobResult{
					Matches: []string{
						"/sys/devices/system/node/node0",
						"/sys/devices/system/node/node1",
					},
				},
			},
			content: map[string]vfs.ReadFileResult{
				"/sys/devices/system/node/node0/cpulist": vfs.ReadFileResult{
					Data: []byte("0-3"),
				},
				"/sys/devices/system/node/node1/cpulist": vfs.ReadFileResult{
					Data: []byte("4-7"),
				},
			},
			allowedCPUs: []int{1, 2},
			expectedMap: map[int]int{
				1: 0,
				2: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := vfs.FakeFS{
				FileContents: tc.content,
				GlobResults:  tc.glob,
			}
			cpuMap, err := GetCPUToNUMANodeMap(fs, "/sys/devices/system/node", tc.allowedCPUs)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(cpuMap, tc.expectedMap) {
				t.Errorf("got %#v expected %#v", cpuMap, tc.expectedMap)
			}
		})
	}
}

func TestGetAllowedCPUList(t *testing.T) {
	type testCase struct {
		name         string
		status       string
		expectedCPUs []int
		wantError    bool
	}
	testCases := []testCase{
		{"empty", "", nil, true},
		{"minimal", minimalStatus, []int{0, 1}, false},
		{"full", fullStatus, []int{0, 1, 2, 3}, false},
		{"gibberish", gibberishStatus, nil, true},
		{"malformed", malformedStatus, nil, true},
	}

	testCaseToReadFileResult := func(tc testCase) vfs.ReadFileResult {
		ret := vfs.ReadFileResult{
			Data: []byte(tc.status),
		}
		if tc.wantError {
			ret.Err = fmt.Errorf("fake generic error")
		}
		return ret
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filename := "/proc/self/status"
			fs := vfs.FakeFS{
				FileContents: map[string]vfs.ReadFileResult{
					filename: testCaseToReadFileResult(tc),
				},
			}
			cpus, err := GetAllowedCPUList(fs, filename)
			if err == nil && tc.wantError {
				t.Errorf("expected error, got none")
			}
			if err != nil && !tc.wantError {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(cpus, tc.expectedCPUs) {
				t.Errorf("got %#v expected %#v", cpus, tc.expectedCPUs)
			}
		})
	}
}

var gibberishStatus string = `Lorem ipsum dolor sit amet,
consectetur adipiscing elit,
Cpus: 23 sed do
eiusmod tempor`

var malformedStatus string = `Name:	cat
Cpus_allowed:	f
Cpus_allowed_list:	xxx`

var minimalStatus string = `Name:	cat
Cpus_allowed:	f
Cpus_allowed_list:	0,1`

var fullStatus string = `Name:	cat
Umask:	0002
State:	R (running)
Tgid:	8322
Ngid:	0
Pid:	8322
PPid:	6558
TracerPid:	0
Uid:	1000	1000	1000	1000
Gid:	1000	1000	1000	1000
FDSize:	256
Groups:	965 981 1000 1002 
NStgid:	8322
NSpid:	8322
NSpgid:	8322
NSsid:	6558
VmPeak:	  215440 kB
VmSize:	  215440 kB
VmLck:	       0 kB
VmPin:	       0 kB
VmHWM:	     524 kB
VmRSS:	     524 kB
RssAnon:	      72 kB
RssFile:	     452 kB
RssShmem:	       0 kB
VmData:	     316 kB
VmStk:	     136 kB
VmExe:	      20 kB
VmLib:	    1484 kB
VmPTE:	      72 kB
VmSwap:	       0 kB
HugetlbPages:	       0 kB
CoreDumping:	0
THP_enabled:	1
Threads:	1
SigQ:	0/63360
SigPnd:	0000000000000000
ShdPnd:	0000000000000000
SigBlk:	0000000000000000
SigIgn:	0000000000000000
SigCgt:	0000000000000000
CapInh:	0000000000000000
CapPrm:	0000000000000000
CapEff:	0000000000000000
CapBnd:	000001ffffffffff
CapAmb:	0000000000000000
NoNewPrivs:	0
Seccomp:	0
Seccomp_filters:	0
Speculation_Store_Bypass:	thread vulnerable
Cpus_allowed:	f
Cpus_allowed_list:	0-3
Mems_allowed:	00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000000,00000001
Mems_allowed_list:	0
voluntary_ctxt_switches:	1
nonvoluntary_ctxt_switches:	0`
