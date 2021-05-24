package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jaypipes/ghw/pkg/snapshot"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	envVarKniSnapshotKeep string = "KNI_SNAPSHOT_KEEP"
)

var (
	knitBaseDir  string
	binariesPath string
	snapshotKeep bool
)

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "E2E Suite")
}

var _ = ginkgo.BeforeSuite(func() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		ginkgo.Fail("Cannot retrieve tests directory")
	}
	basedir := filepath.Dir(file)
	knitBaseDir = filepath.Clean(filepath.Join(basedir, "..", ".."))
	binariesPath = filepath.Clean(filepath.Join(knitBaseDir, "_output"))
	fmt.Fprintf(ginkgo.GinkgoWriter, "using binaries at %q\n", binariesPath)

	if _, ok = os.LookupEnv(envVarKniSnapshotKeep); ok {
		snapshotKeep = true
	}
})

func snapshotBeforeEach(fixtureName, snapshotName string) string {
	path := filepath.Join(dataDirFor(fixtureName), snapshotName)

	unpackedPath, err := snapshot.Unpack(path)
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("Failed to unpack the snapshot %q: %v", path, err))
	}

	fmt.Fprintf(ginkgo.GinkgoWriter, "unpacked snapshot %q at %q\n", path, unpackedPath)
	return unpackedPath
}

func snapshotAfterEach(snapshotRoot string) {
	if snapshotKeep {
		return
	}
	if err := snapshot.Cleanup(snapshotRoot); err != nil {
		ginkgo.Fail(fmt.Sprintf("Failed to cleanup the snapshot at %q: %v", snapshotRoot, err))
	}
}

func getJSONBlobsDiff(want, got []byte) (string, error) {
	var wantObj interface{}
	var gotObj interface{}

	if err := json.Unmarshal(want, &wantObj); err != nil {
		return "", fmt.Errorf("Error unmarshalling data for 'want': %v", err)
	}
	if err := json.Unmarshal(got, &gotObj); err != nil {
		return "", fmt.Errorf("Error unmarshalling data for 'got': %v", err)
	}

	return cmp.Diff(wantObj, gotObj), nil
}

func dataDirFor(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		ginkgo.Fail("Cannot retrieve tests directory")
	}
	basedir := filepath.Dir(file)
	return filepath.Clean(filepath.Join(basedir, "..", "data", name))

}
