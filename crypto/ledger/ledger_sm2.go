package ledger

import (
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/pkg/errors"
	gm "github.com/tjfoc/gmsm/sm2"

	tmbtcec "github.com/tendermint/btcd/btcec"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/sm2"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
)

var (
	discoverLedger discoverLedgerFn
)

type (
	discoverLedgerFn func() (SM2, error)

	SM2 interface {
		Close() error

		GetPublicKeySM2([]uint32) ([]byte, error)

		GetAddressPubKeySM2([]uint32, string) ([]byte, string, error)

		SignSM2([]uint32, []byte) ([]byte, error)
	}

	PrivKeyLedgerSM2 struct {
		CachedPubKey tmcrypto.PubKey
		Path         hd.BIP44Params
	}
)

func NewPrivKeySM2Unsafe(path hd.BIP44Params) (tmcrypto.PrivKey, error) {
	device, err := getDevice()
	if err != nil {
		return nil, err
	}
	defer warnIfErrors(device.Close)

	pubKey, err := getPubKeyUnsafe(device, path)
	if err != nil {
		return nil, err
	}

	return PrivKeyLedgerSM2{pubKey, path}, nil
}

func NewPrivKeySM2(path hd.BIP44Params, hrp string) (tmcrypto.PrivKey, string, error) {
	device, err := getDevice()
	if err != nil {
		return nil, "", err
	}
	defer warnIfErrors(device.Close)

	pubKey, addr, err := getPubKeyAddrSafe(device, path, hrp)
	if err != nil {
		return nil, "", err
	}

	return PrivKeyLedgerSM2{pubKey, path}, addr, nil
}

func (pkl PrivKeyLedgerSM2) PubKey() tmcrypto.PubKey {
	return pkl.CachedPubKey
}

func (pkl PrivKeyLedgerSM2) Sign(message []byte) ([]byte, error) {
	device, err := getDevice()
	if err != nil {
		return nil, err
	}
	defer warnIfErrors(device.Close)

	return sign(device, pkl, message)
}

func ShowAddress(path hd.BIP44Params, expectedPubKey tmcrypto.PubKey,
	accountAddressPrefix string) error {
	device, err := getDevice()
	if err != nil {
		return err
	}
	defer warnIfErrors(device.Close)

	pubKey, err := getPubKeyUnsafe(device, path)
	if err != nil {
		return err
	}

	if !pubKey.Equals(expectedPubKey) {
		return fmt.Errorf("the key's pubkey does not match with the one retrieved from Ledger. Check that the HD path and device are the correct ones")
	}

	pubKey2, _, err := getPubKeyAddrSafe(device, path, accountAddressPrefix)
	if err != nil {
		return err
	}

	if !pubKey2.Equals(expectedPubKey) {
		return fmt.Errorf("the key's pubkey does not match with the one retrieved from Ledger. Check that the HD path and device are the correct ones")
	}

	return nil
}

func (pkl PrivKeyLedgerSM2) ValidateKey() error {
	device, err := getDevice()
	if err != nil {
		return err
	}
	defer warnIfErrors(device.Close)

	return validateKey(device, pkl)
}

func (pkl *PrivKeyLedgerSM2) AssertIsPrivKeyInner() {}

func (pkl PrivKeyLedgerSM2) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(pkl)
}

func (pkl PrivKeyLedgerSM2) Equals(other tmcrypto.PrivKey) bool {
	if otherKey, ok := other.(PrivKeyLedgerSM2); ok {
		return pkl.CachedPubKey.Equals(otherKey.CachedPubKey)
	}
	return false
}

func (pkl PrivKeyLedgerSM2) Type() string { return "PrivKeyLedgerSM2" }

func warnIfErrors(f func() error) {
	if err := f(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, "received error when closing ledger connection", err)
	}
}

func convertDERtoBER(signatureDER []byte) ([]byte, error) {
	sigDER, err := btcec.ParseDERSignature(signatureDER, btcec.S256())
	if err != nil {
		return nil, err
	}
	sigBER := tmbtcec.Signature{R: sigDER.R, S: sigDER.S}
	return sigBER.Serialize(), nil
}

func getDevice() (SM2, error) {
	if discoverLedger == nil {
		return nil, errors.New("no Ledger discovery function defined")
	}

	device, err := discoverLedger()
	if err != nil {
		return nil, errors.Wrap(err, "ledger nano S")
	}

	return device, nil
}

func validateKey(device SM2, pkl PrivKeyLedgerSM2) error {
	pub, err := getPubKeyUnsafe(device, pkl.Path)
	if err != nil {
		return err
	}

	// verify this matches cached address
	if !pub.Equals(pkl.CachedPubKey) {
		return fmt.Errorf("cached key does not match retrieved key")
	}

	return nil
}

func sign(device SM2, pkl PrivKeyLedgerSM2, msg []byte) ([]byte, error) {
	err := validateKey(device, pkl)
	if err != nil {
		return nil, err
	}

	sig, err := device.SignSM2(pkl.Path.DerivationPath(), msg)
	if err != nil {
		return nil, err
	}

	return convertDERtoBER(sig)
}

func getPubKeyUnsafe(device SM2, path hd.BIP44Params) (tmcrypto.PubKey, error) {
	publicKey, err := device.GetPublicKeySM2(path.DerivationPath())
	if err != nil {
		return nil, fmt.Errorf("please open Cosmos app on the Ledger device - error: %v", err)
	}

	cmp, err := sm2.ParsePubKey(publicKey, gm.P256Sm2())
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %v", err)
	}

	compressedPublicKey := make(sm2.PubKey, sm2.PubKeySize)
	copy(compressedPublicKey, gm.Compress(cmp))

	return compressedPublicKey, nil
}

func getPubKeyAddrSafe(device SM2, path hd.BIP44Params, hrp string) (tmcrypto.PubKey, string, error) {
	publicKey, addr, err := device.GetAddressPubKeySM2(path.DerivationPath(), hrp)
	if err != nil {
		return nil, "", fmt.Errorf("address %s rejected", addr)
	}

	cmp, err := sm2.ParsePubKey(publicKey, gm.P256Sm2())
	if err != nil {
		return nil, "", fmt.Errorf("error parsing public key: %v", err)
	}

	compressedPublicKey := make(sm2.PubKey, sm2.PubKeySize)
	copy(compressedPublicKey, gm.Compress(cmp))

	return compressedPublicKey, addr, nil
}
