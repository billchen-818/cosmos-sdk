package ledger

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptoAmino "github.com/cosmos/cosmos-sdk/crypto/codec"
)

var cdc = codec.New()

func init() {
	RegisterAmino(cdc)
	cryptoAmino.RegisterCrypto(cdc)
}

// RegisterAmino registers all go-crypto related types in the given (amino) codec.
func RegisterAmino(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(PrivKeyLedgerSM2{},
		"tendermint/PrivKeyLedgerSM2", nil)
}
