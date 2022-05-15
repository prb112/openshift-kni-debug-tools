package e2e

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("knit ethtool tests", func() {
	g.Context("With loopback interface", func() {
		g.It("should report the interface features", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "knit"),
				"-J",
				"ethtool",
				"-k",
				"lo",
			}
			fmt.Fprintf(g.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter

			out, err := cmd.Output()
			o.Expect(err).ToNot(o.HaveOccurred())

			var features map[string]bool
			if err := json.Unmarshal(out, &features); err != nil {
				g.Fail("failed to unmarshal JSON data")
			}
			o.Expect(features).ToNot(o.BeEmpty())

			countEnabled := 0
			countDisabled := 0
			for _, enabled := range features {
				if enabled {
					countEnabled++
				} else {
					countDisabled++
				}
			}

			o.Expect(countEnabled).ToNot(o.BeZero(), "no enabled features for loopback")
			o.Expect(countDisabled).ToNot(o.BeZero(), "no disabled features for loopback")
		})
	})
})
