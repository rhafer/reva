package metadata_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestXattrs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Xattrs Suite")
}
