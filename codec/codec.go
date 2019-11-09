package codec

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	amino "github.com/wide-key/wrap-amino"
)

func RegisterCrypto(cdc *amino.Codec) {
	// These are all written here instead of
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(ed25519.PubKeyEd25519{},
		ed25519.PubKeyAminoName, nil)
	cdc.RegisterConcrete(secp256k1.PubKeySecp256k1{},
		secp256k1.PubKeyAminoName, nil)
	cdc.RegisterConcrete(multisig.PubKeyMultisigThreshold{},
		multisig.PubKeyMultisigThresholdAminoRoute, nil)

	cdc.RegisterInterface((*crypto.PrivKey)(nil), nil)
	cdc.RegisterConcrete(ed25519.PrivKeyEd25519{},
		ed25519.PrivKeyAminoName, nil)
	cdc.RegisterConcrete(secp256k1.PrivKeySecp256k1{},
		secp256k1.PrivKeyAminoName, nil)
}

// amino codec to marshal/unmarshal
type Codec = amino.Codec

func New() *Codec {
	return amino.NewCodec()
}

// attempt to make some pretty json
func MarshalJSONIndent(cdc *Codec, obj interface{}) ([]byte, error) {
	bz, err := cdc.MarshalJSON(obj)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	err = json.Indent(&out, bz, "", "  ")
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// MustMarshalJSONIndent executes MarshalJSONIndent except it panics upon failure.
func MustMarshalJSONIndent(cdc *Codec, obj interface{}) []byte {
	bz, err := MarshalJSONIndent(cdc, obj)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %s", err))
	}

	return bz
}

//__________________________________________________________________
var initFuncList = make([]func(), 1, 100)

func AddInitFunc(fn func()) {
	initFuncList = append(initFuncList, fn)
}

func SetFirstInitFunc(fn func()) {
	initFuncList[0] = fn
}

func RunInitFuncList() {
	for _, fn := range initFuncList {
		if fn != nil {
			fn()
		}
	}
	initFuncList = nil
}
//__________________________________________________________________

// generic sealed codec to be used throughout sdk
var Cdc *Codec

func init() {
	AddInitFunc(func() {
		cdc := New()
		RegisterCrypto(cdc)
		Cdc = cdc.Seal()
	})
}
