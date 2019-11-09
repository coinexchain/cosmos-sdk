package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func init() {
	codec.AddInitFunc(func() {
		ModuleCdc = codec.New()
		codec.RegisterCrypto(ModuleCdc)
		ModuleCdc.Seal()
	})
}
