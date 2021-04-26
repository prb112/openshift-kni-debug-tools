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
 * Copyright 2021 Red Hat, Inc.
 */

package irqs_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/openshift-kni/debug-tools/pkg/irqs"
)

var nullLog = log.New(ioutil.Discard, "", 0)

func TestReadStats(t *testing.T) {
	rootDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("creating temp dir %v", err)
	}
	defer os.RemoveAll(rootDir) // clean up

	procDir := filepath.Join(rootDir, "proc")
	if err := os.Mkdir(procDir, 0755); err != nil {
		t.Fatalf("Mkdir(%s) failed: %v", procDir, err)
	}
	if err := ioutil.WriteFile(filepath.Join(procDir, "interrupts"), []byte(fakeInterrupts), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	ih := irqs.New(nullLog, procDir)
	counters, err := ih.ReadStats()
	if err != nil {
		t.Errorf("ReadStats(%s) failed: %v", procDir, err)
	}

	var irqTestCases = []struct {
		cpuIdx  int
		irqName string
		value   uint64
	}{
		// some random non-zero values from the fakeInterrupts below.
		// any non-zero value is fine, no special meaning.
		{0, "131", 3949116},
		{0, "LOC", 14926901},
		{1, "139", 21},
		{1, "LOC", 16283403},
		{2, "125", 12356620},
		{2, "LOC", 14699417},
		{3, "12", 713},
		{3, "LOC", 15519974},
		// now some zero values. Same criteria as above.
		{0, "120", 0},
		{1, "120", 0},
		{2, "120", 0},
		{3, "120", 0},
	}
	for _, tt := range irqTestCases {
		t.Run(fmt.Sprintf("cpu %d irq %q", tt.cpuIdx, tt.irqName), func(t *testing.T) {
			v := counters[tt.cpuIdx][tt.irqName]
			if v != tt.value {
				t.Errorf("Counters mismatch got %v expected %v", v, tt.value)
			}
		})
	}
}

const fakeInterrupts string = `            CPU0       CPU1       CPU2       CPU3       
   0:         13          0          0          0  IR-IO-APIC    2-edge      timer
   1:          0         21          0          0  IR-IO-APIC    1-edge      i8042
   8:          0          0          1          0  IR-IO-APIC    8-edge      rtc0
   9:          0       8564          0          0  IR-IO-APIC    9-fasteoi   acpi
  12:          0          0          0        713  IR-IO-APIC   12-edge      i8042
  16:          0          0        227          0  IR-IO-APIC   16-fasteoi   i801_smbus
 120:          0          0          0          0  DMAR-MSI    0-edge      dmar0
 121:          0          0          0          0  DMAR-MSI    1-edge      dmar1
 125:          0          0   12356620          0  IR-PCI-MSI 327680-edge      xhci_hcd
 126:          0          0          0         24  IR-PCI-MSI 31457280-edge      nvme0q0
 127:     107368          0          0          0  IR-PCI-MSI 31457281-edge      nvme0q1
 128:          0     107768          0          0  IR-PCI-MSI 31457282-edge      nvme0q2
 129:          0          0     108890          0  IR-PCI-MSI 31457283-edge      nvme0q3
 130:          0          0          0     100886  IR-PCI-MSI 31457284-edge      nvme0q4
 131:    3949116          0          0          0  IR-PCI-MSI 520192-edge      enp0s31f6
 132:          0    6707981          0          0  IR-PCI-MSI 32768-edge      i915
 133:          0          0          0         77  IR-PCI-MSI 360448-edge      mei_me
 134:         47          0          0          0  IR-PCI-MSI 30408704-edge      iwlwifi
 135:          0       1859          0          0  IR-PCI-MSI 514048-edge      snd_hda_intel:card1
 136:          0          0         21          0     dummy   44  rmi4_smbus
 137:          0          0          0          0      rmi4    0  rmi4-00.fn34
 138:          0          0          0          0      rmi4    1  rmi4-00.fn01
 139:          0         21          0          0      rmi4    2  rmi4-00.fn03
 140:          0          0          0          0      rmi4    3  rmi4-00.fn11
 141:          0          0          0          0      rmi4    4  rmi4-00.fn11
 142:          0          0          0          0      rmi4    5  rmi4-00.fn30
 NMI:        416        405        423        421   Non-maskable interrupts
 LOC:   14926901   16283403   14699417   15519974   Local timer interrupts
 SPU:          0          0          0          0   Spurious interrupts
 PMI:        416        405        423        421   Performance monitoring interrupts
 IWI:     131421    2522925     150274     139737   IRQ work interrupts
 RTR:          0          0          0          0   APIC ICR read retries
 RES:    2384009    1628159    2313879    1735030   Rescheduling interrupts
 CAL:    1910273    1713508    1870416    1758259   Function call interrupts
 TLB:    2231123    2225305    2305323    2335869   TLB shootdowns
 TRM:      39489      39489      39489      39489   Thermal event interrupts
 THR:          0          0          0          0   Threshold APIC interrupts
 DFR:          0          0          0          0   Deferred Error APIC interrupts
 MCE:          0          0          0          0   Machine check exceptions
 MCP:         61         62         62         62   Machine check polls
 ERR:          0
 MIS:          0
 PIN:          0          0          0          0   Posted-interrupt notification event
 NPI:          0          0          0          0   Nested posted-interrupt event
 PIW:          0          0          0          0   Posted-interrupt wakeup event`
