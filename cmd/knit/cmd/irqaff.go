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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/openshift-kni/debug-tools/pkg/irqs"
	softirqs "github.com/openshift-kni/debug-tools/pkg/irqs/soft"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

type irqAffOptions struct {
	checkEffective  bool
	checkSoftirqs   bool
	showEmptySource bool
}

func newIRQAffinityCommand(knitOpts *knitOptions) *cobra.Command {
	opts := &irqAffOptions{}
	irqAff := &cobra.Command{
		Use:   "irqaff",
		Short: "show IRQ/softirq thread affinities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showIRQAffinity(cmd, knitOpts, opts, args)
		},
		Args: cobra.NoArgs,
	}
	irqAff.Flags().BoolVarP(&opts.checkEffective, "effective-affinity", "E", false, "check effective affinity.")
	irqAff.Flags().BoolVarP(&opts.checkSoftirqs, "softirqs", "s", false, "check softirqs counters.")
	irqAff.Flags().BoolVarP(&opts.showEmptySource, "show-empty-source", "e", false, "show infos if IRQ source is not reported.")
	return irqAff
}

func showIRQAffinity(cmd *cobra.Command, knitOpts *knitOptions, opts *irqAffOptions, args []string) error {
	if opts.checkSoftirqs {
		sh := softirqs.New(knitOpts.log, knitOpts.procFSRoot)
		info, err := sh.ReadInfo()

		if err != nil {
			return fmt.Errorf("error parsing softirqs from %q: %v", knitOpts.procFSRoot, err)
		}

		dumpSoftirqInfo(info, knitOpts.cpus)
	} else {
		ih := irqs.New(knitOpts.log, knitOpts.procFSRoot)

		flags := uint(0)
		if opts.checkEffective {
			flags |= irqs.EffectiveAffinity
		}

		irqInfos, err := ih.ReadInfo(flags)
		if err != nil {
			return fmt.Errorf("error parsing irqs from %q: %v", knitOpts.procFSRoot, err)
		}

		dumpIrqInfo(irqInfos, knitOpts.cpus, opts.showEmptySource)
	}

	return nil
}

// dumpIrqInfo displays on stdout the (sorted) list of IRQs, showing for each IRQ the
// IRQ source name and the (sorted) cpuset on which each IRQ may be served.
// note that IRQs without valid source aren't shown in /proc/cpuinfo. Hence add the showEmptySource bool to toggle them on/off
func dumpIrqInfo(infos []irqs.Info, cpus cpuset.CPUSet, showEmptySource bool) {
	for _, irqInfo := range infos {
		cpus := irqInfo.CPUs.Intersection(cpus)
		if cpus.Size() == 0 {
			continue
		}
		if irqInfo.Source == "" && !showEmptySource {
			continue
		}
		fmt.Printf("IRQ %3d [%24s]: can run on %v\n", irqInfo.IRQ, irqInfo.Source, cpus.String())
	}
}

// dumpSoftirqInfo displays on stdout the (sorted) list of softirqs, showing for each softirq the
// (sorted) cpuset on which each softirq was served in the past.
func dumpSoftirqInfo(info *softirqs.Info, cpus cpuset.CPUSet) {
	keys := softirqs.Names()
	for _, key := range keys {
		counters := info.Counters[key]
		cb := cpuset.NewBuilder()
		for idx, counter := range counters {
			if counter > 0 {
				cb.Add(idx)
			}
		}
		usedCPUs := cpus.Intersection(cb.Result())
		fmt.Printf("%8s = %s\n", key, usedCPUs.String())
	}
}
