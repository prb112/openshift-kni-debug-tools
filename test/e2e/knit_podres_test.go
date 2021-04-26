package e2eknit

import (
	"fmt"
	"os/exec"
	"path/filepath"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("knit podresources tests", func() {
	g.Context("With unknown API requested", func() {
		g.It("Fails with known error message", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "knit"),
				"podres",
				"foobar",
			}
			fmt.Fprintf(g.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = g.GinkgoWriter

			_, err := cmd.Output()
			o.Expect(err).To(o.HaveOccurred())
		})
	})
})
