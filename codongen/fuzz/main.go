package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	amino "github.com/tendermint/go-amino"

	"github.com/coinexchain/codon"
	codoncdc "github.com/cosmos/cosmos-sdk/codongen/codec"
	"github.com/coinexchain/randsrc"
)

var Count = 100 * 10000

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s filename\n", os.Args[0])
		return
	}
	r := randsrc.NewRandSrcFromFile(os.Args[1])
	runRandTest(r)
}

func runRandTest(r codoncdc.RandSrc) {
	leafTypes := codoncdc.GetLeafTypes()

	buf := make([]byte, 0, 4096)
	for i := 0; i < Count; i++ {
		if i%10000 == 0 {
			fmt.Printf("=== %d ===\n", i)
		}

		ifc := codoncdc.RandAny(r)
		origS, _ := json.Marshal(ifc)
		buf = buf[:0]

		codoncdc.EncodeAny(&buf, ifc)
		ifcDec, _, err := codoncdc.DecodeAny(buf)
		if err != nil {
			fmt.Printf("Now: %d\n", i)
			codon.ShowInfoForVar(leafTypes, ifc)
			panic(err)
		}
		decS, _ := json.Marshal(ifcDec)
		if !bytes.Equal(origS, decS) {
			fmt.Printf("Now: %d\n%s\n%s\n", i, string(origS), string(decS))
			codon.ShowInfoForVar(leafTypes, ifc)
			panic("Mismatch!")
		}
	}
}

func registerAll(cdcImp *codoncdc.CodecImp, cdcAmino *amino.Codec) {
	for _, entry := range codoncdc.TypeEntryList {
		v := entry.Value
		name := entry.Alias
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() == reflect.Interface {
			cdcImp.RegisterInterface(v, nil)
			cdcAmino.RegisterInterface(v, nil)
		} else {
			cdcImp.RegisterConcrete(v, name, nil)
			cdcAmino.RegisterConcrete(v, name, nil)
		}
	}
}

func findMismatch(a, b []byte) int {
	length := len(a)
	if len(b) < len(a) {
		length = len(b)
	}
	for i := 0; i < length; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	if len(b) != len(a) {
		return length
	}
	return -1
}

