package iavl

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var cdc *codec.Codec

func init() {
	codec.AddInitFunc(func() {
		cdc = codec.New()
	})
}
