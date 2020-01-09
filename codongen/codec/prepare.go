package codec

import (
	"io"

	"github.com/coinexchain/codon"
)

var TypeEntryList = []codon.TypeEntry{
	{Alias: "PubKey", Value: (*PubKey)(nil)},
	{Alias: "PrivKey", Value: (*PrivKey)(nil)},
	{Alias: "Msg", Value: (*Msg)(nil)},
	{Alias: "Account", Value: (*Account)(nil)},
	{Alias: "VestingAccount", Value: (*VestingAccount)(nil)},
	{Alias: "Content", Value: (*Content)(nil)},
	{Alias: "Tx", Value: (*Tx)(nil)},
	{Alias: "ModuleAccountI", Value: (*ModuleAccountI)(nil)},
	{Alias: "SupplyI", Value: (*SupplyI)(nil)},

	//{Alias: "DuplicateVoteEvidence", Value: DuplicateVoteEvidence{}},
	{Alias: "PrivKeyEd25519", Value: PrivKeyEd25519{}},
	{Alias: "PrivKeySecp256k1", Value: PrivKeySecp256k1{}},
	{Alias: "PubKeyEd25519", Value: PubKeyEd25519{}},
	{Alias: "PubKeySecp256k1", Value: PubKeySecp256k1{}},
	{Alias: "PubKeyMultisigThreshold", Value: PubKeyMultisigThreshold{}},
	{Alias: "SignedMsgType", Value: SignedMsgType(0)},
	{Alias: "VoteOption", Value: VoteOption(0)},
	{Alias: "Vote", Value: Vote{}},

	{Alias: "SdkInt", Value: SdkInt{}},
	{Alias: "SdkDec", Value: SdkDec{}},

	{Alias: "uint64", Value: uint64(0)},
	{Alias: "int64", Value: int64(0)},

	{Alias: "ConsAddress", Value: ConsAddress{}},
	{Alias: "Coin", Value: Coin{}},
	{Alias: "DecCoin", Value: DecCoin{}},
	{Alias: "StdSignature", Value: StdSignature{}},
	{Alias: "ParamChange", Value: ParamChange{}},
	{Alias: "Input", Value: Input{}},
	{Alias: "Output", Value: Output{}},
	{Alias: "AccAddress", Value: AccAddress{}},

	{Alias: "BaseAccount", Value: BaseAccount{}},
	{Alias: "BaseVestingAccount", Value: BaseVestingAccount{}},
	{Alias: "ContinuousVestingAccount", Value: ContinuousVestingAccount{}},
	{Alias: "DelayedVestingAccount", Value: DelayedVestingAccount{}},
	{Alias: "ModuleAccount", Value: ModuleAccount{}},
	{Alias: "StdTx", Value: StdTx{}},
	{Alias: "MsgBeginRedelegate", Value: MsgBeginRedelegate{}},
	{Alias: "MsgCreateValidator", Value: MsgCreateValidator{}},
	{Alias: "MsgDelegate", Value: MsgDelegate{}},
	{Alias: "MsgEditValidator", Value: MsgEditValidator{}},
	{Alias: "MsgSetWithdrawAddress", Value: MsgSetWithdrawAddress{}},
	{Alias: "MsgUndelegate", Value: MsgUndelegate{}},
	{Alias: "MsgUnjail", Value: MsgUnjail{}},
	{Alias: "MsgWithdrawDelegatorReward", Value: MsgWithdrawDelegatorReward{}},
	{Alias: "MsgWithdrawValidatorCommission", Value: MsgWithdrawValidatorCommission{}},
	{Alias: "MsgDeposit", Value: MsgDeposit{}},
	{Alias: "MsgVote", Value: MsgVote{}},
	{Alias: "ParameterChangeProposal", Value: ParameterChangeProposal{}},
	{Alias: "SoftwareUpgradeProposal", Value: SoftwareUpgradeProposal{}},
	{Alias: "TextProposal", Value: TextProposal{}},
	{Alias: "CommunityPoolSpendProposal", Value: CommunityPoolSpendProposal{}},
	{Alias: "MsgMultiSend", Value: MsgMultiSend{}},
	{Alias: "FeePool", Value: FeePool{}},
	{Alias: "MsgSend", Value: MsgSend{}},
	{Alias: "MsgVerifyInvariant", Value: MsgVerifyInvariant{}},
	{Alias: "Supply", Value: Supply{}},

	{Alias: "AccAddressList", Value: AccAddressList(nil)},
	{Alias: "CommitInfo", Value: CommitInfo{}},
	{Alias: "StoreInfo", Value: StoreInfo{}},

	{Alias: "Validator", Value: Validator{}},
	{Alias: "Delegation", Value: Delegation{}},
	{Alias: "BondStatus", Value: BondStatus(0)},
	{Alias: "DelegatorStartingInfo", Value: DelegatorStartingInfo{}},
	{Alias: "ValidatorHistoricalRewards", Value: ValidatorHistoricalRewards{}},
	{Alias: "ValidatorCurrentRewards", Value: ValidatorCurrentRewards{}},
	{Alias: "ValidatorSigningInfo", Value: ValidatorSigningInfo{}},
	{Alias: "ValidatorSlashEvent", Value: ValidatorSlashEvent{}},

	{Alias: "DecCoins", Value: DecCoins{}},
	{Alias: "ValAddress", Value: ValAddress{}},
	{Alias: "ValAddressList", Value: ValAddressList(nil)},
}

