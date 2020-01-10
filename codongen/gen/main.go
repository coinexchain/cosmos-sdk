package main

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/codongen/codec"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s [codec|proto]\n", os.Args[0])
		return
	}
	switch os.Args[1] {
	case "codec":
		genCodec()
	case "proto":
		codec.GenerateProtoFile()
	default:
		fmt.Printf("usage: %s [codec|proto]\n", os.Args[0])
	}
}

func genCodec() {
	codec.GenerateCodecFile(os.Stdout)
}
