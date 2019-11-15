package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codongen/codec"
)

func main() {
	//codec.ShowInfo()
	genCode()
}

func genCode() {
	codec.GenerateCodecFile(os.Stdout)
}