func GenerateCodecFile(w io.Writer) {
	extraImports := []string{`"time"`, `"math/big"`, `sdk "github.com/cosmos/cosmos-sdk/types"`,
		`"github.com/cosmos/cosmos-sdk/codec"`}
	extraImports = append(extraImports, codon.ImportsForBridgeLogic...)
	extraLogics := extraLogicsForLeafTypes + codon.BridgeLogic
	ignoreImpl := make(map[string]string)
	ignoreImpl["StdSignature"] = "PubKey"
	ignoreImpl["PubKeyMultisigThreshold"] = "PubKey"
	codon.GenerateCodecFile(w, GetLeafTypes(), ignoreImpl, TypeEntryList, extraLogics, extraImports)
}

func GetLeafTypes() map[string]string {
	leafTypes := make(map[string]string, 20)
	leafTypes["github.com/cosmos/cosmos-sdk/types.Int"] = "sdk.Int"
	leafTypes["github.com/cosmos/cosmos-sdk/types.Dec"] = "sdk.Dec"
	leafTypes["time.Time"] = "time.Time"
	return leafTypes
}

const MaxSliceLength = 10
const MaxStringLength = 100

var extraLogicsForLeafTypes = `
func init() {
	codec.SetFirstInitFunc(func() {
		amino.Stub = &CodonStub{}
	})
}
func EncodeTime(t time.Time) []byte {
	t = t.UTC()
	sec := t.Unix()
	var buf [20]byte
	n := binary.PutVarint(buf[:], sec)

	nanosec := t.Nanosecond()
	m := binary.PutVarint(buf[n:], int64(nanosec))
	return buf[:n+m]
}

func DecodeTime(bz []byte) (t time.Time, m int, err error) {
	var bs []byte
	n, err := codonGetByteSlice(&bs, bz)
	if err != nil {
		return
	}
	t, m, err = DecodeTimeHelp(bs)
	m += n
	return
}

func DecodeTimeHelp(bz []byte) (time.Time, int, error) {
	sec, n := binary.Varint(bz)
	var err error
	if n == 0 {
		// buf too small
		err = errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		err = errors.New("EOF decoding varint")
	}
	if err!=nil {
		return time.Unix(sec,0), n, err
	}

	nanosec, m := binary.Varint(bz[n:])
	if m == 0 {
		// buf too small
		err = errors.New("buffer too small")
	} else if m < 0 {
		// value larger than 64 bits (overflow)
		// and -m is the number of bytes read
		m = -m
		err = errors.New("EOF decoding varint")
	}
	if err!=nil {
		return time.Unix(sec,nanosec), n+m, err
	}

	return time.Unix(sec, nanosec).UTC(), n+m, nil
}

func StringToTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

var maxSec = StringToTime("9999-09-29T08:02:06.647266Z").Unix()

func RandTime(r RandSrc) time.Time {
	sec := r.GetInt64()
	nanosec := r.GetInt64()
	if sec < 0 {
		sec = -sec
	}
	if nanosec < 0 {
		nanosec = -nanosec
	}
	nanosec = nanosec%(1000*1000*1000)
	sec = sec%maxSec
	return time.Unix(sec, nanosec).UTC()
}


func DeepCopyTime(t time.Time) time.Time {
	return t.Add(time.Duration(0))
}

func ByteSliceWithLengthPrefix(bz []byte) []byte {
	buf := make([]byte, binary.MaxVarintLen64+len(bz))
	n := binary.PutUvarint(buf[:], uint64(len(bz)))
	return append(buf[0:n], bz...)
}

func EncodeInt(v sdk.Int) []byte {
	bz := v.BigInt().Bytes()
	res := ByteSliceWithLengthPrefix(bz)

	b := byte(0)
	if v.BigInt().Sign() < 0 {
		b = byte(1)
	}
	res = append(res, b)
	return res
}

func DecodeInt(bz []byte) (v sdk.Int, n int, err error) {
	var bs []byte
	n, err = codonGetByteSlice(&bs, bz)
	if err != nil {
		return
	}
	var k int
	isNeg := codonDecodeBool(bz[n:], &k, &err)
	n = n + 1
	x := big.NewInt(0)
	z := big.NewInt(0)
	x.SetBytes(bs)
	if isNeg {
		z.Neg(x)
		v = sdk.NewIntFromBigInt(z)
	} else {
		v = sdk.NewIntFromBigInt(x)
	}
	return
}

func RandInt(r RandSrc) sdk.Int {
	res := sdk.NewInt(r.GetInt64())
	count := int(r.GetInt64()%3)
	for i:=0; i<count; i++ {
		res = res.MulRaw(r.GetInt64())
	}
	return res
}

func DeepCopyInt(i sdk.Int) sdk.Int {
	return i.AddRaw(0)
}

func EncodeDec(v sdk.Dec) []byte {
	bz := v.Int.Bytes()
	res := ByteSliceWithLengthPrefix(bz)

	b := byte(0)
	if v.Int.Sign() < 0 {
		b = byte(1)
	}
	res = append(res, b)
	return res
}

func DecodeDec(bz []byte) (v sdk.Dec, n int, err error) {
	var bs []byte
	n, err = codonGetByteSlice(&bs, bz)
	if err != nil {
		return
	}
	var k int
	isNeg := codonDecodeBool(bz[n:], &k, &err)
	n = n + 1
	if err != nil {
		return
	}
	v = sdk.ZeroDec()
	v.Int.SetBytes(bs)
	if isNeg {
		v.Int.Neg(v.Int)
	}
	return
}

func RandDec(r RandSrc) sdk.Dec {
	res := sdk.NewDec(r.GetInt64())
	count := int(r.GetInt64()%3)
	for i:=0; i<count; i++ {
		res = res.MulInt64(r.GetInt64())
	}
	res = res.QuoInt64(r.GetInt64()&0xFFFFFFFF)
	return res
}

func DeepCopyDec(d sdk.Dec) sdk.Dec {
	return d.MulInt64(1)
}


`
