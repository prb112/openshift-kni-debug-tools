package e2eknit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
	"github.com/openshift-kni/debug-tools/pkg/irqs"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

var _ = g.Describe("knit IRQ watch tests", func() {

	var fixtureName = "xeon-multinuma-00"
	var snapshotRoot string

	type cmdOutput struct {
		out []byte
		err error
	}

	g.Context("Without isolated CPUs", func() {
		g.It("Produces the expected delta between consecutive irq samples", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "knit"),
				"-P", filepath.Join(snapshotRoot, "proc"),
				"-S", filepath.Join(snapshotRoot, "sys"),
				"-W", "3s",
				"-T", "1",
				"-J",
				"irqwatch",
			}
			fmt.Fprintf(g.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter

			c := make(chan cmdOutput)
			go func(c chan cmdOutput) {
				out, err := cmd.Output()
				c <- cmdOutput{out: out,
					err: err,
				}
			}(c)

			// while the irqwatch is running change the interrupts file and return the changes
			// this is very unlikley to suffer from race condition, since the irqwatch will wait 3 full seconds between each sample,
			// so modifyInterruptsFile function will finish it's job already
			delta, err := modifyInterruptsFile(filepath.Join(snapshotRoot, "proc"))
			o.Expect(err).ToNot(o.HaveOccurred())

			res := <-c
			o.Expect(res.err).ToNot(o.HaveOccurred())

			counters, err := extractCountersFromIRQWatchOutput(res.out)
			o.Expect(res.err).ToNot(o.HaveOccurred())

			ok := areCountersEqual(delta, counters)
			o.Expect(ok).To(o.BeTrue())
		})
	})

	g.BeforeEach(func() {
		snapshotRoot = snapshotBeforeEach(fixtureName, "sysinfo.tgz")
	})

	g.AfterEach(func() {
		snapshotAfterEach(snapshotRoot)
	})
})

func modifyInterruptsFile(procFs string) (irqs.Stats, error) {
	mockLog := log.New(ioutil.Discard, "", 0)

	ih := irqs.New(mockLog, procFs)
	stats, err := ih.ReadStats()
	if err != nil {
		return nil, err
	}
	initialStats := stats.Clone()
	delta := make(irqs.Stats)
	rand.Seed(time.Now().UnixNano())

	// 50 random changes are good enough
	const numOfModifications int = 50
	// just a number that is big enough to make nice distribution of interrupts on random queues
	const maxInterrupts = 1000

	for i := 0; i < numOfModifications; i++ {
		numOfCpus := len(stats)

		// select random cpuid to change its counters
		cpuid := int(rand.Intn(numOfCpus))

		// select random IRQ in the seleced cpuid
		k := selRandKey(stats[cpuid])

		stats[cpuid][k] += uint64(rand.Intn(maxInterrupts))
		delta[cpuid] = make(irqs.Counter)
		delta[cpuid][k] = stats[cpuid][k] - initialStats[cpuid][k]
	}

	// add the random new stat values to the interruts file
	return delta, reWriteInterrupts(procFs, stats)
}

func reWriteInterrupts(procFs string, stats irqs.Stats) error {
	tmpInterrruptsFile := filepath.Join(procFs, "interrupts_temp")
	interrruptsFile := filepath.Join(procFs, "interrupts")

	f, err := os.Create(tmpInterrruptsFile)
	if err != nil {
		g.Fail(fmt.Sprintf("fail to create temp interrupts file %q", tmpInterrruptsFile))
		return err
	}

	defer f.Close()

	// Since it's a test we can assume no offlined cpu, so no holes in the slice (i.e. cpu_0 to cpu_n are presents)
	type cpuidToIrqValue []uint64

	// IRQ name maps to counter values for each cpuid
	// this way is makes it easier to be wrriten back into a file, line by line
	irqMap := make(map[string]cpuidToIrqValue)

	cpuids := getSortedCPUids(stats)

	for cpuid := range cpuids {
		for k, v := range stats[cpuid] {
			irqMap[k] = append(irqMap[k], v)
		}
	}

	tab(f)
	// write cpu ids on the first line
	for cpuid := range cpuids {
		cpuName := fmt.Sprintf("%11s", fmt.Sprintf("CPU%d", cpuid))
		f.WriteString(cpuName)
	}
	newLine(f)

	for irqName, counters := range irqMap {
		f.WriteString(fmt.Sprintf("%4s:", irqName))

		for _, v := range counters {
			f.WriteString(fmt.Sprintf("%11d", v))
		}
		newLine(f)
	}

	// use this approach in order to avoid from other processes to access
	// the original interrupts file while modifing it
	os.Rename(tmpInterrruptsFile, interrruptsFile)
	if err != nil {
		g.Fail(fmt.Sprintf("fail to rename file %q to %q", tmpInterrruptsFile, interrruptsFile))
		return err
	}

	return nil
}

func tab(f *os.File) {
	f.WriteString("\t")
}

func newLine(f *os.File) {
	f.WriteString("\n")
}

func selRandKey(c irqs.Counter) string {
	//fmt.Println(len(c))
	i := rand.Intn(len(c))
	for k := range c {
		if i == 0 {
			return k
		}
	}
	i--
	// default key, never expect to get here
	return "0"
}

func getSortedCPUids(s irqs.Stats) []int {
	keys := make([]int, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func extractCountersFromIRQWatchOutput(b []byte) (irqs.Stats, error) {
	type irqSummary struct {
		Elapsed  time.Duration `json:"elapsed"`
		Counters irqs.Stats    `json:"counters"`
	}

	irqS := irqSummary{}
	if err := json.Unmarshal(b, &irqS); err != nil {
		return nil, fmt.Errorf("Error unmarshalling b: %v", err)
	}

	return irqS.Counters, nil
}

func stats2JSON(s irqs.Stats) ([]byte, error) {
	cpus := cpuset.CPUSet{}
	cnt := struct {
		Counters irqs.Stats
	}{
		Counters: s.ForCPUs(cpus),
	}
	return json.Marshal(cnt)
}

func areCountersEqual(c1, c2 irqs.Stats) bool {
	return len(c1.Delta(c2)) == 0
}
