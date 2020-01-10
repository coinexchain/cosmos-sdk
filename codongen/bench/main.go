package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/coinexchain/randsrc"
	codoncdc "github.com/cosmos/cosmos-sdk/codongen/codec"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s filename\n", os.Args[0])
		return
	}
	r := randsrc.NewRandSrcFromFile(os.Args[1])
	runBench(r)
}

func runBench(r codoncdc.RandSrc) {
	objects := make([]codoncdc.MsgMultiSend, 1000)
	objectsJ := make([][]byte, 1000)
	for i := 0; i < len(objects); i++ {
		objects[i] = codoncdc.RandMsgMultiSend(r)
		s, _ := json.Marshal(objects[i])
		objectsJ[i] = s
		//fmt.Printf("Here %s\n", s)
	}

	// Check correctness of codon
	var err error
	bzList := make([][]byte, 1000)
	for i := 0; i < len(objects); i++ {
		buf := make([]byte, 0, 1024)
		codoncdc.EncodeAny(&buf, objects[i])
		if err != nil {
			panic(err)
		}
		bzList[i] = buf
	}
	for i := 0; i < len(objects); i++ {
		obj, _, err := codoncdc.DecodeAny(bzList[i])
		v := obj.(codoncdc.MsgMultiSend)
		if err != nil {
			panic(err)
		}
		s, _ := json.Marshal(v)
		if !bytes.Equal(s, objectsJ[i]) {
			fmt.Printf("%s\n%s\n%d mismatch!\n", string(s), string(objectsJ[i]), i)
		}

	}

	fmt.Printf("Start running\n")
	totalBytes := 0
	nanoSecCount := int64(0)

	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	nanoSecCount = time.Now().UnixNano()
	for j := 0; j < 50; j++ {
		if j%10 == 9 {
			fmt.Printf("Now %d\n", j)
		}
		for i := 0; i < len(objects); i++ {
			bzList[i], err = cdc.MarshalBinaryBare(objects[i])
			totalBytes += len(bzList[i])
			if err != nil {
				panic(err)
			}
		}
		for i := 0; i < len(objects); i++ {
			err = cdc.UnmarshalBinaryBare(bzList[i], &objects[i])
			if err != nil {
				panic(err)
			}
		}
	}
	span := time.Now().UnixNano() - nanoSecCount
	fmt.Printf("Amino: time = %d, bytes = %d, bytes/ns = %f\n", span, totalBytes, float64(totalBytes)/float64(span))

	totalBytes = 0
	nanoSecCount = time.Now().UnixNano()
	for j := 0; j < 50; j++ {
		if j%10 == 9 {
			fmt.Printf("Now %d\n", j)
		}
		for i := 0; i < len(objects); i++ {
			bzList[i] = bzList[i][:0]
			codoncdc.EncodeAny(&bzList[i], objects[i])
			totalBytes += len(bzList[i])
		}
		for i := 0; i < len(objects); i++ {
			_, _, err = codoncdc.DecodeAny(bzList[i])
			if err != nil {
				panic(err)
			}
		}
	}
	span = time.Now().UnixNano() - nanoSecCount
	fmt.Printf("Codon: time = %d, bytes = %d, bytes/ns = %f\n", span, totalBytes, float64(totalBytes)/float64(span))
}
