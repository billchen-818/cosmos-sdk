package signing_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func MakeTestHandlerMap() signing.SignModeHandler {
	return signing.NewSignModeHandlerMap(
		signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		[]signing.SignModeHandler{
			authtypes.NewStdTxSignModeHandler(),
		},
	)
}

func TestHandlerMap_GetSignBytes(t *testing.T) {
	priv1 := keys.Secp256K1PrivKey{Key: secp256k1.GenPrivKey()}
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	priv2 := keys.Secp256K1PrivKey{Key: secp256k1.GenPrivKey()}
	addr2 := sdk.AccAddress(priv2.PubKey().Address())

	coins := sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}

	fee := authtypes.StdFee{
		Amount: coins,
		Gas:    10000,
	}
	memo := "foo"
	msgs := []sdk.Msg{
		&banktypes.MsgSend{
			FromAddress: addr1,
			ToAddress:   addr2,
			Amount:      coins,
		},
	}

	tx := authtypes.StdTx{
		Msgs:       msgs,
		Fee:        fee,
		Signatures: nil,
		Memo:       memo,
	}

	var (
		chainId        = "test-chain"
		accNum  uint64 = 7
		seqNum  uint64 = 7
	)

	handler := MakeTestHandlerMap()
	aminoJSONHandler := authtypes.NewStdTxSignModeHandler()

	signingData := signing.SignerData{
		ChainID:       chainId,
		AccountNumber: accNum,
		Sequence:      seqNum,
	}
	signBz, err := handler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, signingData, tx)
	require.NoError(t, err)

	expectedSignBz, err := aminoJSONHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, signingData, tx)
	require.NoError(t, err)

	require.Equal(t, expectedSignBz, signBz)

	// expect error with wrong sign mode
	_, err = aminoJSONHandler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_DIRECT, signingData, tx)
	require.Error(t, err)
}

func TestHandlerMap_DefaultMode(t *testing.T) {
	handler := MakeTestHandlerMap()
	require.Equal(t, signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, handler.DefaultMode())
}

func TestHandlerMap_Modes(t *testing.T) {
	handler := MakeTestHandlerMap()
	modes := handler.Modes()
	require.Contains(t, modes, signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	require.Len(t, modes, 1)
}
