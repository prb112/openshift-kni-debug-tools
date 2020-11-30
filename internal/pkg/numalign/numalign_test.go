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
	"testing"

	"github.com/openshift-kni/debug-tools/internal/pkg/vfs"
)

func TestResources(t *testing.T) {
	type testCase struct {
		name     string
		env      []string
		glob     map[string]vfs.GlobResult
		content  map[string]vfs.ReadFileResult
		expected Result
	}

	testCases := []testCase{
		{
			name: "single node",
			env:  []string{"PCIDEVICE_IO_OPENSHIFT_KNI_CARD=0000:00:1f.0"},
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
				"/sys/bus/pci/devices/0000:00:1f.0/numa_node": vfs.ReadFileResult{
					Data: []byte("0"),
				},
				"/proc/self/status": vfs.ReadFileResult{
					Data: []byte(fullStatus),
				},
			},
			expected: Result{
				Aligned:    true,
				NUMACellID: 0,
			},
		},
		{
			name: "dual node",
			env:  []string{"PCIDEVICE_IO_OPENSHIFT_KNI_CARD=0000:00:1f.0"},
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
				"/sys/bus/pci/devices/0000:00:1f.0/numa_node": vfs.ReadFileResult{
					Data: []byte("0"),
				},
				"/proc/self/status": vfs.ReadFileResult{
					Data: []byte(fullStatus),
				},
			},
			expected: Result{
				Aligned:    true,
				NUMACellID: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs := vfs.FakeFS{
				FileContents: tc.content,
				GlobResults:  tc.glob,
			}
			numaRes, err := NewResources(fs, "/proc", "/sys", tc.env, []string{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			res := numaRes.CheckAlignment()
			if res.Aligned != tc.expected.Aligned {
				t.Errorf("alignment mismatch: got %v expected %v", res.Aligned, tc.expected.Aligned)
			}
			if res.NUMACellID != tc.expected.NUMACellID {
				t.Errorf("NUMA cell ID mismatch: got %v expected %v", res.NUMACellID, tc.expected.NUMACellID)
			}
		})
	}
}
