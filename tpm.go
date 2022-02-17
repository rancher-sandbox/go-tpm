package tpm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-attestation/attest"
	"github.com/google/go-tpm-tools/simulator"
	"github.com/rancher-sandbox/go-tpm/backend"
)

// ResolveToken is just syntax sugar around GetPubHash.
// If the token provided is in EK's form it just returns it, otherwise
// retrieves the pubhash
func ResolveToken(token string, opts ...Option) (bool, string, error) {
	if !strings.HasPrefix(token, "tpm://") {
		return false, token, nil
	}

	hash, err := GetPubHash(opts...)
	return true, hash, err
}

// GetPubHash returns the EK's pub hash
func GetPubHash(opts ...Option) (string, error) {
	c := &config{}
	c.apply(opts...)

	ek, err := getEK(c)
	if err != nil {
		return "", fmt.Errorf("getting EK: %w", err)
	}

	hash, err := getPubHash(ek)
	if err != nil {
		return "", fmt.Errorf("hashing EK: %w", err)
	}

	return hash, nil
}

func getTPM(c *config) (*attest.TPM, error) {

	cfg := &attest.OpenConfig{
		TPMVersion: attest.TPMVersion20,
	}
	if c.commandChannel != nil {
		cfg.CommandChannel = c.commandChannel
	}

	if c.emulated && c.seed == 0 {
		sim, err := simulator.Get()
		if err != nil {
			return nil, err
		}
		cfg.CommandChannel = backend.Fake(sim)
	} else {
		sim, err := simulator.GetWithFixedSeedInsecure(c.seed)
		if err != nil {
			return nil, err
		}
		cfg.CommandChannel = backend.Fake(sim)
	}

	return attest.OpenTPM(cfg)

}

func getEK(c *config) (*attest.EK, error) {
	var err error

	tpm, err := getTPM(c)
	if err != nil {
		return nil, fmt.Errorf("opening tpm: %w", err)
	}
	defer tpm.Close()

	eks, err := tpm.EKs()
	if err != nil {
		return nil, fmt.Errorf("getting eks: %w", err)
	}

	if len(eks) == 0 {
		return nil, fmt.Errorf("failed to find EK")
	}

	return &eks[0], nil
}

func getToken(data *attestationData) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshalling attestation data: %w", err)
	}

	return "Bearer TPM" + base64.StdEncoding.EncodeToString(bytes), nil
}

func getAttestationData(c *config) (*attestationData, []byte, error) {
	var err error

	tpm, err := getTPM(c)
	if err != nil {
		return nil, nil, fmt.Errorf("opening tpm: %w", err)
	}
	defer tpm.Close()

	eks, err := tpm.EKs()
	if err != nil {
		return nil, nil, err
	}

	ak, err := tpm.NewAK(nil)
	if err != nil {
		return nil, nil, err
	}
	defer ak.Close(tpm)
	params := ak.AttestationParameters()

	if len(eks) == 0 {
		return nil, nil, fmt.Errorf("failed to find EK")
	}

	ek := &eks[0]
	ekBytes, err := encodeEK(ek)
	if err != nil {
		return nil, nil, err
	}

	aikBytes, err := ak.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("marshaling AK: %w", err)
	}

	return &attestationData{
		EK: ekBytes,
		AK: &params,
	}, aikBytes, nil
}
