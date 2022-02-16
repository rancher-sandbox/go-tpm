package tpm_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/rancher-sandbox/go-tpm"
	. "github.com/rancher-sandbox/go-tpm/backend"
)

// In order to run this suite a swtpm socket is required. e.g.:
// swtpm socket --server type=unixio,path=/tmp/tpm-server --ctrl type=unixio,path=/tmp/tpm-ctrl --tpm2
// SWTPM_SOCKET=/tmp/tpm-ctrl ginkgo -r ./

var _ = Describe("TPM with SWTPM", func() {
	socket := os.Getenv("SWTPM_SOCKET")
	Context("opening socket connection", func() {
		// Note, this doesn't work
		PIt("dials in just fine", func() {
			if socket == "" {
				Skip("No socket file specified")
			}

			b, err := Socket(socket)
			Expect(err).ToNot(HaveOccurred())

			str, err := GetPubHash(WithCommandChannel(b))
			Expect(err).ToNot(HaveOccurred())

			Expect(str).ToNot(BeEmpty())
		})
	})
})

var _ = Describe("Simulated TPM", func() {
	Context("opening socket connection", func() {
		It("dials in just fine", func() {
			str, err := GetPubHash(Emulated)
			Expect(err).ToNot(HaveOccurred())
			Expect(str).ToNot(BeEmpty())
		})
	})

	Context("specifying a seed", func() {
		It("same pubkey with same seed", func() {
			str, err := GetPubHash(Emulated, WithSeed(1))
			Expect(err).ToNot(HaveOccurred())
			str2, err := GetPubHash(Emulated, WithSeed(1))
			Expect(err).ToNot(HaveOccurred())
			Expect(str).To(Equal(str2))
		})
	})
})
