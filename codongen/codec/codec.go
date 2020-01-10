//nolint
package codec

import (
	"encoding/binary"
	"errors"
	"fmt"
	amino "github.com/coinexchain/codon/wrap-amino"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"io"
	"math/big"
	"reflect"
	"time"
)

type RandSrc interface {
	GetBool() bool
	GetInt() int
	GetInt8() int8
	GetInt16() int16
	GetInt32() int32
	GetInt64() int64
	GetUint() uint
	GetUint8() uint8
	GetUint16() uint16
	GetUint32() uint32
	GetUint64() uint64
	GetFloat32() float32
	GetFloat64() float64
	GetString(n int) string
	GetBytes(n int) []byte
}

func codonWriteVarint(w *[]byte, v int64) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], v)
	*w = append(*w, buf[0:n]...)
}
func codonWriteUvarint(w *[]byte, v uint64) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], v)
	*w = append(*w, buf[0:n]...)
}

func codonEncodeBool(n int, w *[]byte, v bool) {
	codonWriteUvarint(w, uint64(n)<<3)
	if v {
		codonWriteUvarint(w, uint64(1))
	} else {
		codonWriteUvarint(w, uint64(0))
	}
}
func codonEncodeVarint(n int, w *[]byte, v int64) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteVarint(w, int64(v))
}
func codonEncodeInt8(n int, w *[]byte, v int8) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteVarint(w, int64(v))
}
func codonEncodeInt16(n int, w *[]byte, v int16) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteVarint(w, int64(v))
}
func codonEncodeUvarint(n int, w *[]byte, v uint64) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteUvarint(w, v)
}
func codonEncodeUint8(n int, w *[]byte, v uint8) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteUvarint(w, uint64(v))
}
func codonEncodeUint16(n int, w *[]byte, v uint16) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteUvarint(w, uint64(v))
}

func codonEncodeByteSlice(n int, w *[]byte, v []byte) {
	codonWriteUvarint(w, (uint64(n)<<3)|2)
	codonWriteUvarint(w, uint64(len(v)))
	*w = append(*w, v...)
}
func codonEncodeString(n int, w *[]byte, v string) {
	codonEncodeByteSlice(n, w, []byte(v))
}
func codonDecodeBool(bz []byte, n *int, err *error) bool {
	return codonDecodeInt64(bz, n, err) != 0
}
func codonDecodeInt(bz []byte, n *int, err *error) int {
	return int(codonDecodeInt64(bz, n, err))
}
func codonDecodeInt8(bz []byte, n *int, err *error) int8 {
	return int8(codonDecodeInt64(bz, n, err))
}
func codonDecodeInt16(bz []byte, n *int, err *error) int16 {
	return int16(codonDecodeInt64(bz, n, err))
}
func codonDecodeInt32(bz []byte, n *int, err *error) int32 {
	return int32(codonDecodeInt64(bz, n, err))
}
func codonDecodeInt64(bz []byte, m *int, err *error) int64 {
	i, n := binary.Varint(bz)
	if n == 0 {
		// buf too small
		*err = errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		*err = errors.New("EOF decoding varint")
	}
	*m = n
	*err = nil
	return int64(i)
}
func codonDecodeUint(bz []byte, n *int, err *error) uint {
	return uint(codonDecodeUint64(bz, n, err))
}
func codonDecodeUint8(bz []byte, n *int, err *error) uint8 {
	return uint8(codonDecodeUint64(bz, n, err))
}
func codonDecodeUint16(bz []byte, n *int, err *error) uint16 {
	return uint16(codonDecodeUint64(bz, n, err))
}
func codonDecodeUint32(bz []byte, n *int, err *error) uint32 {
	return uint32(codonDecodeUint64(bz, n, err))
}
func codonDecodeUint64(bz []byte, m *int, err *error) uint64 {
	i, n := binary.Uvarint(bz)
	if n == 0 {
		// buf too small
		*err = errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		*err = errors.New("EOF decoding varint")
	}
	*m = n
	*err = nil
	return uint64(i)
}
func codonGetByteSlice(res *[]byte, bz []byte) (int, error) {
	length, n := binary.Uvarint(bz)
	if n == 0 {
		// buf too small
		return n, errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		return n, errors.New("EOF decoding varint")
	}
	if length == 0 {
		*res = nil
		return 0, nil
	}
	bz = bz[n:]
	if len(bz) < int(length) {
		*res = nil
		return 0, errors.New("Not enough bytes to read")
	}
	if *res == nil {
		*res = append(*res, bz[:length]...)
	} else {
		*res = append((*res)[:0], bz[:length]...)
	}
	return n + int(length), nil
}
func codonDecodeString(bz []byte, n *int, err *error) string {
	var res []byte
	*n, *err = codonGetByteSlice(&res, bz)
	return string(res)
}

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

func DecodeTime(bz []byte) (time.Time, int, error) {
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
	if err != nil {
		return time.Unix(sec, 0), n, err
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
	if err != nil {
		return time.Unix(sec, nanosec), n + m, err
	}

	return time.Unix(sec, nanosec).UTC(), n + m, nil
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
	nanosec = nanosec % (1000 * 1000 * 1000)
	sec = sec % maxSec
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
	b := byte(0)
	if v.BigInt().Sign() < 0 {
		b = byte(1)
	}
	bz := v.BigInt().Bytes()
	return append(bz, b)
}

func DecodeInt(bz []byte) (v sdk.Int, n int, err error) {
	isNeg := bz[len(bz)-1] != 0
	n = len(bz)
	x := big.NewInt(0)
	z := big.NewInt(0)
	x.SetBytes(bz[:len(bz)-1])
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
	count := int(r.GetInt64() % 3)
	for i := 0; i < count; i++ {
		res = res.MulRaw(r.GetInt64())
	}
	return res
}

func DeepCopyInt(i sdk.Int) sdk.Int {
	return i.AddRaw(0)
}

func EncodeDec(v sdk.Dec) []byte {
	b := byte(0)
	if v.Int.Sign() < 0 {
		b = byte(1)
	}
	bz := v.Int.Bytes()
	return append(bz, b)
}

func DecodeDec(bz []byte) (v sdk.Dec, n int, err error) {
	isNeg := bz[len(bz)-1] != 0
	n = len(bz)
	v = sdk.ZeroDec()
	v.Int.SetBytes(bz[:len(bz)-1])
	if isNeg {
		v.Int.Neg(v.Int)
	}
	return
}

func RandDec(r RandSrc) sdk.Dec {
	res := sdk.NewDec(r.GetInt64())
	count := int(r.GetInt64() % 3)
	for i := 0; i < count; i++ {
		res = res.MulInt64(r.GetInt64())
	}
	res = res.QuoInt64(r.GetInt64() & 0xFFFFFFFF)
	return res
}

func DeepCopyDec(d sdk.Dec) sdk.Dec {
	return d.MulInt64(1)
}

// ========= BridgeBegin ============
type CodecImp struct {
	sealed bool
}

var _ amino.Sealer = &CodecImp{}
var _ amino.CodecIfc = &CodecImp{}

func (cdc *CodecImp) MarshalBinaryBare(o interface{}) ([]byte, error) {
	s := CodonStub{}
	return s.MarshalBinaryBare(o)
}
func (cdc *CodecImp) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	s := CodonStub{}
	return s.MarshalBinaryLengthPrefixed(o)
}
func (cdc *CodecImp) MarshalBinaryLengthPrefixedWriter(w io.Writer, o interface{}) (n int64, err error) {
	bz, err := cdc.MarshalBinaryLengthPrefixed(o)
	m, err := w.Write(bz)
	return int64(m), err
}
func (cdc *CodecImp) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	s := CodonStub{}
	return s.UnmarshalBinaryBare(bz, ptr)
}
func (cdc *CodecImp) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	s := CodonStub{}
	return s.UnmarshalBinaryLengthPrefixed(bz, ptr)
}
func (cdc *CodecImp) UnmarshalBinaryLengthPrefixedReader(r io.Reader, ptr interface{}, maxSize int64) (n int64, err error) {
	if maxSize < 0 {
		panic("maxSize cannot be negative.")
	}

	// Read byte-length prefix.
	var l int64
	var buf [binary.MaxVarintLen64]byte
	for i := 0; i < len(buf); i++ {
		_, err = r.Read(buf[i : i+1])
		if err != nil {
			return
		}
		n += 1
		if buf[i]&0x80 == 0 {
			break
		}
		if n >= maxSize {
			err = fmt.Errorf("Read overflow, maxSize is %v but uvarint(length-prefix) is itself greater than maxSize.", maxSize)
		}
	}
	u64, _ := binary.Uvarint(buf[:])
	if err != nil {
		return
	}
	if maxSize > 0 {
		if uint64(maxSize) < u64 {
			err = fmt.Errorf("Read overflow, maxSize is %v but this amino binary object is %v bytes.", maxSize, u64)
			return
		}
		if (maxSize - n) < int64(u64) {
			err = fmt.Errorf("Read overflow, maxSize is %v but this length-prefixed amino binary object is %v+%v bytes.", maxSize, n, u64)
			return
		}
	}
	l = int64(u64)
	if l < 0 {
		err = fmt.Errorf("Read overflow, this implementation can't read this because, why would anyone have this much data?")
	}

	// Read that many bytes.
	var bz = make([]byte, l, l)
	_, err = io.ReadFull(r, bz)
	if err != nil {
		return
	}
	n += l

	// Decode.
	err = cdc.UnmarshalBinaryBare(bz, ptr)
	return
}

//------

func (cdc *CodecImp) MustMarshalBinaryBare(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryBare(o)
	if err != nil {
		panic(err)
	}
	return bz
}
func (cdc *CodecImp) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryLengthPrefixed(o)
	if err != nil {
		panic(err)
	}
	return bz
}
func (cdc *CodecImp) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		panic(err)
	}
}
func (cdc *CodecImp) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// ====================
func derefPtr(v interface{}) reflect.Type {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (cdc *CodecImp) PrintTypes(out io.Writer) error {
	for _, entry := range GetSupportList() {
		_, err := out.Write([]byte(entry))
		if err != nil {
			return err
		}
		_, err = out.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
func (cdc *CodecImp) RegisterConcrete(o interface{}, name string, copts *amino.ConcreteOptions) {
	if cdc.sealed {
		panic("Codec is already sealed")
	}
	t := derefPtr(o)
	path := t.PkgPath() + "." + t.Name()
	found := false
	for _, entry := range GetSupportList() {
		if path == entry {
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Sprintf("%s is not supported", path))
	}
}
func (cdc *CodecImp) RegisterInterface(o interface{}, _ *amino.InterfaceOptions) {
	if cdc.sealed {
		panic("Codec is already sealed")
	}
	t := derefPtr(o)
	path := t.PkgPath() + "." + t.Name()
	found := false
	for _, entry := range GetSupportList() {
		if path == entry {
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Sprintf("%s is not supported", path))
	}
}
func (cdc *CodecImp) SealImp() {
	if cdc.sealed {
		panic("Codec is already sealed")
	}
	cdc.sealed = true
}

// ========================================

type CodonStub struct {
}

func (_ *CodonStub) NewCodecImp() amino.CodecIfc {
	return &CodecImp{}
}
func (_ *CodonStub) DeepCopy(o interface{}) (r interface{}) {
	r = DeepCopyAny(o)
	return
}

func (_ *CodonStub) MarshalBinaryBare(o interface{}) ([]byte, error) {
	if _, ok := getMagicNumOfVar(o); !ok {
		return nil, errors.New("Not Supported Type")
	}
	buf := make([]byte, 0, 1024)
	EncodeAny(&buf, o)
	return buf, nil
}
func (s *CodonStub) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	if _, ok := getMagicNumOfVar(o); !ok {
		return nil, errors.New("Not Supported Type")
	}
	bz, err := s.MarshalBinaryBare(o)
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], uint64(len(bz)))
	return append(buf[:n], bz...), err
}
func (_ *CodonStub) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		panic("Unmarshal expects a pointer")
	}

	if len(bz) <= 4 {
		return fmt.Errorf("Byte slice is too short: %d", len(bz))
	}
	o, _, err := DecodeAny(bz)
	if rv.Elem().Kind() == reflect.Interface {
		AssignIfcPtrFromStruct(ptr, o)
	} else {
		rv.Elem().Set(reflect.ValueOf(o))
	}
	return err
}
func (s *CodonStub) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	if len(bz) == 0 {
		return errors.New("UnmarshalBinaryLengthPrefixed cannot decode empty bytes")
	}
	// Read byte-length prefix.
	u64, n := binary.Uvarint(bz)
	if n < 0 {
		return fmt.Errorf("Error reading msg byte-length prefix: got code %v", n)
	}
	if u64 > uint64(len(bz)-n) {
		return fmt.Errorf("Not enough bytes to read in UnmarshalBinaryLengthPrefixed, want %v more bytes but only have %v",
			u64, len(bz)-n)
	} else if u64 < uint64(len(bz)-n) {
		return fmt.Errorf("Bytes left over in UnmarshalBinaryLengthPrefixed, should read %v more bytes but have %v",
			u64, len(bz)-n)
	}
	bz = bz[n:]
	return s.UnmarshalBinaryBare(bz, ptr)
}
func (s *CodonStub) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	bz, err := s.MarshalBinaryLengthPrefixed(o)
	if err != nil {
		panic(err)
	}
	return bz
}

// ========================================
func (_ *CodonStub) UvarintSize(u uint64) int {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], u)
	return n
}
func (_ *CodonStub) EncodeByteSlice(w io.Writer, bz []byte) error {
	_, err := w.Write(ByteSliceWithLengthPrefix(bz))
	return err
}
func (s *CodonStub) ByteSliceSize(bz []byte) int {
	return s.UvarintSize(uint64(len(bz))) + len(bz)
}
func (_ *CodonStub) EncodeVarint(w io.Writer, i int64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], i)
	_, err := w.Write(buf[:n])
	return err
}
func (s *CodonStub) EncodeInt8(w io.Writer, i int8) error {
	return s.EncodeVarint(w, int64(i))
}
func (s *CodonStub) EncodeInt16(w io.Writer, i int16) error {
	return s.EncodeVarint(w, int64(i))
}
func (s *CodonStub) EncodeInt32(w io.Writer, i int32) error {
	return s.EncodeVarint(w, int64(i))
}
func (s *CodonStub) EncodeInt64(w io.Writer, i int64) error {
	return s.EncodeVarint(w, int64(i))
}
func (_ *CodonStub) EncodeUvarint(w io.Writer, u uint64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], u)
	_, err := w.Write(buf[:n])
	return err
}
func (s *CodonStub) EncodeByte(w io.Writer, b byte) error {
	return s.EncodeUvarint(w, uint64(b))
}
func (s *CodonStub) EncodeUint8(w io.Writer, u uint8) error {
	return s.EncodeUvarint(w, uint64(u))
}
func (s *CodonStub) EncodeUint16(w io.Writer, u uint16) error {
	return s.EncodeUvarint(w, uint64(u))
}
func (s *CodonStub) EncodeUint32(w io.Writer, u uint32) error {
	return s.EncodeUvarint(w, uint64(u))
}
func (s *CodonStub) EncodeUint64(w io.Writer, u uint64) error {
	return s.EncodeUvarint(w, uint64(u))
}
func (_ *CodonStub) EncodeBool(w io.Writer, b bool) error {
	u := byte(0)
	if b {
		u = byte(1)
	}
	_, err := w.Write([]byte{u})
	return err
}
func (s *CodonStub) EncodeString(w io.Writer, str string) error {
	return s.EncodeByteSlice(w, []byte(str))
}
func (_ *CodonStub) DecodeInt8(bz []byte) (i int8, n int, err error) {
	i = codonDecodeInt8(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeInt16(bz []byte) (i int16, n int, err error) {
	i = codonDecodeInt16(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeInt32(bz []byte) (i int32, n int, err error) {
	i = codonDecodeInt32(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeInt64(bz []byte) (i int64, n int, err error) {
	i = codonDecodeInt64(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeVarint(bz []byte) (i int64, n int, err error) {
	i = codonDecodeInt64(bz, &n, &err)
	return
}
func (s *CodonStub) DecodeByte(bz []byte) (b byte, n int, err error) {
	b = codonDecodeUint8(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeUint8(bz []byte) (u uint8, n int, err error) {
	u = codonDecodeUint8(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeUint16(bz []byte) (u uint16, n int, err error) {
	u = codonDecodeUint16(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeUint32(bz []byte) (u uint32, n int, err error) {
	u = codonDecodeUint32(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeUint64(bz []byte) (u uint64, n int, err error) {
	u = codonDecodeUint64(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeUvarint(bz []byte) (u uint64, n int, err error) {
	u = codonDecodeUint64(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeBool(bz []byte) (b bool, n int, err error) {
	b = codonDecodeBool(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error) {
	m, err := codonGetByteSlice(&bz2, bz)
	n += m
	return
}
func (_ *CodonStub) DecodeString(bz []byte) (s string, n int, err error) {
	s = codonDecodeString(bz, &n, &err)
	return
}
func (_ *CodonStub) VarintSize(i int64) int {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], i)
	return n
}

// ========= BridgeEnd ============
// Non-Interface
func EncodePrivKeyEd25519(w *[]byte, v PrivKeyEd25519) {
	codonEncodeByteSlice(0, w, v[:])
} //End of EncodePrivKeyEd25519

func DecodePrivKeyEd25519(bz []byte) (v PrivKeyEd25519, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			o := v[:]
			n, err = codonGetByteSlice(&o, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodePrivKeyEd25519

func RandPrivKeyEd25519(r RandSrc) PrivKeyEd25519 {
	var length int
	var v PrivKeyEd25519
	length = 64
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //array of uint8
		v[_0] = r.GetUint8()
	}
	return v
} //End of RandPrivKeyEd25519

func DeepCopyPrivKeyEd25519(in PrivKeyEd25519) (out PrivKeyEd25519) {
	var length int
	length = len(in)
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //array of uint8
		out[_0] = in[_0]
	}
	return
} //End of DeepCopyPrivKeyEd25519

// Non-Interface
func EncodePrivKeySecp256k1(w *[]byte, v PrivKeySecp256k1) {
	codonEncodeByteSlice(0, w, v[:])
} //End of EncodePrivKeySecp256k1

func DecodePrivKeySecp256k1(bz []byte) (v PrivKeySecp256k1, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			o := v[:]
			n, err = codonGetByteSlice(&o, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodePrivKeySecp256k1

func RandPrivKeySecp256k1(r RandSrc) PrivKeySecp256k1 {
	var length int
	var v PrivKeySecp256k1
	length = 32
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //array of uint8
		v[_0] = r.GetUint8()
	}
	return v
} //End of RandPrivKeySecp256k1

func DeepCopyPrivKeySecp256k1(in PrivKeySecp256k1) (out PrivKeySecp256k1) {
	var length int
	length = len(in)
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //array of uint8
		out[_0] = in[_0]
	}
	return
} //End of DeepCopyPrivKeySecp256k1

// Non-Interface
func EncodePubKeyEd25519(w *[]byte, v PubKeyEd25519) {
	codonEncodeByteSlice(0, w, v[:])
} //End of EncodePubKeyEd25519

func DecodePubKeyEd25519(bz []byte) (v PubKeyEd25519, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			o := v[:]
			n, err = codonGetByteSlice(&o, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodePubKeyEd25519

func RandPubKeyEd25519(r RandSrc) PubKeyEd25519 {
	var length int
	var v PubKeyEd25519
	length = 32
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //array of uint8
		v[_0] = r.GetUint8()
	}
	return v
} //End of RandPubKeyEd25519

func DeepCopyPubKeyEd25519(in PubKeyEd25519) (out PubKeyEd25519) {
	var length int
	length = len(in)
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //array of uint8
		out[_0] = in[_0]
	}
	return
} //End of DeepCopyPubKeyEd25519

// Non-Interface
func EncodePubKeySecp256k1(w *[]byte, v PubKeySecp256k1) {
	codonEncodeByteSlice(0, w, v[:])
} //End of EncodePubKeySecp256k1

func DecodePubKeySecp256k1(bz []byte) (v PubKeySecp256k1, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			o := v[:]
			n, err = codonGetByteSlice(&o, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodePubKeySecp256k1

func RandPubKeySecp256k1(r RandSrc) PubKeySecp256k1 {
	var length int
	var v PubKeySecp256k1
	length = 33
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //array of uint8
		v[_0] = r.GetUint8()
	}
	return v
} //End of RandPubKeySecp256k1

func DeepCopyPubKeySecp256k1(in PubKeySecp256k1) (out PubKeySecp256k1) {
	var length int
	length = len(in)
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //array of uint8
		out[_0] = in[_0]
	}
	return
} //End of DeepCopyPubKeySecp256k1

// Non-Interface
func EncodePubKeyMultisigThreshold(w *[]byte, v PubKeyMultisigThreshold) {
	codonEncodeUvarint(0, w, uint64(v.K))
	for _0 := 0; _0 < len(v.PubKeys); _0++ {
		codonEncodeByteSlice(1, w, func() []byte {
			w := make([]byte, 0, 64)
			EncodePubKey(&w, v.PubKeys[_0]) // interface_encode
			return w
		}()) // end of v.PubKeys[_0]
	}
} //End of EncodePubKeyMultisigThreshold

func DecodePubKeyMultisigThreshold(bz []byte) (v PubKeyMultisigThreshold, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.K
			v.K = uint(codonDecodeUint(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.PubKeys
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp PubKey
			tmp, n, err = DecodePubKey(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.PubKeys = append(v.PubKeys, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodePubKeyMultisigThreshold

func RandPubKeyMultisigThreshold(r RandSrc) PubKeyMultisigThreshold {
	var length int
	var v PubKeyMultisigThreshold
	v.K = r.GetUint()
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.PubKeys = nil
	} else {
		v.PubKeys = make([]PubKey, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of interface
		v.PubKeys[_0] = RandPubKey(r)
	}
	return v
} //End of RandPubKeyMultisigThreshold

func DeepCopyPubKeyMultisigThreshold(in PubKeyMultisigThreshold) (out PubKeyMultisigThreshold) {
	var length int
	out.K = in.K
	length = len(in.PubKeys)
	if length == 0 {
		out.PubKeys = nil
	} else {
		out.PubKeys = make([]PubKey, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of interface
		out.PubKeys[_0] = DeepCopyPubKey(in.PubKeys[_0])
	}
	return
} //End of DeepCopyPubKeyMultisigThreshold

// Non-Interface
func EncodeSignedMsgType(w *[]byte, v SignedMsgType) {
	codonEncodeUint8(0, w, uint8(v))
} //End of EncodeSignedMsgType

func DecodeSignedMsgType(bz []byte) (v SignedMsgType, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			v = SignedMsgType(codonDecodeUint8(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeSignedMsgType

func RandSignedMsgType(r RandSrc) SignedMsgType {
	var v SignedMsgType
	v = SignedMsgType(r.GetUint8())
	return v
} //End of RandSignedMsgType

func DeepCopySignedMsgType(in SignedMsgType) (out SignedMsgType) {
	out = in
	return
} //End of DeepCopySignedMsgType

// Non-Interface
func EncodeVoteOption(w *[]byte, v VoteOption) {
	codonEncodeUint8(0, w, uint8(v))
} //End of EncodeVoteOption

func DecodeVoteOption(bz []byte) (v VoteOption, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			v = VoteOption(codonDecodeUint8(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeVoteOption

func RandVoteOption(r RandSrc) VoteOption {
	var v VoteOption
	v = VoteOption(r.GetUint8())
	return v
} //End of RandVoteOption

func DeepCopyVoteOption(in VoteOption) (out VoteOption) {
	out = in
	return
} //End of DeepCopyVoteOption

// Non-Interface
func EncodeVote(w *[]byte, v Vote) {
	codonEncodeUint8(0, w, uint8(v.Type))
	codonEncodeVarint(1, w, int64(v.Height))
	codonEncodeVarint(2, w, int64(v.Round))
	codonEncodeByteSlice(3, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeByteSlice(0, w, v.BlockID.Hash[:])
		codonEncodeByteSlice(1, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeVarint(0, w, int64(v.BlockID.PartsHeader.Total))
			codonEncodeByteSlice(1, w, v.BlockID.PartsHeader.Hash[:])
			return wBuf
		}()) // end of v.BlockID.PartsHeader
		return wBuf
	}()) // end of v.BlockID
	codonEncodeByteSlice(4, w, EncodeTime(v.Timestamp))
	codonEncodeByteSlice(5, w, v.ValidatorAddress[:])
	codonEncodeVarint(6, w, int64(v.ValidatorIndex))
	codonEncodeByteSlice(7, w, v.Signature[:])
} //End of EncodeVote

func DecodeVote(bz []byte) (v Vote, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Type
			v.Type = SignedMsgType(codonDecodeUint8(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Height
			v.Height = int64(codonDecodeInt64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 2: // v.Round
			v.Round = int(codonDecodeInt(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 3: // v.BlockID
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.BlockID.Hash
						var tmpBz []byte
						n, err = codonGetByteSlice(&tmpBz, bz)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						v.BlockID.Hash = tmpBz
					case 1: // v.BlockID.PartsHeader
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						func(bz []byte) {
							for len(bz) != 0 {
								tag := codonDecodeUint64(bz, &n, &err)
								if err != nil {
									return
								}
								bz = bz[n:]
								total += n
								tag = tag >> 3
								switch tag {
								case 0: // v.BlockID.PartsHeader.Total
									v.BlockID.PartsHeader.Total = int(codonDecodeInt(bz, &n, &err))
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
								case 1: // v.BlockID.PartsHeader.Hash
									var tmpBz []byte
									n, err = codonGetByteSlice(&tmpBz, bz)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									v.BlockID.PartsHeader.Hash = tmpBz
								default:
									err = errors.New("Unknown Field")
									return
								}
							} // end for
						}(bz[:l]) // end func
						if err != nil {
							return
						}
						bz = bz[l:]
						n += int(l)
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		case 4: // v.Timestamp
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.Timestamp, n, err = DecodeTime(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 5: // v.ValidatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorAddress = tmpBz
		case 6: // v.ValidatorIndex
			v.ValidatorIndex = int(codonDecodeInt(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 7: // v.Signature
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Signature = tmpBz
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeVote

func RandVote(r RandSrc) Vote {
	var length int
	var v Vote
	v.Type = SignedMsgType(r.GetUint8())
	v.Height = r.GetInt64()
	v.Round = r.GetInt()
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.BlockID.Hash = r.GetBytes(length)
	v.BlockID.PartsHeader.Total = r.GetInt()
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.BlockID.PartsHeader.Hash = r.GetBytes(length)
	// end of v.BlockID.PartsHeader
	// end of v.BlockID
	v.Timestamp = RandTime(r)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorAddress = r.GetBytes(length)
	v.ValidatorIndex = r.GetInt()
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Signature = r.GetBytes(length)
	return v
} //End of RandVote

func DeepCopyVote(in Vote) (out Vote) {
	var length int
	out.Type = in.Type
	out.Height = in.Height
	out.Round = in.Round
	length = len(in.BlockID.Hash)
	if length == 0 {
		out.BlockID.Hash = nil
	} else {
		out.BlockID.Hash = make([]uint8, length)
	}
	copy(out.BlockID.Hash[:], in.BlockID.Hash[:])
	out.BlockID.PartsHeader.Total = in.BlockID.PartsHeader.Total
	length = len(in.BlockID.PartsHeader.Hash)
	if length == 0 {
		out.BlockID.PartsHeader.Hash = nil
	} else {
		out.BlockID.PartsHeader.Hash = make([]uint8, length)
	}
	copy(out.BlockID.PartsHeader.Hash[:], in.BlockID.PartsHeader.Hash[:])
	// end of .BlockID.PartsHeader
	// end of .BlockID
	out.Timestamp = DeepCopyTime(in.Timestamp)
	length = len(in.ValidatorAddress)
	if length == 0 {
		out.ValidatorAddress = nil
	} else {
		out.ValidatorAddress = make([]uint8, length)
	}
	copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
	out.ValidatorIndex = in.ValidatorIndex
	length = len(in.Signature)
	if length == 0 {
		out.Signature = nil
	} else {
		out.Signature = make([]uint8, length)
	}
	copy(out.Signature[:], in.Signature[:])
	return
} //End of DeepCopyVote

// Non-Interface
func EncodeSdkInt(w *[]byte, v SdkInt) {
	codonEncodeByteSlice(0, w, EncodeInt(v))
} //End of EncodeSdkInt

func DecodeSdkInt(bz []byte) (v SdkInt, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v, n, err = DecodeInt(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeSdkInt

func RandSdkInt(r RandSrc) SdkInt {
	var v SdkInt
	v = RandInt(r)
	return v
} //End of RandSdkInt

func DeepCopySdkInt(in SdkInt) (out SdkInt) {
	out = DeepCopyInt(in)
	return
} //End of DeepCopySdkInt

// Non-Interface
func EncodeSdkDec(w *[]byte, v SdkDec) {
	codonEncodeByteSlice(0, w, EncodeDec(v))
} //End of EncodeSdkDec

func DecodeSdkDec(bz []byte) (v SdkDec, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v, n, err = DecodeDec(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeSdkDec

func RandSdkDec(r RandSrc) SdkDec {
	var v SdkDec
	v = RandDec(r)
	return v
} //End of RandSdkDec

func DeepCopySdkDec(in SdkDec) (out SdkDec) {
	out = DeepCopyDec(in)
	return
} //End of DeepCopySdkDec

// Non-Interface
func Encodeuint64(w *[]byte, v uint64) {
	codonEncodeUvarint(0, w, uint64(v))
} //End of Encodeuint64

func Decodeuint64(bz []byte) (v uint64, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			v = uint64(codonDecodeUint64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of Decodeuint64

func Randuint64(r RandSrc) uint64 {
	var v uint64
	v = r.GetUint64()
	return v
} //End of Randuint64

func DeepCopyuint64(in uint64) (out uint64) {
	out = in
	return
} //End of DeepCopyuint64

// Non-Interface
func Encodeint64(w *[]byte, v int64) {
	codonEncodeVarint(0, w, int64(v))
} //End of Encodeint64

func Decodeint64(bz []byte) (v int64, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			v = int64(codonDecodeInt64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of Decodeint64

func Randint64(r RandSrc) int64 {
	var v int64
	v = r.GetInt64()
	return v
} //End of Randint64

func DeepCopyint64(in int64) (out int64) {
	out = in
	return
} //End of DeepCopyint64

// Non-Interface
func EncodeConsAddress(w *[]byte, v ConsAddress) {
	codonEncodeByteSlice(0, w, v[:])
} //End of EncodeConsAddress

func DecodeConsAddress(bz []byte) (v ConsAddress, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v = tmpBz
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeConsAddress

func RandConsAddress(r RandSrc) ConsAddress {
	var length int
	var v ConsAddress
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v = r.GetBytes(length)
	return v
} //End of RandConsAddress

func DeepCopyConsAddress(in ConsAddress) (out ConsAddress) {
	var length int
	length = len(in)
	if length == 0 {
		out = nil
	} else {
		out = make([]uint8, length)
	}
	copy(out[:], in[:])
	return
} //End of DeepCopyConsAddress

// Non-Interface
func EncodeCoin(w *[]byte, v Coin) {
	codonEncodeString(0, w, v.Denom)
	codonEncodeByteSlice(1, w, EncodeInt(v.Amount))
} //End of EncodeCoin

func DecodeCoin(bz []byte) (v Coin, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Denom
			v.Denom = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Amount
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.Amount, n, err = DecodeInt(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeCoin

func RandCoin(r RandSrc) Coin {
	var v Coin
	v.Denom = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Amount = RandInt(r)
	return v
} //End of RandCoin

func DeepCopyCoin(in Coin) (out Coin) {
	out.Denom = in.Denom
	out.Amount = DeepCopyInt(in.Amount)
	return
} //End of DeepCopyCoin

// Non-Interface
func EncodeDecCoin(w *[]byte, v DecCoin) {
	codonEncodeString(0, w, v.Denom)
	codonEncodeByteSlice(1, w, EncodeDec(v.Amount))
} //End of EncodeDecCoin

func DecodeDecCoin(bz []byte) (v DecCoin, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Denom
			v.Denom = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Amount
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.Amount, n, err = DecodeDec(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeDecCoin

func RandDecCoin(r RandSrc) DecCoin {
	var v DecCoin
	v.Denom = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Amount = RandDec(r)
	return v
} //End of RandDecCoin

func DeepCopyDecCoin(in DecCoin) (out DecCoin) {
	out.Denom = in.Denom
	out.Amount = DeepCopyDec(in.Amount)
	return
} //End of DeepCopyDecCoin

// Non-Interface
func EncodeStdSignature(w *[]byte, v StdSignature) {
	codonEncodeByteSlice(0, w, func() []byte {
		w := make([]byte, 0, 64)
		EncodePubKey(&w, v.PubKey) // interface_encode
		return w
	}()) // end of v.PubKey
	codonEncodeByteSlice(1, w, v.Signature[:])
} //End of EncodeStdSignature

func DecodeStdSignature(bz []byte) (v StdSignature, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.PubKey
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.PubKey, n, err = DecodePubKey(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n // interface_decode
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 1: // v.Signature
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Signature = tmpBz
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeStdSignature

func RandStdSignature(r RandSrc) StdSignature {
	var length int
	var v StdSignature
	v.PubKey = RandPubKey(r) // interface_decode
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Signature = r.GetBytes(length)
	return v
} //End of RandStdSignature

func DeepCopyStdSignature(in StdSignature) (out StdSignature) {
	var length int
	out.PubKey = DeepCopyPubKey(in.PubKey)
	length = len(in.Signature)
	if length == 0 {
		out.Signature = nil
	} else {
		out.Signature = make([]uint8, length)
	}
	copy(out.Signature[:], in.Signature[:])
	return
} //End of DeepCopyStdSignature

// Non-Interface
func EncodeParamChange(w *[]byte, v ParamChange) {
	codonEncodeString(0, w, v.Subspace)
	codonEncodeString(1, w, v.Key)
	codonEncodeString(2, w, v.Subkey)
	codonEncodeString(3, w, v.Value)
} //End of EncodeParamChange

func DecodeParamChange(bz []byte) (v ParamChange, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Subspace
			v.Subspace = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Key
			v.Key = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 2: // v.Subkey
			v.Subkey = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 3: // v.Value
			v.Value = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeParamChange

func RandParamChange(r RandSrc) ParamChange {
	var v ParamChange
	v.Subspace = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Key = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Subkey = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Value = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	return v
} //End of RandParamChange

func DeepCopyParamChange(in ParamChange) (out ParamChange) {
	out.Subspace = in.Subspace
	out.Key = in.Key
	out.Subkey = in.Subkey
	out.Value = in.Value
	return
} //End of DeepCopyParamChange

// Non-Interface
func EncodeInput(w *[]byte, v Input) {
	codonEncodeByteSlice(0, w, v.Address[:])
	for _0 := 0; _0 < len(v.Coins); _0++ {
		codonEncodeByteSlice(1, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.Coins[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeInt(v.Coins[_0].Amount))
			return wBuf
		}()) // end of v.Coins[_0]
	}
} //End of EncodeInput

func DecodeInput(bz []byte) (v Input, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Address
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Address = tmpBz
		case 1: // v.Coins
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Coin
			tmp, n, err = DecodeCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Coins = append(v.Coins, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeInput

func RandInput(r RandSrc) Input {
	var length int
	var v Input
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Address = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Coins = nil
	} else {
		v.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Coins[_0] = RandCoin(r)
	}
	return v
} //End of RandInput

func DeepCopyInput(in Input) (out Input) {
	var length int
	length = len(in.Address)
	if length == 0 {
		out.Address = nil
	} else {
		out.Address = make([]uint8, length)
	}
	copy(out.Address[:], in.Address[:])
	length = len(in.Coins)
	if length == 0 {
		out.Coins = nil
	} else {
		out.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Coins[_0] = DeepCopyCoin(in.Coins[_0])
	}
	return
} //End of DeepCopyInput

// Non-Interface
func EncodeOutput(w *[]byte, v Output) {
	codonEncodeByteSlice(0, w, v.Address[:])
	for _0 := 0; _0 < len(v.Coins); _0++ {
		codonEncodeByteSlice(1, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.Coins[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeInt(v.Coins[_0].Amount))
			return wBuf
		}()) // end of v.Coins[_0]
	}
} //End of EncodeOutput

func DecodeOutput(bz []byte) (v Output, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Address
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Address = tmpBz
		case 1: // v.Coins
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Coin
			tmp, n, err = DecodeCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Coins = append(v.Coins, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeOutput

func RandOutput(r RandSrc) Output {
	var length int
	var v Output
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Address = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Coins = nil
	} else {
		v.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Coins[_0] = RandCoin(r)
	}
	return v
} //End of RandOutput

func DeepCopyOutput(in Output) (out Output) {
	var length int
	length = len(in.Address)
	if length == 0 {
		out.Address = nil
	} else {
		out.Address = make([]uint8, length)
	}
	copy(out.Address[:], in.Address[:])
	length = len(in.Coins)
	if length == 0 {
		out.Coins = nil
	} else {
		out.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Coins[_0] = DeepCopyCoin(in.Coins[_0])
	}
	return
} //End of DeepCopyOutput

// Non-Interface
func EncodeAccAddress(w *[]byte, v AccAddress) {
	codonEncodeByteSlice(0, w, v[:])
} //End of EncodeAccAddress

func DecodeAccAddress(bz []byte) (v AccAddress, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v = tmpBz
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeAccAddress

func RandAccAddress(r RandSrc) AccAddress {
	var length int
	var v AccAddress
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v = r.GetBytes(length)
	return v
} //End of RandAccAddress

func DeepCopyAccAddress(in AccAddress) (out AccAddress) {
	var length int
	length = len(in)
	if length == 0 {
		out = nil
	} else {
		out = make([]uint8, length)
	}
	copy(out[:], in[:])
	return
} //End of DeepCopyAccAddress

// Non-Interface
func EncodeBaseAccount(w *[]byte, v BaseAccount) {
	codonEncodeByteSlice(0, w, v.Address[:])
	for _0 := 0; _0 < len(v.Coins); _0++ {
		codonEncodeByteSlice(1, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.Coins[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeInt(v.Coins[_0].Amount))
			return wBuf
		}()) // end of v.Coins[_0]
	}
	codonEncodeByteSlice(2, w, func() []byte {
		w := make([]byte, 0, 64)
		EncodePubKey(&w, v.PubKey) // interface_encode
		return w
	}()) // end of v.PubKey
	codonEncodeUvarint(3, w, uint64(v.AccountNumber))
	codonEncodeUvarint(4, w, uint64(v.Sequence))
} //End of EncodeBaseAccount

func DecodeBaseAccount(bz []byte) (v BaseAccount, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Address
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Address = tmpBz
		case 1: // v.Coins
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Coin
			tmp, n, err = DecodeCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Coins = append(v.Coins, tmp)
		case 2: // v.PubKey
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.PubKey, n, err = DecodePubKey(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n // interface_decode
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 3: // v.AccountNumber
			v.AccountNumber = uint64(codonDecodeUint64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 4: // v.Sequence
			v.Sequence = uint64(codonDecodeUint64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeBaseAccount

func RandBaseAccount(r RandSrc) BaseAccount {
	var length int
	var v BaseAccount
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Address = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Coins = nil
	} else {
		v.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Coins[_0] = RandCoin(r)
	}
	v.PubKey = RandPubKey(r) // interface_decode
	v.AccountNumber = r.GetUint64()
	v.Sequence = r.GetUint64()
	return v
} //End of RandBaseAccount

func DeepCopyBaseAccount(in BaseAccount) (out BaseAccount) {
	var length int
	length = len(in.Address)
	if length == 0 {
		out.Address = nil
	} else {
		out.Address = make([]uint8, length)
	}
	copy(out.Address[:], in.Address[:])
	length = len(in.Coins)
	if length == 0 {
		out.Coins = nil
	} else {
		out.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Coins[_0] = DeepCopyCoin(in.Coins[_0])
	}
	out.PubKey = DeepCopyPubKey(in.PubKey)
	out.AccountNumber = in.AccountNumber
	out.Sequence = in.Sequence
	return
} //End of DeepCopyBaseAccount

// Non-Interface
func EncodeBaseVestingAccount(w *[]byte, v BaseVestingAccount) {
	codonEncodeByteSlice(0, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeByteSlice(0, w, v.BaseAccount.Address[:])
		for _0 := 0; _0 < len(v.BaseAccount.Coins); _0++ {
			codonEncodeByteSlice(1, w, func() []byte {
				wBuf := make([]byte, 0, 64)
				w := &wBuf
				codonEncodeString(0, w, v.BaseAccount.Coins[_0].Denom)
				codonEncodeByteSlice(1, w, EncodeInt(v.BaseAccount.Coins[_0].Amount))
				return wBuf
			}()) // end of v.BaseAccount.Coins[_0]
		}
		codonEncodeByteSlice(2, w, func() []byte {
			w := make([]byte, 0, 64)
			EncodePubKey(&w, v.BaseAccount.PubKey) // interface_encode
			return w
		}()) // end of v.BaseAccount.PubKey
		codonEncodeUvarint(3, w, uint64(v.BaseAccount.AccountNumber))
		codonEncodeUvarint(4, w, uint64(v.BaseAccount.Sequence))
		return wBuf
	}()) // end of v.BaseAccount
	for _0 := 0; _0 < len(v.OriginalVesting); _0++ {
		codonEncodeByteSlice(1, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.OriginalVesting[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeInt(v.OriginalVesting[_0].Amount))
			return wBuf
		}()) // end of v.OriginalVesting[_0]
	}
	for _0 := 0; _0 < len(v.DelegatedFree); _0++ {
		codonEncodeByteSlice(2, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.DelegatedFree[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeInt(v.DelegatedFree[_0].Amount))
			return wBuf
		}()) // end of v.DelegatedFree[_0]
	}
	for _0 := 0; _0 < len(v.DelegatedVesting); _0++ {
		codonEncodeByteSlice(3, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.DelegatedVesting[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeInt(v.DelegatedVesting[_0].Amount))
			return wBuf
		}()) // end of v.DelegatedVesting[_0]
	}
	codonEncodeVarint(4, w, int64(v.EndTime))
} //End of EncodeBaseVestingAccount

func DecodeBaseVestingAccount(bz []byte) (v BaseVestingAccount, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.BaseAccount
			v.BaseAccount = &BaseAccount{}
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.BaseAccount.Address
						var tmpBz []byte
						n, err = codonGetByteSlice(&tmpBz, bz)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						v.BaseAccount.Address = tmpBz
					case 1: // v.BaseAccount.Coins
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						var tmp Coin
						tmp, n, err = DecodeCoin(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
						v.BaseAccount.Coins = append(v.BaseAccount.Coins, tmp)
					case 2: // v.BaseAccount.PubKey
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						v.BaseAccount.PubKey, n, err = DecodePubKey(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n // interface_decode
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
					case 3: // v.BaseAccount.AccountNumber
						v.BaseAccount.AccountNumber = uint64(codonDecodeUint64(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 4: // v.BaseAccount.Sequence
						v.BaseAccount.Sequence = uint64(codonDecodeUint64(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		case 1: // v.OriginalVesting
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Coin
			tmp, n, err = DecodeCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.OriginalVesting = append(v.OriginalVesting, tmp)
		case 2: // v.DelegatedFree
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Coin
			tmp, n, err = DecodeCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.DelegatedFree = append(v.DelegatedFree, tmp)
		case 3: // v.DelegatedVesting
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Coin
			tmp, n, err = DecodeCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.DelegatedVesting = append(v.DelegatedVesting, tmp)
		case 4: // v.EndTime
			v.EndTime = int64(codonDecodeInt64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeBaseVestingAccount

func RandBaseVestingAccount(r RandSrc) BaseVestingAccount {
	var length int
	var v BaseVestingAccount
	v.BaseAccount = &BaseAccount{}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.BaseAccount.Address = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.BaseAccount.Coins = nil
	} else {
		v.BaseAccount.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.BaseAccount.Coins[_0] = RandCoin(r)
	}
	v.BaseAccount.PubKey = RandPubKey(r) // interface_decode
	v.BaseAccount.AccountNumber = r.GetUint64()
	v.BaseAccount.Sequence = r.GetUint64()
	// end of v.BaseAccount
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.OriginalVesting = nil
	} else {
		v.OriginalVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.OriginalVesting[_0] = RandCoin(r)
	}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.DelegatedFree = nil
	} else {
		v.DelegatedFree = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.DelegatedFree[_0] = RandCoin(r)
	}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.DelegatedVesting = nil
	} else {
		v.DelegatedVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.DelegatedVesting[_0] = RandCoin(r)
	}
	v.EndTime = r.GetInt64()
	return v
} //End of RandBaseVestingAccount

func DeepCopyBaseVestingAccount(in BaseVestingAccount) (out BaseVestingAccount) {
	var length int
	out.BaseAccount = &BaseAccount{}
	length = len(in.BaseAccount.Address)
	if length == 0 {
		out.BaseAccount.Address = nil
	} else {
		out.BaseAccount.Address = make([]uint8, length)
	}
	copy(out.BaseAccount.Address[:], in.BaseAccount.Address[:])
	length = len(in.BaseAccount.Coins)
	if length == 0 {
		out.BaseAccount.Coins = nil
	} else {
		out.BaseAccount.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.BaseAccount.Coins[_0] = DeepCopyCoin(in.BaseAccount.Coins[_0])
	}
	out.BaseAccount.PubKey = DeepCopyPubKey(in.BaseAccount.PubKey)
	out.BaseAccount.AccountNumber = in.BaseAccount.AccountNumber
	out.BaseAccount.Sequence = in.BaseAccount.Sequence
	// end of .BaseAccount
	length = len(in.OriginalVesting)
	if length == 0 {
		out.OriginalVesting = nil
	} else {
		out.OriginalVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.OriginalVesting[_0] = DeepCopyCoin(in.OriginalVesting[_0])
	}
	length = len(in.DelegatedFree)
	if length == 0 {
		out.DelegatedFree = nil
	} else {
		out.DelegatedFree = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.DelegatedFree[_0] = DeepCopyCoin(in.DelegatedFree[_0])
	}
	length = len(in.DelegatedVesting)
	if length == 0 {
		out.DelegatedVesting = nil
	} else {
		out.DelegatedVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.DelegatedVesting[_0] = DeepCopyCoin(in.DelegatedVesting[_0])
	}
	out.EndTime = in.EndTime
	return
} //End of DeepCopyBaseVestingAccount

// Non-Interface
func EncodeContinuousVestingAccount(w *[]byte, v ContinuousVestingAccount) {
	codonEncodeByteSlice(0, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeByteSlice(0, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeByteSlice(0, w, v.BaseVestingAccount.BaseAccount.Address[:])
			for _0 := 0; _0 < len(v.BaseVestingAccount.BaseAccount.Coins); _0++ {
				codonEncodeByteSlice(1, w, func() []byte {
					wBuf := make([]byte, 0, 64)
					w := &wBuf
					codonEncodeString(0, w, v.BaseVestingAccount.BaseAccount.Coins[_0].Denom)
					codonEncodeByteSlice(1, w, EncodeInt(v.BaseVestingAccount.BaseAccount.Coins[_0].Amount))
					return wBuf
				}()) // end of v.BaseVestingAccount.BaseAccount.Coins[_0]
			}
			codonEncodeByteSlice(2, w, func() []byte {
				w := make([]byte, 0, 64)
				EncodePubKey(&w, v.BaseVestingAccount.BaseAccount.PubKey) // interface_encode
				return w
			}()) // end of v.BaseVestingAccount.BaseAccount.PubKey
			codonEncodeUvarint(3, w, uint64(v.BaseVestingAccount.BaseAccount.AccountNumber))
			codonEncodeUvarint(4, w, uint64(v.BaseVestingAccount.BaseAccount.Sequence))
			return wBuf
		}()) // end of v.BaseVestingAccount.BaseAccount
		for _0 := 0; _0 < len(v.BaseVestingAccount.OriginalVesting); _0++ {
			codonEncodeByteSlice(1, w, func() []byte {
				wBuf := make([]byte, 0, 64)
				w := &wBuf
				codonEncodeString(0, w, v.BaseVestingAccount.OriginalVesting[_0].Denom)
				codonEncodeByteSlice(1, w, EncodeInt(v.BaseVestingAccount.OriginalVesting[_0].Amount))
				return wBuf
			}()) // end of v.BaseVestingAccount.OriginalVesting[_0]
		}
		for _0 := 0; _0 < len(v.BaseVestingAccount.DelegatedFree); _0++ {
			codonEncodeByteSlice(2, w, func() []byte {
				wBuf := make([]byte, 0, 64)
				w := &wBuf
				codonEncodeString(0, w, v.BaseVestingAccount.DelegatedFree[_0].Denom)
				codonEncodeByteSlice(1, w, EncodeInt(v.BaseVestingAccount.DelegatedFree[_0].Amount))
				return wBuf
			}()) // end of v.BaseVestingAccount.DelegatedFree[_0]
		}
		for _0 := 0; _0 < len(v.BaseVestingAccount.DelegatedVesting); _0++ {
			codonEncodeByteSlice(3, w, func() []byte {
				wBuf := make([]byte, 0, 64)
				w := &wBuf
				codonEncodeString(0, w, v.BaseVestingAccount.DelegatedVesting[_0].Denom)
				codonEncodeByteSlice(1, w, EncodeInt(v.BaseVestingAccount.DelegatedVesting[_0].Amount))
				return wBuf
			}()) // end of v.BaseVestingAccount.DelegatedVesting[_0]
		}
		codonEncodeVarint(4, w, int64(v.BaseVestingAccount.EndTime))
		return wBuf
	}()) // end of v.BaseVestingAccount
	codonEncodeVarint(1, w, int64(v.StartTime))
} //End of EncodeContinuousVestingAccount

func DecodeContinuousVestingAccount(bz []byte) (v ContinuousVestingAccount, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.BaseVestingAccount
			v.BaseVestingAccount = &BaseVestingAccount{}
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.BaseVestingAccount.BaseAccount
						v.BaseVestingAccount.BaseAccount = &BaseAccount{}
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						func(bz []byte) {
							for len(bz) != 0 {
								tag := codonDecodeUint64(bz, &n, &err)
								if err != nil {
									return
								}
								bz = bz[n:]
								total += n
								tag = tag >> 3
								switch tag {
								case 0: // v.BaseVestingAccount.BaseAccount.Address
									var tmpBz []byte
									n, err = codonGetByteSlice(&tmpBz, bz)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									v.BaseVestingAccount.BaseAccount.Address = tmpBz
								case 1: // v.BaseVestingAccount.BaseAccount.Coins
									l := codonDecodeUint64(bz, &n, &err)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) > len(bz) {
										err = errors.New("Length Too Large")
										return
									}
									var tmp Coin
									tmp, n, err = DecodeCoin(bz[:l])
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) != n {
										err = errors.New("Length Mismatch")
										return
									}
									v.BaseVestingAccount.BaseAccount.Coins = append(v.BaseVestingAccount.BaseAccount.Coins, tmp)
								case 2: // v.BaseVestingAccount.BaseAccount.PubKey
									l := codonDecodeUint64(bz, &n, &err)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) > len(bz) {
										err = errors.New("Length Too Large")
										return
									}
									v.BaseVestingAccount.BaseAccount.PubKey, n, err = DecodePubKey(bz[:l])
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n // interface_decode
									if int(l) != n {
										err = errors.New("Length Mismatch")
										return
									}
								case 3: // v.BaseVestingAccount.BaseAccount.AccountNumber
									v.BaseVestingAccount.BaseAccount.AccountNumber = uint64(codonDecodeUint64(bz, &n, &err))
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
								case 4: // v.BaseVestingAccount.BaseAccount.Sequence
									v.BaseVestingAccount.BaseAccount.Sequence = uint64(codonDecodeUint64(bz, &n, &err))
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
								default:
									err = errors.New("Unknown Field")
									return
								}
							} // end for
						}(bz[:l]) // end func
						if err != nil {
							return
						}
						bz = bz[l:]
						n += int(l)
					case 1: // v.BaseVestingAccount.OriginalVesting
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						var tmp Coin
						tmp, n, err = DecodeCoin(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
						v.BaseVestingAccount.OriginalVesting = append(v.BaseVestingAccount.OriginalVesting, tmp)
					case 2: // v.BaseVestingAccount.DelegatedFree
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						var tmp Coin
						tmp, n, err = DecodeCoin(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
						v.BaseVestingAccount.DelegatedFree = append(v.BaseVestingAccount.DelegatedFree, tmp)
					case 3: // v.BaseVestingAccount.DelegatedVesting
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						var tmp Coin
						tmp, n, err = DecodeCoin(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
						v.BaseVestingAccount.DelegatedVesting = append(v.BaseVestingAccount.DelegatedVesting, tmp)
					case 4: // v.BaseVestingAccount.EndTime
						v.BaseVestingAccount.EndTime = int64(codonDecodeInt64(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		case 1: // v.StartTime
			v.StartTime = int64(codonDecodeInt64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeContinuousVestingAccount

func RandContinuousVestingAccount(r RandSrc) ContinuousVestingAccount {
	var length int
	var v ContinuousVestingAccount
	v.BaseVestingAccount = &BaseVestingAccount{}
	v.BaseVestingAccount.BaseAccount = &BaseAccount{}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.BaseVestingAccount.BaseAccount.Address = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.BaseVestingAccount.BaseAccount.Coins = nil
	} else {
		v.BaseVestingAccount.BaseAccount.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.BaseVestingAccount.BaseAccount.Coins[_0] = RandCoin(r)
	}
	v.BaseVestingAccount.BaseAccount.PubKey = RandPubKey(r) // interface_decode
	v.BaseVestingAccount.BaseAccount.AccountNumber = r.GetUint64()
	v.BaseVestingAccount.BaseAccount.Sequence = r.GetUint64()
	// end of v.BaseVestingAccount.BaseAccount
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.BaseVestingAccount.OriginalVesting = nil
	} else {
		v.BaseVestingAccount.OriginalVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.BaseVestingAccount.OriginalVesting[_0] = RandCoin(r)
	}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.BaseVestingAccount.DelegatedFree = nil
	} else {
		v.BaseVestingAccount.DelegatedFree = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.BaseVestingAccount.DelegatedFree[_0] = RandCoin(r)
	}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.BaseVestingAccount.DelegatedVesting = nil
	} else {
		v.BaseVestingAccount.DelegatedVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.BaseVestingAccount.DelegatedVesting[_0] = RandCoin(r)
	}
	v.BaseVestingAccount.EndTime = r.GetInt64()
	// end of v.BaseVestingAccount
	v.StartTime = r.GetInt64()
	return v
} //End of RandContinuousVestingAccount

func DeepCopyContinuousVestingAccount(in ContinuousVestingAccount) (out ContinuousVestingAccount) {
	var length int
	out.BaseVestingAccount = &BaseVestingAccount{}
	out.BaseVestingAccount.BaseAccount = &BaseAccount{}
	length = len(in.BaseVestingAccount.BaseAccount.Address)
	if length == 0 {
		out.BaseVestingAccount.BaseAccount.Address = nil
	} else {
		out.BaseVestingAccount.BaseAccount.Address = make([]uint8, length)
	}
	copy(out.BaseVestingAccount.BaseAccount.Address[:], in.BaseVestingAccount.BaseAccount.Address[:])
	length = len(in.BaseVestingAccount.BaseAccount.Coins)
	if length == 0 {
		out.BaseVestingAccount.BaseAccount.Coins = nil
	} else {
		out.BaseVestingAccount.BaseAccount.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.BaseVestingAccount.BaseAccount.Coins[_0] = DeepCopyCoin(in.BaseVestingAccount.BaseAccount.Coins[_0])
	}
	out.BaseVestingAccount.BaseAccount.PubKey = DeepCopyPubKey(in.BaseVestingAccount.BaseAccount.PubKey)
	out.BaseVestingAccount.BaseAccount.AccountNumber = in.BaseVestingAccount.BaseAccount.AccountNumber
	out.BaseVestingAccount.BaseAccount.Sequence = in.BaseVestingAccount.BaseAccount.Sequence
	// end of .BaseVestingAccount.BaseAccount
	length = len(in.BaseVestingAccount.OriginalVesting)
	if length == 0 {
		out.BaseVestingAccount.OriginalVesting = nil
	} else {
		out.BaseVestingAccount.OriginalVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.BaseVestingAccount.OriginalVesting[_0] = DeepCopyCoin(in.BaseVestingAccount.OriginalVesting[_0])
	}
	length = len(in.BaseVestingAccount.DelegatedFree)
	if length == 0 {
		out.BaseVestingAccount.DelegatedFree = nil
	} else {
		out.BaseVestingAccount.DelegatedFree = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.BaseVestingAccount.DelegatedFree[_0] = DeepCopyCoin(in.BaseVestingAccount.DelegatedFree[_0])
	}
	length = len(in.BaseVestingAccount.DelegatedVesting)
	if length == 0 {
		out.BaseVestingAccount.DelegatedVesting = nil
	} else {
		out.BaseVestingAccount.DelegatedVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.BaseVestingAccount.DelegatedVesting[_0] = DeepCopyCoin(in.BaseVestingAccount.DelegatedVesting[_0])
	}
	out.BaseVestingAccount.EndTime = in.BaseVestingAccount.EndTime
	// end of .BaseVestingAccount
	out.StartTime = in.StartTime
	return
} //End of DeepCopyContinuousVestingAccount

// Non-Interface
func EncodeDelayedVestingAccount(w *[]byte, v DelayedVestingAccount) {
	codonEncodeByteSlice(0, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeByteSlice(0, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeByteSlice(0, w, v.BaseVestingAccount.BaseAccount.Address[:])
			for _0 := 0; _0 < len(v.BaseVestingAccount.BaseAccount.Coins); _0++ {
				codonEncodeByteSlice(1, w, func() []byte {
					wBuf := make([]byte, 0, 64)
					w := &wBuf
					codonEncodeString(0, w, v.BaseVestingAccount.BaseAccount.Coins[_0].Denom)
					codonEncodeByteSlice(1, w, EncodeInt(v.BaseVestingAccount.BaseAccount.Coins[_0].Amount))
					return wBuf
				}()) // end of v.BaseVestingAccount.BaseAccount.Coins[_0]
			}
			codonEncodeByteSlice(2, w, func() []byte {
				w := make([]byte, 0, 64)
				EncodePubKey(&w, v.BaseVestingAccount.BaseAccount.PubKey) // interface_encode
				return w
			}()) // end of v.BaseVestingAccount.BaseAccount.PubKey
			codonEncodeUvarint(3, w, uint64(v.BaseVestingAccount.BaseAccount.AccountNumber))
			codonEncodeUvarint(4, w, uint64(v.BaseVestingAccount.BaseAccount.Sequence))
			return wBuf
		}()) // end of v.BaseVestingAccount.BaseAccount
		for _0 := 0; _0 < len(v.BaseVestingAccount.OriginalVesting); _0++ {
			codonEncodeByteSlice(1, w, func() []byte {
				wBuf := make([]byte, 0, 64)
				w := &wBuf
				codonEncodeString(0, w, v.BaseVestingAccount.OriginalVesting[_0].Denom)
				codonEncodeByteSlice(1, w, EncodeInt(v.BaseVestingAccount.OriginalVesting[_0].Amount))
				return wBuf
			}()) // end of v.BaseVestingAccount.OriginalVesting[_0]
		}
		for _0 := 0; _0 < len(v.BaseVestingAccount.DelegatedFree); _0++ {
			codonEncodeByteSlice(2, w, func() []byte {
				wBuf := make([]byte, 0, 64)
				w := &wBuf
				codonEncodeString(0, w, v.BaseVestingAccount.DelegatedFree[_0].Denom)
				codonEncodeByteSlice(1, w, EncodeInt(v.BaseVestingAccount.DelegatedFree[_0].Amount))
				return wBuf
			}()) // end of v.BaseVestingAccount.DelegatedFree[_0]
		}
		for _0 := 0; _0 < len(v.BaseVestingAccount.DelegatedVesting); _0++ {
			codonEncodeByteSlice(3, w, func() []byte {
				wBuf := make([]byte, 0, 64)
				w := &wBuf
				codonEncodeString(0, w, v.BaseVestingAccount.DelegatedVesting[_0].Denom)
				codonEncodeByteSlice(1, w, EncodeInt(v.BaseVestingAccount.DelegatedVesting[_0].Amount))
				return wBuf
			}()) // end of v.BaseVestingAccount.DelegatedVesting[_0]
		}
		codonEncodeVarint(4, w, int64(v.BaseVestingAccount.EndTime))
		return wBuf
	}()) // end of v.BaseVestingAccount
} //End of EncodeDelayedVestingAccount

func DecodeDelayedVestingAccount(bz []byte) (v DelayedVestingAccount, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.BaseVestingAccount
			v.BaseVestingAccount = &BaseVestingAccount{}
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.BaseVestingAccount.BaseAccount
						v.BaseVestingAccount.BaseAccount = &BaseAccount{}
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						func(bz []byte) {
							for len(bz) != 0 {
								tag := codonDecodeUint64(bz, &n, &err)
								if err != nil {
									return
								}
								bz = bz[n:]
								total += n
								tag = tag >> 3
								switch tag {
								case 0: // v.BaseVestingAccount.BaseAccount.Address
									var tmpBz []byte
									n, err = codonGetByteSlice(&tmpBz, bz)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									v.BaseVestingAccount.BaseAccount.Address = tmpBz
								case 1: // v.BaseVestingAccount.BaseAccount.Coins
									l := codonDecodeUint64(bz, &n, &err)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) > len(bz) {
										err = errors.New("Length Too Large")
										return
									}
									var tmp Coin
									tmp, n, err = DecodeCoin(bz[:l])
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) != n {
										err = errors.New("Length Mismatch")
										return
									}
									v.BaseVestingAccount.BaseAccount.Coins = append(v.BaseVestingAccount.BaseAccount.Coins, tmp)
								case 2: // v.BaseVestingAccount.BaseAccount.PubKey
									l := codonDecodeUint64(bz, &n, &err)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) > len(bz) {
										err = errors.New("Length Too Large")
										return
									}
									v.BaseVestingAccount.BaseAccount.PubKey, n, err = DecodePubKey(bz[:l])
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n // interface_decode
									if int(l) != n {
										err = errors.New("Length Mismatch")
										return
									}
								case 3: // v.BaseVestingAccount.BaseAccount.AccountNumber
									v.BaseVestingAccount.BaseAccount.AccountNumber = uint64(codonDecodeUint64(bz, &n, &err))
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
								case 4: // v.BaseVestingAccount.BaseAccount.Sequence
									v.BaseVestingAccount.BaseAccount.Sequence = uint64(codonDecodeUint64(bz, &n, &err))
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
								default:
									err = errors.New("Unknown Field")
									return
								}
							} // end for
						}(bz[:l]) // end func
						if err != nil {
							return
						}
						bz = bz[l:]
						n += int(l)
					case 1: // v.BaseVestingAccount.OriginalVesting
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						var tmp Coin
						tmp, n, err = DecodeCoin(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
						v.BaseVestingAccount.OriginalVesting = append(v.BaseVestingAccount.OriginalVesting, tmp)
					case 2: // v.BaseVestingAccount.DelegatedFree
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						var tmp Coin
						tmp, n, err = DecodeCoin(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
						v.BaseVestingAccount.DelegatedFree = append(v.BaseVestingAccount.DelegatedFree, tmp)
					case 3: // v.BaseVestingAccount.DelegatedVesting
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						var tmp Coin
						tmp, n, err = DecodeCoin(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
						v.BaseVestingAccount.DelegatedVesting = append(v.BaseVestingAccount.DelegatedVesting, tmp)
					case 4: // v.BaseVestingAccount.EndTime
						v.BaseVestingAccount.EndTime = int64(codonDecodeInt64(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeDelayedVestingAccount

func RandDelayedVestingAccount(r RandSrc) DelayedVestingAccount {
	var length int
	var v DelayedVestingAccount
	v.BaseVestingAccount = &BaseVestingAccount{}
	v.BaseVestingAccount.BaseAccount = &BaseAccount{}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.BaseVestingAccount.BaseAccount.Address = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.BaseVestingAccount.BaseAccount.Coins = nil
	} else {
		v.BaseVestingAccount.BaseAccount.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.BaseVestingAccount.BaseAccount.Coins[_0] = RandCoin(r)
	}
	v.BaseVestingAccount.BaseAccount.PubKey = RandPubKey(r) // interface_decode
	v.BaseVestingAccount.BaseAccount.AccountNumber = r.GetUint64()
	v.BaseVestingAccount.BaseAccount.Sequence = r.GetUint64()
	// end of v.BaseVestingAccount.BaseAccount
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.BaseVestingAccount.OriginalVesting = nil
	} else {
		v.BaseVestingAccount.OriginalVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.BaseVestingAccount.OriginalVesting[_0] = RandCoin(r)
	}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.BaseVestingAccount.DelegatedFree = nil
	} else {
		v.BaseVestingAccount.DelegatedFree = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.BaseVestingAccount.DelegatedFree[_0] = RandCoin(r)
	}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.BaseVestingAccount.DelegatedVesting = nil
	} else {
		v.BaseVestingAccount.DelegatedVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.BaseVestingAccount.DelegatedVesting[_0] = RandCoin(r)
	}
	v.BaseVestingAccount.EndTime = r.GetInt64()
	// end of v.BaseVestingAccount
	return v
} //End of RandDelayedVestingAccount

func DeepCopyDelayedVestingAccount(in DelayedVestingAccount) (out DelayedVestingAccount) {
	var length int
	out.BaseVestingAccount = &BaseVestingAccount{}
	out.BaseVestingAccount.BaseAccount = &BaseAccount{}
	length = len(in.BaseVestingAccount.BaseAccount.Address)
	if length == 0 {
		out.BaseVestingAccount.BaseAccount.Address = nil
	} else {
		out.BaseVestingAccount.BaseAccount.Address = make([]uint8, length)
	}
	copy(out.BaseVestingAccount.BaseAccount.Address[:], in.BaseVestingAccount.BaseAccount.Address[:])
	length = len(in.BaseVestingAccount.BaseAccount.Coins)
	if length == 0 {
		out.BaseVestingAccount.BaseAccount.Coins = nil
	} else {
		out.BaseVestingAccount.BaseAccount.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.BaseVestingAccount.BaseAccount.Coins[_0] = DeepCopyCoin(in.BaseVestingAccount.BaseAccount.Coins[_0])
	}
	out.BaseVestingAccount.BaseAccount.PubKey = DeepCopyPubKey(in.BaseVestingAccount.BaseAccount.PubKey)
	out.BaseVestingAccount.BaseAccount.AccountNumber = in.BaseVestingAccount.BaseAccount.AccountNumber
	out.BaseVestingAccount.BaseAccount.Sequence = in.BaseVestingAccount.BaseAccount.Sequence
	// end of .BaseVestingAccount.BaseAccount
	length = len(in.BaseVestingAccount.OriginalVesting)
	if length == 0 {
		out.BaseVestingAccount.OriginalVesting = nil
	} else {
		out.BaseVestingAccount.OriginalVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.BaseVestingAccount.OriginalVesting[_0] = DeepCopyCoin(in.BaseVestingAccount.OriginalVesting[_0])
	}
	length = len(in.BaseVestingAccount.DelegatedFree)
	if length == 0 {
		out.BaseVestingAccount.DelegatedFree = nil
	} else {
		out.BaseVestingAccount.DelegatedFree = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.BaseVestingAccount.DelegatedFree[_0] = DeepCopyCoin(in.BaseVestingAccount.DelegatedFree[_0])
	}
	length = len(in.BaseVestingAccount.DelegatedVesting)
	if length == 0 {
		out.BaseVestingAccount.DelegatedVesting = nil
	} else {
		out.BaseVestingAccount.DelegatedVesting = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.BaseVestingAccount.DelegatedVesting[_0] = DeepCopyCoin(in.BaseVestingAccount.DelegatedVesting[_0])
	}
	out.BaseVestingAccount.EndTime = in.BaseVestingAccount.EndTime
	// end of .BaseVestingAccount
	return
} //End of DeepCopyDelayedVestingAccount

// Non-Interface
func EncodeModuleAccount(w *[]byte, v ModuleAccount) {
	codonEncodeByteSlice(0, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeByteSlice(0, w, v.BaseAccount.Address[:])
		for _0 := 0; _0 < len(v.BaseAccount.Coins); _0++ {
			codonEncodeByteSlice(1, w, func() []byte {
				wBuf := make([]byte, 0, 64)
				w := &wBuf
				codonEncodeString(0, w, v.BaseAccount.Coins[_0].Denom)
				codonEncodeByteSlice(1, w, EncodeInt(v.BaseAccount.Coins[_0].Amount))
				return wBuf
			}()) // end of v.BaseAccount.Coins[_0]
		}
		codonEncodeByteSlice(2, w, func() []byte {
			w := make([]byte, 0, 64)
			EncodePubKey(&w, v.BaseAccount.PubKey) // interface_encode
			return w
		}()) // end of v.BaseAccount.PubKey
		codonEncodeUvarint(3, w, uint64(v.BaseAccount.AccountNumber))
		codonEncodeUvarint(4, w, uint64(v.BaseAccount.Sequence))
		return wBuf
	}()) // end of v.BaseAccount
	codonEncodeString(1, w, v.Name)
	for _0 := 0; _0 < len(v.Permissions); _0++ {
		codonEncodeString(2, w, v.Permissions[_0])
	}
} //End of EncodeModuleAccount

func DecodeModuleAccount(bz []byte) (v ModuleAccount, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.BaseAccount
			v.BaseAccount = &BaseAccount{}
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.BaseAccount.Address
						var tmpBz []byte
						n, err = codonGetByteSlice(&tmpBz, bz)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						v.BaseAccount.Address = tmpBz
					case 1: // v.BaseAccount.Coins
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						var tmp Coin
						tmp, n, err = DecodeCoin(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
						v.BaseAccount.Coins = append(v.BaseAccount.Coins, tmp)
					case 2: // v.BaseAccount.PubKey
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						v.BaseAccount.PubKey, n, err = DecodePubKey(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n // interface_decode
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
					case 3: // v.BaseAccount.AccountNumber
						v.BaseAccount.AccountNumber = uint64(codonDecodeUint64(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 4: // v.BaseAccount.Sequence
						v.BaseAccount.Sequence = uint64(codonDecodeUint64(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		case 1: // v.Name
			v.Name = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 2: // v.Permissions
			var tmp string
			tmp = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Permissions = append(v.Permissions, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeModuleAccount

func RandModuleAccount(r RandSrc) ModuleAccount {
	var length int
	var v ModuleAccount
	v.BaseAccount = &BaseAccount{}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.BaseAccount.Address = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.BaseAccount.Coins = nil
	} else {
		v.BaseAccount.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.BaseAccount.Coins[_0] = RandCoin(r)
	}
	v.BaseAccount.PubKey = RandPubKey(r) // interface_decode
	v.BaseAccount.AccountNumber = r.GetUint64()
	v.BaseAccount.Sequence = r.GetUint64()
	// end of v.BaseAccount
	v.Name = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Permissions = nil
	} else {
		v.Permissions = make([]string, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of string
		v.Permissions[_0] = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	}
	return v
} //End of RandModuleAccount

func DeepCopyModuleAccount(in ModuleAccount) (out ModuleAccount) {
	var length int
	out.BaseAccount = &BaseAccount{}
	length = len(in.BaseAccount.Address)
	if length == 0 {
		out.BaseAccount.Address = nil
	} else {
		out.BaseAccount.Address = make([]uint8, length)
	}
	copy(out.BaseAccount.Address[:], in.BaseAccount.Address[:])
	length = len(in.BaseAccount.Coins)
	if length == 0 {
		out.BaseAccount.Coins = nil
	} else {
		out.BaseAccount.Coins = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.BaseAccount.Coins[_0] = DeepCopyCoin(in.BaseAccount.Coins[_0])
	}
	out.BaseAccount.PubKey = DeepCopyPubKey(in.BaseAccount.PubKey)
	out.BaseAccount.AccountNumber = in.BaseAccount.AccountNumber
	out.BaseAccount.Sequence = in.BaseAccount.Sequence
	// end of .BaseAccount
	out.Name = in.Name
	length = len(in.Permissions)
	if length == 0 {
		out.Permissions = nil
	} else {
		out.Permissions = make([]string, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of string
		out.Permissions[_0] = in.Permissions[_0]
	}
	return
} //End of DeepCopyModuleAccount

// Non-Interface
func EncodeStdTx(w *[]byte, v StdTx) {
	for _0 := 0; _0 < len(v.Msgs); _0++ {
		codonEncodeByteSlice(0, w, func() []byte {
			w := make([]byte, 0, 64)
			EncodeMsg(&w, v.Msgs[_0]) // interface_encode
			return w
		}()) // end of v.Msgs[_0]
	}
	codonEncodeByteSlice(1, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		for _0 := 0; _0 < len(v.Fee.Amount); _0++ {
			codonEncodeByteSlice(0, w, func() []byte {
				wBuf := make([]byte, 0, 64)
				w := &wBuf
				codonEncodeString(0, w, v.Fee.Amount[_0].Denom)
				codonEncodeByteSlice(1, w, EncodeInt(v.Fee.Amount[_0].Amount))
				return wBuf
			}()) // end of v.Fee.Amount[_0]
		}
		codonEncodeUvarint(1, w, uint64(v.Fee.Gas))
		return wBuf
	}()) // end of v.Fee
	for _0 := 0; _0 < len(v.Signatures); _0++ {
		codonEncodeByteSlice(2, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeByteSlice(0, w, func() []byte {
				w := make([]byte, 0, 64)
				EncodePubKey(&w, v.Signatures[_0].PubKey) // interface_encode
				return w
			}()) // end of v.Signatures[_0].PubKey
			codonEncodeByteSlice(1, w, v.Signatures[_0].Signature[:])
			return wBuf
		}()) // end of v.Signatures[_0]
	}
	codonEncodeString(3, w, v.Memo)
} //End of EncodeStdTx

func DecodeStdTx(bz []byte) (v StdTx, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Msgs
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Msg
			tmp, n, err = DecodeMsg(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Msgs = append(v.Msgs, tmp)
		case 1: // v.Fee
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Fee.Amount
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						var tmp Coin
						tmp, n, err = DecodeCoin(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
						v.Fee.Amount = append(v.Fee.Amount, tmp)
					case 1: // v.Fee.Gas
						v.Fee.Gas = uint64(codonDecodeUint64(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		case 2: // v.Signatures
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp StdSignature
			tmp, n, err = DecodeStdSignature(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Signatures = append(v.Signatures, tmp)
		case 3: // v.Memo
			v.Memo = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeStdTx

func RandStdTx(r RandSrc) StdTx {
	var length int
	var v StdTx
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Msgs = nil
	} else {
		v.Msgs = make([]Msg, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of interface
		v.Msgs[_0] = RandMsg(r)
	}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Fee.Amount = nil
	} else {
		v.Fee.Amount = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Fee.Amount[_0] = RandCoin(r)
	}
	v.Fee.Gas = r.GetUint64()
	// end of v.Fee
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Signatures = nil
	} else {
		v.Signatures = make([]StdSignature, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Signatures[_0] = RandStdSignature(r)
	}
	v.Memo = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	return v
} //End of RandStdTx

func DeepCopyStdTx(in StdTx) (out StdTx) {
	var length int
	length = len(in.Msgs)
	if length == 0 {
		out.Msgs = nil
	} else {
		out.Msgs = make([]Msg, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of interface
		out.Msgs[_0] = DeepCopyMsg(in.Msgs[_0])
	}
	length = len(in.Fee.Amount)
	if length == 0 {
		out.Fee.Amount = nil
	} else {
		out.Fee.Amount = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Fee.Amount[_0] = DeepCopyCoin(in.Fee.Amount[_0])
	}
	out.Fee.Gas = in.Fee.Gas
	// end of .Fee
	length = len(in.Signatures)
	if length == 0 {
		out.Signatures = nil
	} else {
		out.Signatures = make([]StdSignature, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Signatures[_0] = DeepCopyStdSignature(in.Signatures[_0])
	}
	out.Memo = in.Memo
	return
} //End of DeepCopyStdTx

// Non-Interface
func EncodeMsgBeginRedelegate(w *[]byte, v MsgBeginRedelegate) {
	codonEncodeByteSlice(0, w, v.DelegatorAddress[:])
	codonEncodeByteSlice(1, w, v.ValidatorSrcAddress[:])
	codonEncodeByteSlice(2, w, v.ValidatorDstAddress[:])
	codonEncodeByteSlice(3, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeString(0, w, v.Amount.Denom)
		codonEncodeByteSlice(1, w, EncodeInt(v.Amount.Amount))
		return wBuf
	}()) // end of v.Amount
} //End of EncodeMsgBeginRedelegate

func DecodeMsgBeginRedelegate(bz []byte) (v MsgBeginRedelegate, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.DelegatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.DelegatorAddress = tmpBz
		case 1: // v.ValidatorSrcAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorSrcAddress = tmpBz
		case 2: // v.ValidatorDstAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorDstAddress = tmpBz
		case 3: // v.Amount
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Amount.Denom
						v.Amount.Denom = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 1: // v.Amount.Amount
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						v.Amount.Amount, n, err = DecodeInt(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgBeginRedelegate

func RandMsgBeginRedelegate(r RandSrc) MsgBeginRedelegate {
	var length int
	var v MsgBeginRedelegate
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.DelegatorAddress = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorSrcAddress = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorDstAddress = r.GetBytes(length)
	v.Amount.Denom = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Amount.Amount = RandInt(r)
	// end of v.Amount
	return v
} //End of RandMsgBeginRedelegate

func DeepCopyMsgBeginRedelegate(in MsgBeginRedelegate) (out MsgBeginRedelegate) {
	var length int
	length = len(in.DelegatorAddress)
	if length == 0 {
		out.DelegatorAddress = nil
	} else {
		out.DelegatorAddress = make([]uint8, length)
	}
	copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
	length = len(in.ValidatorSrcAddress)
	if length == 0 {
		out.ValidatorSrcAddress = nil
	} else {
		out.ValidatorSrcAddress = make([]uint8, length)
	}
	copy(out.ValidatorSrcAddress[:], in.ValidatorSrcAddress[:])
	length = len(in.ValidatorDstAddress)
	if length == 0 {
		out.ValidatorDstAddress = nil
	} else {
		out.ValidatorDstAddress = make([]uint8, length)
	}
	copy(out.ValidatorDstAddress[:], in.ValidatorDstAddress[:])
	out.Amount.Denom = in.Amount.Denom
	out.Amount.Amount = DeepCopyInt(in.Amount.Amount)
	// end of .Amount
	return
} //End of DeepCopyMsgBeginRedelegate

// Non-Interface
func EncodeMsgCreateValidator(w *[]byte, v MsgCreateValidator) {
	codonEncodeByteSlice(0, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeString(0, w, v.Description.Moniker)
		codonEncodeString(1, w, v.Description.Identity)
		codonEncodeString(2, w, v.Description.Website)
		codonEncodeString(3, w, v.Description.Details)
		return wBuf
	}()) // end of v.Description
	codonEncodeByteSlice(1, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeByteSlice(0, w, EncodeDec(v.Commission.Rate))
		codonEncodeByteSlice(1, w, EncodeDec(v.Commission.MaxRate))
		codonEncodeByteSlice(2, w, EncodeDec(v.Commission.MaxChangeRate))
		return wBuf
	}()) // end of v.Commission
	codonEncodeByteSlice(2, w, EncodeInt(v.MinSelfDelegation))
	codonEncodeByteSlice(3, w, v.DelegatorAddress[:])
	codonEncodeByteSlice(4, w, v.ValidatorAddress[:])
	codonEncodeByteSlice(5, w, func() []byte {
		w := make([]byte, 0, 64)
		EncodePubKey(&w, v.PubKey) // interface_encode
		return w
	}()) // end of v.PubKey
	codonEncodeByteSlice(6, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeString(0, w, v.Value.Denom)
		codonEncodeByteSlice(1, w, EncodeInt(v.Value.Amount))
		return wBuf
	}()) // end of v.Value
} //End of EncodeMsgCreateValidator

func DecodeMsgCreateValidator(bz []byte) (v MsgCreateValidator, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Description
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Description.Moniker
						v.Description.Moniker = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 1: // v.Description.Identity
						v.Description.Identity = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 2: // v.Description.Website
						v.Description.Website = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 3: // v.Description.Details
						v.Description.Details = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		case 1: // v.Commission
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Commission.Rate
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						v.Commission.Rate, n, err = DecodeDec(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
					case 1: // v.Commission.MaxRate
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						v.Commission.MaxRate, n, err = DecodeDec(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
					case 2: // v.Commission.MaxChangeRate
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						v.Commission.MaxChangeRate, n, err = DecodeDec(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		case 2: // v.MinSelfDelegation
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.MinSelfDelegation, n, err = DecodeInt(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 3: // v.DelegatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.DelegatorAddress = tmpBz
		case 4: // v.ValidatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorAddress = tmpBz
		case 5: // v.PubKey
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.PubKey, n, err = DecodePubKey(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n // interface_decode
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 6: // v.Value
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Value.Denom
						v.Value.Denom = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 1: // v.Value.Amount
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						v.Value.Amount, n, err = DecodeInt(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgCreateValidator

func RandMsgCreateValidator(r RandSrc) MsgCreateValidator {
	var length int
	var v MsgCreateValidator
	v.Description.Moniker = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description.Identity = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description.Website = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description.Details = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	// end of v.Description
	v.Commission.Rate = RandDec(r)
	v.Commission.MaxRate = RandDec(r)
	v.Commission.MaxChangeRate = RandDec(r)
	// end of v.Commission
	v.MinSelfDelegation = RandInt(r)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.DelegatorAddress = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorAddress = r.GetBytes(length)
	v.PubKey = RandPubKey(r) // interface_decode
	v.Value.Denom = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Value.Amount = RandInt(r)
	// end of v.Value
	return v
} //End of RandMsgCreateValidator

func DeepCopyMsgCreateValidator(in MsgCreateValidator) (out MsgCreateValidator) {
	var length int
	out.Description.Moniker = in.Description.Moniker
	out.Description.Identity = in.Description.Identity
	out.Description.Website = in.Description.Website
	out.Description.Details = in.Description.Details
	// end of .Description
	out.Commission.Rate = DeepCopyDec(in.Commission.Rate)
	out.Commission.MaxRate = DeepCopyDec(in.Commission.MaxRate)
	out.Commission.MaxChangeRate = DeepCopyDec(in.Commission.MaxChangeRate)
	// end of .Commission
	out.MinSelfDelegation = DeepCopyInt(in.MinSelfDelegation)
	length = len(in.DelegatorAddress)
	if length == 0 {
		out.DelegatorAddress = nil
	} else {
		out.DelegatorAddress = make([]uint8, length)
	}
	copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
	length = len(in.ValidatorAddress)
	if length == 0 {
		out.ValidatorAddress = nil
	} else {
		out.ValidatorAddress = make([]uint8, length)
	}
	copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
	out.PubKey = DeepCopyPubKey(in.PubKey)
	out.Value.Denom = in.Value.Denom
	out.Value.Amount = DeepCopyInt(in.Value.Amount)
	// end of .Value
	return
} //End of DeepCopyMsgCreateValidator

// Non-Interface
func EncodeMsgDelegate(w *[]byte, v MsgDelegate) {
	codonEncodeByteSlice(0, w, v.DelegatorAddress[:])
	codonEncodeByteSlice(1, w, v.ValidatorAddress[:])
	codonEncodeByteSlice(2, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeString(0, w, v.Amount.Denom)
		codonEncodeByteSlice(1, w, EncodeInt(v.Amount.Amount))
		return wBuf
	}()) // end of v.Amount
} //End of EncodeMsgDelegate

func DecodeMsgDelegate(bz []byte) (v MsgDelegate, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.DelegatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.DelegatorAddress = tmpBz
		case 1: // v.ValidatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorAddress = tmpBz
		case 2: // v.Amount
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Amount.Denom
						v.Amount.Denom = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 1: // v.Amount.Amount
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						v.Amount.Amount, n, err = DecodeInt(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgDelegate

func RandMsgDelegate(r RandSrc) MsgDelegate {
	var length int
	var v MsgDelegate
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.DelegatorAddress = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorAddress = r.GetBytes(length)
	v.Amount.Denom = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Amount.Amount = RandInt(r)
	// end of v.Amount
	return v
} //End of RandMsgDelegate

func DeepCopyMsgDelegate(in MsgDelegate) (out MsgDelegate) {
	var length int
	length = len(in.DelegatorAddress)
	if length == 0 {
		out.DelegatorAddress = nil
	} else {
		out.DelegatorAddress = make([]uint8, length)
	}
	copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
	length = len(in.ValidatorAddress)
	if length == 0 {
		out.ValidatorAddress = nil
	} else {
		out.ValidatorAddress = make([]uint8, length)
	}
	copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
	out.Amount.Denom = in.Amount.Denom
	out.Amount.Amount = DeepCopyInt(in.Amount.Amount)
	// end of .Amount
	return
} //End of DeepCopyMsgDelegate

// Non-Interface
func EncodeMsgEditValidator(w *[]byte, v MsgEditValidator) {
	codonEncodeByteSlice(0, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeString(0, w, v.Description.Moniker)
		codonEncodeString(1, w, v.Description.Identity)
		codonEncodeString(2, w, v.Description.Website)
		codonEncodeString(3, w, v.Description.Details)
		return wBuf
	}()) // end of v.Description
	codonEncodeByteSlice(1, w, v.ValidatorAddress[:])
	codonEncodeByteSlice(2, w, EncodeDec(*(v.CommissionRate)))
	codonEncodeByteSlice(3, w, EncodeInt(*(v.MinSelfDelegation)))
} //End of EncodeMsgEditValidator

func DecodeMsgEditValidator(bz []byte) (v MsgEditValidator, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Description
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Description.Moniker
						v.Description.Moniker = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 1: // v.Description.Identity
						v.Description.Identity = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 2: // v.Description.Website
						v.Description.Website = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 3: // v.Description.Details
						v.Description.Details = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		case 1: // v.ValidatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorAddress = tmpBz
		case 2: // v.CommissionRate
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.CommissionRate = &SdkDec{}
			*(v.CommissionRate), n, err = DecodeDec(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 3: // v.MinSelfDelegation
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.MinSelfDelegation = &SdkInt{}
			*(v.MinSelfDelegation), n, err = DecodeInt(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgEditValidator

func RandMsgEditValidator(r RandSrc) MsgEditValidator {
	var length int
	var v MsgEditValidator
	v.Description.Moniker = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description.Identity = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description.Website = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description.Details = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	// end of v.Description
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorAddress = r.GetBytes(length)
	v.CommissionRate = &SdkDec{}
	*(v.CommissionRate) = RandDec(r)
	v.MinSelfDelegation = &SdkInt{}
	*(v.MinSelfDelegation) = RandInt(r)
	return v
} //End of RandMsgEditValidator

func DeepCopyMsgEditValidator(in MsgEditValidator) (out MsgEditValidator) {
	var length int
	out.Description.Moniker = in.Description.Moniker
	out.Description.Identity = in.Description.Identity
	out.Description.Website = in.Description.Website
	out.Description.Details = in.Description.Details
	// end of .Description
	length = len(in.ValidatorAddress)
	if length == 0 {
		out.ValidatorAddress = nil
	} else {
		out.ValidatorAddress = make([]uint8, length)
	}
	copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
	out.CommissionRate = &SdkDec{}
	*(out.CommissionRate) = DeepCopyDec(*(in.CommissionRate))
	out.MinSelfDelegation = &SdkInt{}
	*(out.MinSelfDelegation) = DeepCopyInt(*(in.MinSelfDelegation))
	return
} //End of DeepCopyMsgEditValidator

// Non-Interface
func EncodeMsgSetWithdrawAddress(w *[]byte, v MsgSetWithdrawAddress) {
	codonEncodeByteSlice(0, w, v.DelegatorAddress[:])
	codonEncodeByteSlice(1, w, v.WithdrawAddress[:])
} //End of EncodeMsgSetWithdrawAddress

func DecodeMsgSetWithdrawAddress(bz []byte) (v MsgSetWithdrawAddress, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.DelegatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.DelegatorAddress = tmpBz
		case 1: // v.WithdrawAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.WithdrawAddress = tmpBz
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgSetWithdrawAddress

func RandMsgSetWithdrawAddress(r RandSrc) MsgSetWithdrawAddress {
	var length int
	var v MsgSetWithdrawAddress
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.DelegatorAddress = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.WithdrawAddress = r.GetBytes(length)
	return v
} //End of RandMsgSetWithdrawAddress

func DeepCopyMsgSetWithdrawAddress(in MsgSetWithdrawAddress) (out MsgSetWithdrawAddress) {
	var length int
	length = len(in.DelegatorAddress)
	if length == 0 {
		out.DelegatorAddress = nil
	} else {
		out.DelegatorAddress = make([]uint8, length)
	}
	copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
	length = len(in.WithdrawAddress)
	if length == 0 {
		out.WithdrawAddress = nil
	} else {
		out.WithdrawAddress = make([]uint8, length)
	}
	copy(out.WithdrawAddress[:], in.WithdrawAddress[:])
	return
} //End of DeepCopyMsgSetWithdrawAddress

// Non-Interface
func EncodeMsgUndelegate(w *[]byte, v MsgUndelegate) {
	codonEncodeByteSlice(0, w, v.DelegatorAddress[:])
	codonEncodeByteSlice(1, w, v.ValidatorAddress[:])
	codonEncodeByteSlice(2, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeString(0, w, v.Amount.Denom)
		codonEncodeByteSlice(1, w, EncodeInt(v.Amount.Amount))
		return wBuf
	}()) // end of v.Amount
} //End of EncodeMsgUndelegate

func DecodeMsgUndelegate(bz []byte) (v MsgUndelegate, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.DelegatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.DelegatorAddress = tmpBz
		case 1: // v.ValidatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorAddress = tmpBz
		case 2: // v.Amount
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Amount.Denom
						v.Amount.Denom = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 1: // v.Amount.Amount
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						v.Amount.Amount, n, err = DecodeInt(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgUndelegate

func RandMsgUndelegate(r RandSrc) MsgUndelegate {
	var length int
	var v MsgUndelegate
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.DelegatorAddress = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorAddress = r.GetBytes(length)
	v.Amount.Denom = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Amount.Amount = RandInt(r)
	// end of v.Amount
	return v
} //End of RandMsgUndelegate

func DeepCopyMsgUndelegate(in MsgUndelegate) (out MsgUndelegate) {
	var length int
	length = len(in.DelegatorAddress)
	if length == 0 {
		out.DelegatorAddress = nil
	} else {
		out.DelegatorAddress = make([]uint8, length)
	}
	copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
	length = len(in.ValidatorAddress)
	if length == 0 {
		out.ValidatorAddress = nil
	} else {
		out.ValidatorAddress = make([]uint8, length)
	}
	copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
	out.Amount.Denom = in.Amount.Denom
	out.Amount.Amount = DeepCopyInt(in.Amount.Amount)
	// end of .Amount
	return
} //End of DeepCopyMsgUndelegate

// Non-Interface
func EncodeMsgUnjail(w *[]byte, v MsgUnjail) {
	codonEncodeByteSlice(0, w, v.ValidatorAddr[:])
} //End of EncodeMsgUnjail

func DecodeMsgUnjail(bz []byte) (v MsgUnjail, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.ValidatorAddr
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorAddr = tmpBz
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgUnjail

func RandMsgUnjail(r RandSrc) MsgUnjail {
	var length int
	var v MsgUnjail
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorAddr = r.GetBytes(length)
	return v
} //End of RandMsgUnjail

func DeepCopyMsgUnjail(in MsgUnjail) (out MsgUnjail) {
	var length int
	length = len(in.ValidatorAddr)
	if length == 0 {
		out.ValidatorAddr = nil
	} else {
		out.ValidatorAddr = make([]uint8, length)
	}
	copy(out.ValidatorAddr[:], in.ValidatorAddr[:])
	return
} //End of DeepCopyMsgUnjail

// Non-Interface
func EncodeMsgWithdrawDelegatorReward(w *[]byte, v MsgWithdrawDelegatorReward) {
	codonEncodeByteSlice(0, w, v.DelegatorAddress[:])
	codonEncodeByteSlice(1, w, v.ValidatorAddress[:])
} //End of EncodeMsgWithdrawDelegatorReward

func DecodeMsgWithdrawDelegatorReward(bz []byte) (v MsgWithdrawDelegatorReward, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.DelegatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.DelegatorAddress = tmpBz
		case 1: // v.ValidatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorAddress = tmpBz
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgWithdrawDelegatorReward

func RandMsgWithdrawDelegatorReward(r RandSrc) MsgWithdrawDelegatorReward {
	var length int
	var v MsgWithdrawDelegatorReward
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.DelegatorAddress = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorAddress = r.GetBytes(length)
	return v
} //End of RandMsgWithdrawDelegatorReward

func DeepCopyMsgWithdrawDelegatorReward(in MsgWithdrawDelegatorReward) (out MsgWithdrawDelegatorReward) {
	var length int
	length = len(in.DelegatorAddress)
	if length == 0 {
		out.DelegatorAddress = nil
	} else {
		out.DelegatorAddress = make([]uint8, length)
	}
	copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
	length = len(in.ValidatorAddress)
	if length == 0 {
		out.ValidatorAddress = nil
	} else {
		out.ValidatorAddress = make([]uint8, length)
	}
	copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
	return
} //End of DeepCopyMsgWithdrawDelegatorReward

// Non-Interface
func EncodeMsgWithdrawValidatorCommission(w *[]byte, v MsgWithdrawValidatorCommission) {
	codonEncodeByteSlice(0, w, v.ValidatorAddress[:])
} //End of EncodeMsgWithdrawValidatorCommission

func DecodeMsgWithdrawValidatorCommission(bz []byte) (v MsgWithdrawValidatorCommission, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.ValidatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorAddress = tmpBz
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgWithdrawValidatorCommission

func RandMsgWithdrawValidatorCommission(r RandSrc) MsgWithdrawValidatorCommission {
	var length int
	var v MsgWithdrawValidatorCommission
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorAddress = r.GetBytes(length)
	return v
} //End of RandMsgWithdrawValidatorCommission

func DeepCopyMsgWithdrawValidatorCommission(in MsgWithdrawValidatorCommission) (out MsgWithdrawValidatorCommission) {
	var length int
	length = len(in.ValidatorAddress)
	if length == 0 {
		out.ValidatorAddress = nil
	} else {
		out.ValidatorAddress = make([]uint8, length)
	}
	copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
	return
} //End of DeepCopyMsgWithdrawValidatorCommission

// Non-Interface
func EncodeMsgDeposit(w *[]byte, v MsgDeposit) {
	codonEncodeUvarint(0, w, uint64(v.ProposalID))
	codonEncodeByteSlice(1, w, v.Depositor[:])
	for _0 := 0; _0 < len(v.Amount); _0++ {
		codonEncodeByteSlice(2, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.Amount[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeInt(v.Amount[_0].Amount))
			return wBuf
		}()) // end of v.Amount[_0]
	}
} //End of EncodeMsgDeposit

func DecodeMsgDeposit(bz []byte) (v MsgDeposit, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.ProposalID
			v.ProposalID = uint64(codonDecodeUint64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Depositor
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Depositor = tmpBz
		case 2: // v.Amount
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Coin
			tmp, n, err = DecodeCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Amount = append(v.Amount, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgDeposit

func RandMsgDeposit(r RandSrc) MsgDeposit {
	var length int
	var v MsgDeposit
	v.ProposalID = r.GetUint64()
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Depositor = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Amount = nil
	} else {
		v.Amount = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Amount[_0] = RandCoin(r)
	}
	return v
} //End of RandMsgDeposit

func DeepCopyMsgDeposit(in MsgDeposit) (out MsgDeposit) {
	var length int
	out.ProposalID = in.ProposalID
	length = len(in.Depositor)
	if length == 0 {
		out.Depositor = nil
	} else {
		out.Depositor = make([]uint8, length)
	}
	copy(out.Depositor[:], in.Depositor[:])
	length = len(in.Amount)
	if length == 0 {
		out.Amount = nil
	} else {
		out.Amount = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Amount[_0] = DeepCopyCoin(in.Amount[_0])
	}
	return
} //End of DeepCopyMsgDeposit

// Non-Interface
func EncodeMsgVote(w *[]byte, v MsgVote) {
	codonEncodeUvarint(0, w, uint64(v.ProposalID))
	codonEncodeByteSlice(1, w, v.Voter[:])
	codonEncodeUint8(2, w, uint8(v.Option))
} //End of EncodeMsgVote

func DecodeMsgVote(bz []byte) (v MsgVote, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.ProposalID
			v.ProposalID = uint64(codonDecodeUint64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Voter
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Voter = tmpBz
		case 2: // v.Option
			v.Option = VoteOption(codonDecodeUint8(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgVote

func RandMsgVote(r RandSrc) MsgVote {
	var length int
	var v MsgVote
	v.ProposalID = r.GetUint64()
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Voter = r.GetBytes(length)
	v.Option = VoteOption(r.GetUint8())
	return v
} //End of RandMsgVote

func DeepCopyMsgVote(in MsgVote) (out MsgVote) {
	var length int
	out.ProposalID = in.ProposalID
	length = len(in.Voter)
	if length == 0 {
		out.Voter = nil
	} else {
		out.Voter = make([]uint8, length)
	}
	copy(out.Voter[:], in.Voter[:])
	out.Option = in.Option
	return
} //End of DeepCopyMsgVote

// Non-Interface
func EncodeParameterChangeProposal(w *[]byte, v ParameterChangeProposal) {
	codonEncodeString(0, w, v.Title)
	codonEncodeString(1, w, v.Description)
	for _0 := 0; _0 < len(v.Changes); _0++ {
		codonEncodeByteSlice(2, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.Changes[_0].Subspace)
			codonEncodeString(1, w, v.Changes[_0].Key)
			codonEncodeString(2, w, v.Changes[_0].Subkey)
			codonEncodeString(3, w, v.Changes[_0].Value)
			return wBuf
		}()) // end of v.Changes[_0]
	}
} //End of EncodeParameterChangeProposal

func DecodeParameterChangeProposal(bz []byte) (v ParameterChangeProposal, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Title
			v.Title = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Description
			v.Description = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 2: // v.Changes
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp ParamChange
			tmp, n, err = DecodeParamChange(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Changes = append(v.Changes, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeParameterChangeProposal

func RandParameterChangeProposal(r RandSrc) ParameterChangeProposal {
	var length int
	var v ParameterChangeProposal
	v.Title = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Changes = nil
	} else {
		v.Changes = make([]ParamChange, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Changes[_0] = RandParamChange(r)
	}
	return v
} //End of RandParameterChangeProposal

func DeepCopyParameterChangeProposal(in ParameterChangeProposal) (out ParameterChangeProposal) {
	var length int
	out.Title = in.Title
	out.Description = in.Description
	length = len(in.Changes)
	if length == 0 {
		out.Changes = nil
	} else {
		out.Changes = make([]ParamChange, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Changes[_0] = DeepCopyParamChange(in.Changes[_0])
	}
	return
} //End of DeepCopyParameterChangeProposal

// Non-Interface
func EncodeSoftwareUpgradeProposal(w *[]byte, v SoftwareUpgradeProposal) {
	codonEncodeString(0, w, v.Title)
	codonEncodeString(1, w, v.Description)
} //End of EncodeSoftwareUpgradeProposal

func DecodeSoftwareUpgradeProposal(bz []byte) (v SoftwareUpgradeProposal, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Title
			v.Title = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Description
			v.Description = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeSoftwareUpgradeProposal

func RandSoftwareUpgradeProposal(r RandSrc) SoftwareUpgradeProposal {
	var v SoftwareUpgradeProposal
	v.Title = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	return v
} //End of RandSoftwareUpgradeProposal

func DeepCopySoftwareUpgradeProposal(in SoftwareUpgradeProposal) (out SoftwareUpgradeProposal) {
	out.Title = in.Title
	out.Description = in.Description
	return
} //End of DeepCopySoftwareUpgradeProposal

// Non-Interface
func EncodeTextProposal(w *[]byte, v TextProposal) {
	codonEncodeString(0, w, v.Title)
	codonEncodeString(1, w, v.Description)
} //End of EncodeTextProposal

func DecodeTextProposal(bz []byte) (v TextProposal, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Title
			v.Title = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Description
			v.Description = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeTextProposal

func RandTextProposal(r RandSrc) TextProposal {
	var v TextProposal
	v.Title = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	return v
} //End of RandTextProposal

func DeepCopyTextProposal(in TextProposal) (out TextProposal) {
	out.Title = in.Title
	out.Description = in.Description
	return
} //End of DeepCopyTextProposal

// Non-Interface
func EncodeCommunityPoolSpendProposal(w *[]byte, v CommunityPoolSpendProposal) {
	codonEncodeString(0, w, v.Title)
	codonEncodeString(1, w, v.Description)
	codonEncodeByteSlice(2, w, v.Recipient[:])
	for _0 := 0; _0 < len(v.Amount); _0++ {
		codonEncodeByteSlice(3, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.Amount[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeInt(v.Amount[_0].Amount))
			return wBuf
		}()) // end of v.Amount[_0]
	}
} //End of EncodeCommunityPoolSpendProposal

func DecodeCommunityPoolSpendProposal(bz []byte) (v CommunityPoolSpendProposal, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Title
			v.Title = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Description
			v.Description = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 2: // v.Recipient
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Recipient = tmpBz
		case 3: // v.Amount
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Coin
			tmp, n, err = DecodeCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Amount = append(v.Amount, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeCommunityPoolSpendProposal

func RandCommunityPoolSpendProposal(r RandSrc) CommunityPoolSpendProposal {
	var length int
	var v CommunityPoolSpendProposal
	v.Title = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Recipient = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Amount = nil
	} else {
		v.Amount = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Amount[_0] = RandCoin(r)
	}
	return v
} //End of RandCommunityPoolSpendProposal

func DeepCopyCommunityPoolSpendProposal(in CommunityPoolSpendProposal) (out CommunityPoolSpendProposal) {
	var length int
	out.Title = in.Title
	out.Description = in.Description
	length = len(in.Recipient)
	if length == 0 {
		out.Recipient = nil
	} else {
		out.Recipient = make([]uint8, length)
	}
	copy(out.Recipient[:], in.Recipient[:])
	length = len(in.Amount)
	if length == 0 {
		out.Amount = nil
	} else {
		out.Amount = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Amount[_0] = DeepCopyCoin(in.Amount[_0])
	}
	return
} //End of DeepCopyCommunityPoolSpendProposal

// Non-Interface
func EncodeMsgMultiSend(w *[]byte, v MsgMultiSend) {
	for _0 := 0; _0 < len(v.Inputs); _0++ {
		codonEncodeByteSlice(0, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeByteSlice(0, w, v.Inputs[_0].Address[:])
			for _1 := 0; _1 < len(v.Inputs[_0].Coins); _1++ {
				codonEncodeByteSlice(1, w, func() []byte {
					wBuf := make([]byte, 0, 64)
					w := &wBuf
					codonEncodeString(0, w, v.Inputs[_0].Coins[_1].Denom)
					codonEncodeByteSlice(1, w, EncodeInt(v.Inputs[_0].Coins[_1].Amount))
					return wBuf
				}()) // end of v.Inputs[_0].Coins[_1]
			}
			return wBuf
		}()) // end of v.Inputs[_0]
	}
	for _0 := 0; _0 < len(v.Outputs); _0++ {
		codonEncodeByteSlice(1, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeByteSlice(0, w, v.Outputs[_0].Address[:])
			for _1 := 0; _1 < len(v.Outputs[_0].Coins); _1++ {
				codonEncodeByteSlice(1, w, func() []byte {
					wBuf := make([]byte, 0, 64)
					w := &wBuf
					codonEncodeString(0, w, v.Outputs[_0].Coins[_1].Denom)
					codonEncodeByteSlice(1, w, EncodeInt(v.Outputs[_0].Coins[_1].Amount))
					return wBuf
				}()) // end of v.Outputs[_0].Coins[_1]
			}
			return wBuf
		}()) // end of v.Outputs[_0]
	}
} //End of EncodeMsgMultiSend

func DecodeMsgMultiSend(bz []byte) (v MsgMultiSend, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Inputs
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Input
			tmp, n, err = DecodeInput(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Inputs = append(v.Inputs, tmp)
		case 1: // v.Outputs
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Output
			tmp, n, err = DecodeOutput(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Outputs = append(v.Outputs, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgMultiSend

func RandMsgMultiSend(r RandSrc) MsgMultiSend {
	var length int
	var v MsgMultiSend
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Inputs = nil
	} else {
		v.Inputs = make([]Input, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Inputs[_0] = RandInput(r)
	}
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Outputs = nil
	} else {
		v.Outputs = make([]Output, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Outputs[_0] = RandOutput(r)
	}
	return v
} //End of RandMsgMultiSend

func DeepCopyMsgMultiSend(in MsgMultiSend) (out MsgMultiSend) {
	var length int
	length = len(in.Inputs)
	if length == 0 {
		out.Inputs = nil
	} else {
		out.Inputs = make([]Input, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Inputs[_0] = DeepCopyInput(in.Inputs[_0])
	}
	length = len(in.Outputs)
	if length == 0 {
		out.Outputs = nil
	} else {
		out.Outputs = make([]Output, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Outputs[_0] = DeepCopyOutput(in.Outputs[_0])
	}
	return
} //End of DeepCopyMsgMultiSend

// Non-Interface
func EncodeFeePool(w *[]byte, v FeePool) {
	for _0 := 0; _0 < len(v.CommunityPool); _0++ {
		codonEncodeByteSlice(0, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.CommunityPool[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeDec(v.CommunityPool[_0].Amount))
			return wBuf
		}()) // end of v.CommunityPool[_0]
	}
} //End of EncodeFeePool

func DecodeFeePool(bz []byte) (v FeePool, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.CommunityPool
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp DecCoin
			tmp, n, err = DecodeDecCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.CommunityPool = append(v.CommunityPool, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeFeePool

func RandFeePool(r RandSrc) FeePool {
	var length int
	var v FeePool
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.CommunityPool = nil
	} else {
		v.CommunityPool = make([]DecCoin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.CommunityPool[_0] = RandDecCoin(r)
	}
	return v
} //End of RandFeePool

func DeepCopyFeePool(in FeePool) (out FeePool) {
	var length int
	length = len(in.CommunityPool)
	if length == 0 {
		out.CommunityPool = nil
	} else {
		out.CommunityPool = make([]DecCoin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.CommunityPool[_0] = DeepCopyDecCoin(in.CommunityPool[_0])
	}
	return
} //End of DeepCopyFeePool

// Non-Interface
func EncodeMsgSend(w *[]byte, v MsgSend) {
	codonEncodeByteSlice(0, w, v.FromAddress[:])
	codonEncodeByteSlice(1, w, v.ToAddress[:])
	for _0 := 0; _0 < len(v.Amount); _0++ {
		codonEncodeByteSlice(2, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.Amount[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeInt(v.Amount[_0].Amount))
			return wBuf
		}()) // end of v.Amount[_0]
	}
} //End of EncodeMsgSend

func DecodeMsgSend(bz []byte) (v MsgSend, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.FromAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.FromAddress = tmpBz
		case 1: // v.ToAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ToAddress = tmpBz
		case 2: // v.Amount
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Coin
			tmp, n, err = DecodeCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Amount = append(v.Amount, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgSend

func RandMsgSend(r RandSrc) MsgSend {
	var length int
	var v MsgSend
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.FromAddress = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ToAddress = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Amount = nil
	} else {
		v.Amount = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Amount[_0] = RandCoin(r)
	}
	return v
} //End of RandMsgSend

func DeepCopyMsgSend(in MsgSend) (out MsgSend) {
	var length int
	length = len(in.FromAddress)
	if length == 0 {
		out.FromAddress = nil
	} else {
		out.FromAddress = make([]uint8, length)
	}
	copy(out.FromAddress[:], in.FromAddress[:])
	length = len(in.ToAddress)
	if length == 0 {
		out.ToAddress = nil
	} else {
		out.ToAddress = make([]uint8, length)
	}
	copy(out.ToAddress[:], in.ToAddress[:])
	length = len(in.Amount)
	if length == 0 {
		out.Amount = nil
	} else {
		out.Amount = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Amount[_0] = DeepCopyCoin(in.Amount[_0])
	}
	return
} //End of DeepCopyMsgSend

// Non-Interface
func EncodeMsgVerifyInvariant(w *[]byte, v MsgVerifyInvariant) {
	codonEncodeByteSlice(0, w, v.Sender[:])
	codonEncodeString(1, w, v.InvariantModuleName)
	codonEncodeString(2, w, v.InvariantRoute)
} //End of EncodeMsgVerifyInvariant

func DecodeMsgVerifyInvariant(bz []byte) (v MsgVerifyInvariant, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Sender
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Sender = tmpBz
		case 1: // v.InvariantModuleName
			v.InvariantModuleName = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 2: // v.InvariantRoute
			v.InvariantRoute = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeMsgVerifyInvariant

func RandMsgVerifyInvariant(r RandSrc) MsgVerifyInvariant {
	var length int
	var v MsgVerifyInvariant
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Sender = r.GetBytes(length)
	v.InvariantModuleName = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.InvariantRoute = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	return v
} //End of RandMsgVerifyInvariant

func DeepCopyMsgVerifyInvariant(in MsgVerifyInvariant) (out MsgVerifyInvariant) {
	var length int
	length = len(in.Sender)
	if length == 0 {
		out.Sender = nil
	} else {
		out.Sender = make([]uint8, length)
	}
	copy(out.Sender[:], in.Sender[:])
	out.InvariantModuleName = in.InvariantModuleName
	out.InvariantRoute = in.InvariantRoute
	return
} //End of DeepCopyMsgVerifyInvariant

// Non-Interface
func EncodeSupply(w *[]byte, v Supply) {
	for _0 := 0; _0 < len(v.Total); _0++ {
		codonEncodeByteSlice(0, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.Total[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeInt(v.Total[_0].Amount))
			return wBuf
		}()) // end of v.Total[_0]
	}
} //End of EncodeSupply

func DecodeSupply(bz []byte) (v Supply, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Total
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp Coin
			tmp, n, err = DecodeCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Total = append(v.Total, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeSupply

func RandSupply(r RandSrc) Supply {
	var length int
	var v Supply
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Total = nil
	} else {
		v.Total = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Total[_0] = RandCoin(r)
	}
	return v
} //End of RandSupply

func DeepCopySupply(in Supply) (out Supply) {
	var length int
	length = len(in.Total)
	if length == 0 {
		out.Total = nil
	} else {
		out.Total = make([]Coin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Total[_0] = DeepCopyCoin(in.Total[_0])
	}
	return
} //End of DeepCopySupply

// Non-Interface
func EncodeAccAddressList(w *[]byte, v AccAddressList) {
	for _0 := 0; _0 < len(v); _0++ {
		codonEncodeByteSlice(0, w, v[_0][:])
	}
} //End of EncodeAccAddressList

func DecodeAccAddressList(bz []byte) (v AccAddressList, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			var tmp AccAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			tmp = tmpBz
			v = append(v, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeAccAddressList

func RandAccAddressList(r RandSrc) AccAddressList {
	var length int
	var v AccAddressList
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v = nil
	} else {
		v = make([]AccAddress, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of slice
		length = 1 + int(r.GetUint()%(MaxSliceLength-1))
		v[_0] = r.GetBytes(length)
	}
	return v
} //End of RandAccAddressList

func DeepCopyAccAddressList(in AccAddressList) (out AccAddressList) {
	var length int
	length = len(in)
	if length == 0 {
		out = nil
	} else {
		out = make([]AccAddress, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of slice
		length = len(in[_0])
		if length == 0 {
			out[_0] = nil
		} else {
			out[_0] = make([]uint8, length)
		}
		copy(out[_0][:], in[_0][:])
	}
	return
} //End of DeepCopyAccAddressList

// Non-Interface
func EncodeCommitInfo(w *[]byte, v CommitInfo) {
	codonEncodeVarint(0, w, int64(v.Version))
	for _0 := 0; _0 < len(v.StoreInfos); _0++ {
		codonEncodeByteSlice(1, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.StoreInfos[_0].Name)
			codonEncodeByteSlice(1, w, func() []byte {
				wBuf := make([]byte, 0, 64)
				w := &wBuf
				codonEncodeByteSlice(0, w, func() []byte {
					wBuf := make([]byte, 0, 64)
					w := &wBuf
					codonEncodeVarint(0, w, int64(v.StoreInfos[_0].Core.CommitID.Version))
					codonEncodeByteSlice(1, w, v.StoreInfos[_0].Core.CommitID.Hash[:])
					return wBuf
				}()) // end of v.StoreInfos[_0].Core.CommitID
				return wBuf
			}()) // end of v.StoreInfos[_0].Core
			return wBuf
		}()) // end of v.StoreInfos[_0]
	}
} //End of EncodeCommitInfo

func DecodeCommitInfo(bz []byte) (v CommitInfo, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Version
			v.Version = int64(codonDecodeInt64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.StoreInfos
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp StoreInfo
			tmp, n, err = DecodeStoreInfo(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.StoreInfos = append(v.StoreInfos, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeCommitInfo

func RandCommitInfo(r RandSrc) CommitInfo {
	var length int
	var v CommitInfo
	v.Version = r.GetInt64()
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.StoreInfos = nil
	} else {
		v.StoreInfos = make([]StoreInfo, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.StoreInfos[_0] = RandStoreInfo(r)
	}
	return v
} //End of RandCommitInfo

func DeepCopyCommitInfo(in CommitInfo) (out CommitInfo) {
	var length int
	out.Version = in.Version
	length = len(in.StoreInfos)
	if length == 0 {
		out.StoreInfos = nil
	} else {
		out.StoreInfos = make([]StoreInfo, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.StoreInfos[_0] = DeepCopyStoreInfo(in.StoreInfos[_0])
	}
	return
} //End of DeepCopyCommitInfo

// Non-Interface
func EncodeStoreInfo(w *[]byte, v StoreInfo) {
	codonEncodeString(0, w, v.Name)
	codonEncodeByteSlice(1, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeByteSlice(0, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeVarint(0, w, int64(v.Core.CommitID.Version))
			codonEncodeByteSlice(1, w, v.Core.CommitID.Hash[:])
			return wBuf
		}()) // end of v.Core.CommitID
		return wBuf
	}()) // end of v.Core
} //End of EncodeStoreInfo

func DecodeStoreInfo(bz []byte) (v StoreInfo, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Name
			v.Name = string(codonDecodeString(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Core
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Core.CommitID
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						func(bz []byte) {
							for len(bz) != 0 {
								tag := codonDecodeUint64(bz, &n, &err)
								if err != nil {
									return
								}
								bz = bz[n:]
								total += n
								tag = tag >> 3
								switch tag {
								case 0: // v.Core.CommitID.Version
									v.Core.CommitID.Version = int64(codonDecodeInt64(bz, &n, &err))
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
								case 1: // v.Core.CommitID.Hash
									var tmpBz []byte
									n, err = codonGetByteSlice(&tmpBz, bz)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									v.Core.CommitID.Hash = tmpBz
								default:
									err = errors.New("Unknown Field")
									return
								}
							} // end for
						}(bz[:l]) // end func
						if err != nil {
							return
						}
						bz = bz[l:]
						n += int(l)
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeStoreInfo

func RandStoreInfo(r RandSrc) StoreInfo {
	var length int
	var v StoreInfo
	v.Name = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Core.CommitID.Version = r.GetInt64()
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Core.CommitID.Hash = r.GetBytes(length)
	// end of v.Core.CommitID
	// end of v.Core
	return v
} //End of RandStoreInfo

func DeepCopyStoreInfo(in StoreInfo) (out StoreInfo) {
	var length int
	out.Name = in.Name
	out.Core.CommitID.Version = in.Core.CommitID.Version
	length = len(in.Core.CommitID.Hash)
	if length == 0 {
		out.Core.CommitID.Hash = nil
	} else {
		out.Core.CommitID.Hash = make([]uint8, length)
	}
	copy(out.Core.CommitID.Hash[:], in.Core.CommitID.Hash[:])
	// end of .Core.CommitID
	// end of .Core
	return
} //End of DeepCopyStoreInfo

// Non-Interface
func EncodeValidator(w *[]byte, v Validator) {
	codonEncodeByteSlice(0, w, v.OperatorAddress[:])
	codonEncodeByteSlice(1, w, func() []byte {
		w := make([]byte, 0, 64)
		EncodePubKey(&w, v.ConsPubKey) // interface_encode
		return w
	}()) // end of v.ConsPubKey
	codonEncodeBool(2, w, v.Jailed)
	codonEncodeUint8(3, w, uint8(v.Status))
	codonEncodeByteSlice(4, w, EncodeInt(v.Tokens))
	codonEncodeByteSlice(5, w, EncodeDec(v.DelegatorShares))
	codonEncodeByteSlice(6, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeString(0, w, v.Description.Moniker)
		codonEncodeString(1, w, v.Description.Identity)
		codonEncodeString(2, w, v.Description.Website)
		codonEncodeString(3, w, v.Description.Details)
		return wBuf
	}()) // end of v.Description
	codonEncodeVarint(7, w, int64(v.UnbondingHeight))
	codonEncodeByteSlice(8, w, EncodeTime(v.UnbondingCompletionTime))
	codonEncodeByteSlice(9, w, func() []byte {
		wBuf := make([]byte, 0, 64)
		w := &wBuf
		codonEncodeByteSlice(0, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeByteSlice(0, w, EncodeDec(v.Commission.CommissionRates.Rate))
			codonEncodeByteSlice(1, w, EncodeDec(v.Commission.CommissionRates.MaxRate))
			codonEncodeByteSlice(2, w, EncodeDec(v.Commission.CommissionRates.MaxChangeRate))
			return wBuf
		}()) // end of v.Commission.CommissionRates
		codonEncodeByteSlice(1, w, EncodeTime(v.Commission.UpdateTime))
		return wBuf
	}()) // end of v.Commission
	codonEncodeByteSlice(10, w, EncodeInt(v.MinSelfDelegation))
} //End of EncodeValidator

func DecodeValidator(bz []byte) (v Validator, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.OperatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.OperatorAddress = tmpBz
		case 1: // v.ConsPubKey
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.ConsPubKey, n, err = DecodePubKey(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n // interface_decode
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 2: // v.Jailed
			v.Jailed = bool(codonDecodeBool(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 3: // v.Status
			v.Status = BondStatus(codonDecodeUint8(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 4: // v.Tokens
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.Tokens, n, err = DecodeInt(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 5: // v.DelegatorShares
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.DelegatorShares, n, err = DecodeDec(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 6: // v.Description
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Description.Moniker
						v.Description.Moniker = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 1: // v.Description.Identity
						v.Description.Identity = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 2: // v.Description.Website
						v.Description.Website = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					case 3: // v.Description.Details
						v.Description.Details = string(codonDecodeString(bz, &n, &err))
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		case 7: // v.UnbondingHeight
			v.UnbondingHeight = int64(codonDecodeInt64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 8: // v.UnbondingCompletionTime
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.UnbondingCompletionTime, n, err = DecodeTime(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 9: // v.Commission
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			func(bz []byte) {
				for len(bz) != 0 {
					tag := codonDecodeUint64(bz, &n, &err)
					if err != nil {
						return
					}
					bz = bz[n:]
					total += n
					tag = tag >> 3
					switch tag {
					case 0: // v.Commission.CommissionRates
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						func(bz []byte) {
							for len(bz) != 0 {
								tag := codonDecodeUint64(bz, &n, &err)
								if err != nil {
									return
								}
								bz = bz[n:]
								total += n
								tag = tag >> 3
								switch tag {
								case 0: // v.Commission.CommissionRates.Rate
									l := codonDecodeUint64(bz, &n, &err)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) > len(bz) {
										err = errors.New("Length Too Large")
										return
									}
									v.Commission.CommissionRates.Rate, n, err = DecodeDec(bz[:l])
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) != n {
										err = errors.New("Length Mismatch")
										return
									}
								case 1: // v.Commission.CommissionRates.MaxRate
									l := codonDecodeUint64(bz, &n, &err)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) > len(bz) {
										err = errors.New("Length Too Large")
										return
									}
									v.Commission.CommissionRates.MaxRate, n, err = DecodeDec(bz[:l])
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) != n {
										err = errors.New("Length Mismatch")
										return
									}
								case 2: // v.Commission.CommissionRates.MaxChangeRate
									l := codonDecodeUint64(bz, &n, &err)
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) > len(bz) {
										err = errors.New("Length Too Large")
										return
									}
									v.Commission.CommissionRates.MaxChangeRate, n, err = DecodeDec(bz[:l])
									if err != nil {
										return
									}
									bz = bz[n:]
									total += n
									if int(l) != n {
										err = errors.New("Length Mismatch")
										return
									}
								default:
									err = errors.New("Unknown Field")
									return
								}
							} // end for
						}(bz[:l]) // end func
						if err != nil {
							return
						}
						bz = bz[l:]
						n += int(l)
					case 1: // v.Commission.UpdateTime
						l := codonDecodeUint64(bz, &n, &err)
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) > len(bz) {
							err = errors.New("Length Too Large")
							return
						}
						v.Commission.UpdateTime, n, err = DecodeTime(bz[:l])
						if err != nil {
							return
						}
						bz = bz[n:]
						total += n
						if int(l) != n {
							err = errors.New("Length Mismatch")
							return
						}
					default:
						err = errors.New("Unknown Field")
						return
					}
				} // end for
			}(bz[:l]) // end func
			if err != nil {
				return
			}
			bz = bz[l:]
			n += int(l)
		case 10: // v.MinSelfDelegation
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.MinSelfDelegation, n, err = DecodeInt(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeValidator

func RandValidator(r RandSrc) Validator {
	var length int
	var v Validator
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.OperatorAddress = r.GetBytes(length)
	v.ConsPubKey = RandPubKey(r) // interface_decode
	v.Jailed = r.GetBool()
	v.Status = BondStatus(r.GetUint8())
	v.Tokens = RandInt(r)
	v.DelegatorShares = RandDec(r)
	v.Description.Moniker = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description.Identity = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description.Website = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	v.Description.Details = r.GetString(1 + int(r.GetUint()%(MaxStringLength-1)))
	// end of v.Description
	v.UnbondingHeight = r.GetInt64()
	v.UnbondingCompletionTime = RandTime(r)
	v.Commission.CommissionRates.Rate = RandDec(r)
	v.Commission.CommissionRates.MaxRate = RandDec(r)
	v.Commission.CommissionRates.MaxChangeRate = RandDec(r)
	// end of v.Commission.CommissionRates
	v.Commission.UpdateTime = RandTime(r)
	// end of v.Commission
	v.MinSelfDelegation = RandInt(r)
	return v
} //End of RandValidator

func DeepCopyValidator(in Validator) (out Validator) {
	var length int
	length = len(in.OperatorAddress)
	if length == 0 {
		out.OperatorAddress = nil
	} else {
		out.OperatorAddress = make([]uint8, length)
	}
	copy(out.OperatorAddress[:], in.OperatorAddress[:])
	out.ConsPubKey = DeepCopyPubKey(in.ConsPubKey)
	out.Jailed = in.Jailed
	out.Status = in.Status
	out.Tokens = DeepCopyInt(in.Tokens)
	out.DelegatorShares = DeepCopyDec(in.DelegatorShares)
	out.Description.Moniker = in.Description.Moniker
	out.Description.Identity = in.Description.Identity
	out.Description.Website = in.Description.Website
	out.Description.Details = in.Description.Details
	// end of .Description
	out.UnbondingHeight = in.UnbondingHeight
	out.UnbondingCompletionTime = DeepCopyTime(in.UnbondingCompletionTime)
	out.Commission.CommissionRates.Rate = DeepCopyDec(in.Commission.CommissionRates.Rate)
	out.Commission.CommissionRates.MaxRate = DeepCopyDec(in.Commission.CommissionRates.MaxRate)
	out.Commission.CommissionRates.MaxChangeRate = DeepCopyDec(in.Commission.CommissionRates.MaxChangeRate)
	// end of .Commission.CommissionRates
	out.Commission.UpdateTime = DeepCopyTime(in.Commission.UpdateTime)
	// end of .Commission
	out.MinSelfDelegation = DeepCopyInt(in.MinSelfDelegation)
	return
} //End of DeepCopyValidator

// Non-Interface
func EncodeDelegation(w *[]byte, v Delegation) {
	codonEncodeByteSlice(0, w, v.DelegatorAddress[:])
	codonEncodeByteSlice(1, w, v.ValidatorAddress[:])
	codonEncodeByteSlice(2, w, EncodeDec(v.Shares))
} //End of EncodeDelegation

func DecodeDelegation(bz []byte) (v Delegation, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.DelegatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.DelegatorAddress = tmpBz
		case 1: // v.ValidatorAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.ValidatorAddress = tmpBz
		case 2: // v.Shares
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.Shares, n, err = DecodeDec(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeDelegation

func RandDelegation(r RandSrc) Delegation {
	var length int
	var v Delegation
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.DelegatorAddress = r.GetBytes(length)
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.ValidatorAddress = r.GetBytes(length)
	v.Shares = RandDec(r)
	return v
} //End of RandDelegation

func DeepCopyDelegation(in Delegation) (out Delegation) {
	var length int
	length = len(in.DelegatorAddress)
	if length == 0 {
		out.DelegatorAddress = nil
	} else {
		out.DelegatorAddress = make([]uint8, length)
	}
	copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
	length = len(in.ValidatorAddress)
	if length == 0 {
		out.ValidatorAddress = nil
	} else {
		out.ValidatorAddress = make([]uint8, length)
	}
	copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
	out.Shares = DeepCopyDec(in.Shares)
	return
} //End of DeepCopyDelegation

// Non-Interface
func EncodeBondStatus(w *[]byte, v BondStatus) {
	codonEncodeUint8(0, w, uint8(v))
} //End of EncodeBondStatus

func DecodeBondStatus(bz []byte) (v BondStatus, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			v = BondStatus(codonDecodeUint8(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeBondStatus

func RandBondStatus(r RandSrc) BondStatus {
	var v BondStatus
	v = BondStatus(r.GetUint8())
	return v
} //End of RandBondStatus

func DeepCopyBondStatus(in BondStatus) (out BondStatus) {
	out = in
	return
} //End of DeepCopyBondStatus

// Non-Interface
func EncodeDelegatorStartingInfo(w *[]byte, v DelegatorStartingInfo) {
	codonEncodeUvarint(0, w, uint64(v.PreviousPeriod))
	codonEncodeByteSlice(1, w, EncodeDec(v.Stake))
	codonEncodeUvarint(2, w, uint64(v.Height))
} //End of EncodeDelegatorStartingInfo

func DecodeDelegatorStartingInfo(bz []byte) (v DelegatorStartingInfo, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.PreviousPeriod
			v.PreviousPeriod = uint64(codonDecodeUint64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Stake
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.Stake, n, err = DecodeDec(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 2: // v.Height
			v.Height = uint64(codonDecodeUint64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeDelegatorStartingInfo

func RandDelegatorStartingInfo(r RandSrc) DelegatorStartingInfo {
	var v DelegatorStartingInfo
	v.PreviousPeriod = r.GetUint64()
	v.Stake = RandDec(r)
	v.Height = r.GetUint64()
	return v
} //End of RandDelegatorStartingInfo

func DeepCopyDelegatorStartingInfo(in DelegatorStartingInfo) (out DelegatorStartingInfo) {
	out.PreviousPeriod = in.PreviousPeriod
	out.Stake = DeepCopyDec(in.Stake)
	out.Height = in.Height
	return
} //End of DeepCopyDelegatorStartingInfo

// Non-Interface
func EncodeValidatorHistoricalRewards(w *[]byte, v ValidatorHistoricalRewards) {
	for _0 := 0; _0 < len(v.CumulativeRewardRatio); _0++ {
		codonEncodeByteSlice(0, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.CumulativeRewardRatio[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeDec(v.CumulativeRewardRatio[_0].Amount))
			return wBuf
		}()) // end of v.CumulativeRewardRatio[_0]
	}
	codonEncodeUint16(1, w, v.ReferenceCount)
} //End of EncodeValidatorHistoricalRewards

func DecodeValidatorHistoricalRewards(bz []byte) (v ValidatorHistoricalRewards, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.CumulativeRewardRatio
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp DecCoin
			tmp, n, err = DecodeDecCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.CumulativeRewardRatio = append(v.CumulativeRewardRatio, tmp)
		case 1: // v.ReferenceCount
			v.ReferenceCount = uint16(codonDecodeUint16(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeValidatorHistoricalRewards

func RandValidatorHistoricalRewards(r RandSrc) ValidatorHistoricalRewards {
	var length int
	var v ValidatorHistoricalRewards
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.CumulativeRewardRatio = nil
	} else {
		v.CumulativeRewardRatio = make([]DecCoin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.CumulativeRewardRatio[_0] = RandDecCoin(r)
	}
	v.ReferenceCount = r.GetUint16()
	return v
} //End of RandValidatorHistoricalRewards

func DeepCopyValidatorHistoricalRewards(in ValidatorHistoricalRewards) (out ValidatorHistoricalRewards) {
	var length int
	length = len(in.CumulativeRewardRatio)
	if length == 0 {
		out.CumulativeRewardRatio = nil
	} else {
		out.CumulativeRewardRatio = make([]DecCoin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.CumulativeRewardRatio[_0] = DeepCopyDecCoin(in.CumulativeRewardRatio[_0])
	}
	out.ReferenceCount = in.ReferenceCount
	return
} //End of DeepCopyValidatorHistoricalRewards

// Non-Interface
func EncodeValidatorCurrentRewards(w *[]byte, v ValidatorCurrentRewards) {
	for _0 := 0; _0 < len(v.Rewards); _0++ {
		codonEncodeByteSlice(0, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v.Rewards[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeDec(v.Rewards[_0].Amount))
			return wBuf
		}()) // end of v.Rewards[_0]
	}
	codonEncodeUvarint(1, w, uint64(v.Period))
} //End of EncodeValidatorCurrentRewards

func DecodeValidatorCurrentRewards(bz []byte) (v ValidatorCurrentRewards, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Rewards
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp DecCoin
			tmp, n, err = DecodeDecCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v.Rewards = append(v.Rewards, tmp)
		case 1: // v.Period
			v.Period = uint64(codonDecodeUint64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeValidatorCurrentRewards

func RandValidatorCurrentRewards(r RandSrc) ValidatorCurrentRewards {
	var length int
	var v ValidatorCurrentRewards
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v.Rewards = nil
	} else {
		v.Rewards = make([]DecCoin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v.Rewards[_0] = RandDecCoin(r)
	}
	v.Period = r.GetUint64()
	return v
} //End of RandValidatorCurrentRewards

func DeepCopyValidatorCurrentRewards(in ValidatorCurrentRewards) (out ValidatorCurrentRewards) {
	var length int
	length = len(in.Rewards)
	if length == 0 {
		out.Rewards = nil
	} else {
		out.Rewards = make([]DecCoin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out.Rewards[_0] = DeepCopyDecCoin(in.Rewards[_0])
	}
	out.Period = in.Period
	return
} //End of DeepCopyValidatorCurrentRewards

// Non-Interface
func EncodeValidatorSigningInfo(w *[]byte, v ValidatorSigningInfo) {
	codonEncodeByteSlice(0, w, v.Address[:])
	codonEncodeVarint(1, w, int64(v.StartHeight))
	codonEncodeVarint(2, w, int64(v.IndexOffset))
	codonEncodeByteSlice(3, w, EncodeTime(v.JailedUntil))
	codonEncodeBool(4, w, v.Tombstoned)
	codonEncodeVarint(5, w, int64(v.MissedBlocksCounter))
} //End of EncodeValidatorSigningInfo

func DecodeValidatorSigningInfo(bz []byte) (v ValidatorSigningInfo, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.Address
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v.Address = tmpBz
		case 1: // v.StartHeight
			v.StartHeight = int64(codonDecodeInt64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 2: // v.IndexOffset
			v.IndexOffset = int64(codonDecodeInt64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 3: // v.JailedUntil
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.JailedUntil, n, err = DecodeTime(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		case 4: // v.Tombstoned
			v.Tombstoned = bool(codonDecodeBool(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 5: // v.MissedBlocksCounter
			v.MissedBlocksCounter = int64(codonDecodeInt64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeValidatorSigningInfo

func RandValidatorSigningInfo(r RandSrc) ValidatorSigningInfo {
	var length int
	var v ValidatorSigningInfo
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v.Address = r.GetBytes(length)
	v.StartHeight = r.GetInt64()
	v.IndexOffset = r.GetInt64()
	v.JailedUntil = RandTime(r)
	v.Tombstoned = r.GetBool()
	v.MissedBlocksCounter = r.GetInt64()
	return v
} //End of RandValidatorSigningInfo

func DeepCopyValidatorSigningInfo(in ValidatorSigningInfo) (out ValidatorSigningInfo) {
	var length int
	length = len(in.Address)
	if length == 0 {
		out.Address = nil
	} else {
		out.Address = make([]uint8, length)
	}
	copy(out.Address[:], in.Address[:])
	out.StartHeight = in.StartHeight
	out.IndexOffset = in.IndexOffset
	out.JailedUntil = DeepCopyTime(in.JailedUntil)
	out.Tombstoned = in.Tombstoned
	out.MissedBlocksCounter = in.MissedBlocksCounter
	return
} //End of DeepCopyValidatorSigningInfo

// Non-Interface
func EncodeValidatorSlashEvent(w *[]byte, v ValidatorSlashEvent) {
	codonEncodeUvarint(0, w, uint64(v.ValidatorPeriod))
	codonEncodeByteSlice(1, w, EncodeDec(v.Fraction))
} //End of EncodeValidatorSlashEvent

func DecodeValidatorSlashEvent(bz []byte) (v ValidatorSlashEvent, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0: // v.ValidatorPeriod
			v.ValidatorPeriod = uint64(codonDecodeUint64(bz, &n, &err))
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
		case 1: // v.Fraction
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			v.Fraction, n, err = DecodeDec(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeValidatorSlashEvent

func RandValidatorSlashEvent(r RandSrc) ValidatorSlashEvent {
	var v ValidatorSlashEvent
	v.ValidatorPeriod = r.GetUint64()
	v.Fraction = RandDec(r)
	return v
} //End of RandValidatorSlashEvent

func DeepCopyValidatorSlashEvent(in ValidatorSlashEvent) (out ValidatorSlashEvent) {
	out.ValidatorPeriod = in.ValidatorPeriod
	out.Fraction = DeepCopyDec(in.Fraction)
	return
} //End of DeepCopyValidatorSlashEvent

// Non-Interface
func EncodeDecCoins(w *[]byte, v DecCoins) {
	for _0 := 0; _0 < len(v); _0++ {
		codonEncodeByteSlice(0, w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			codonEncodeString(0, w, v[_0].Denom)
			codonEncodeByteSlice(1, w, EncodeDec(v[_0].Amount))
			return wBuf
		}()) // end of v[_0]
	}
} //End of EncodeDecCoins

func DecodeDecCoins(bz []byte) (v DecCoins, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			l := codonDecodeUint64(bz, &n, &err)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) > len(bz) {
				err = errors.New("Length Too Large")
				return
			}
			var tmp DecCoin
			tmp, n, err = DecodeDecCoin(bz[:l])
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			if int(l) != n {
				err = errors.New("Length Mismatch")
				return
			}
			v = append(v, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeDecCoins

func RandDecCoins(r RandSrc) DecCoins {
	var length int
	var v DecCoins
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v = nil
	} else {
		v = make([]DecCoin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		v[_0] = RandDecCoin(r)
	}
	return v
} //End of RandDecCoins

func DeepCopyDecCoins(in DecCoins) (out DecCoins) {
	var length int
	length = len(in)
	if length == 0 {
		out = nil
	} else {
		out = make([]DecCoin, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of struct
		out[_0] = DeepCopyDecCoin(in[_0])
	}
	return
} //End of DeepCopyDecCoins

// Non-Interface
func EncodeValAddress(w *[]byte, v ValAddress) {
	codonEncodeByteSlice(0, w, v[:])
} //End of EncodeValAddress

func DecodeValAddress(bz []byte) (v ValAddress, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			v = tmpBz
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeValAddress

func RandValAddress(r RandSrc) ValAddress {
	var length int
	var v ValAddress
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	v = r.GetBytes(length)
	return v
} //End of RandValAddress

func DeepCopyValAddress(in ValAddress) (out ValAddress) {
	var length int
	length = len(in)
	if length == 0 {
		out = nil
	} else {
		out = make([]uint8, length)
	}
	copy(out[:], in[:])
	return
} //End of DeepCopyValAddress

// Non-Interface
func EncodeValAddressList(w *[]byte, v ValAddressList) {
	for _0 := 0; _0 < len(v); _0++ {
		codonEncodeByteSlice(0, w, v[_0][:])
	}
} //End of EncodeValAddressList

func DecodeValAddressList(bz []byte) (v ValAddressList, total int, err error) {
	var n int
	for len(bz) != 0 {
		tag := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return v, total, err
		}
		bz = bz[n:]
		total += n
		tag = tag >> 3
		switch tag {
		case 0:
			var tmp ValAddress
			var tmpBz []byte
			n, err = codonGetByteSlice(&tmpBz, bz)
			if err != nil {
				return
			}
			bz = bz[n:]
			total += n
			tmp = tmpBz
			v = append(v, tmp)
		default:
			err = errors.New("Unknown Field")
			return
		}
	} // end for
	return v, total, nil
} //End of DecodeValAddressList

func RandValAddressList(r RandSrc) ValAddressList {
	var length int
	var v ValAddressList
	length = 1 + int(r.GetUint()%(MaxSliceLength-1))
	if length == 0 {
		v = nil
	} else {
		v = make([]ValAddress, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of slice
		length = 1 + int(r.GetUint()%(MaxSliceLength-1))
		v[_0] = r.GetBytes(length)
	}
	return v
} //End of RandValAddressList

func DeepCopyValAddressList(in ValAddressList) (out ValAddressList) {
	var length int
	length = len(in)
	if length == 0 {
		out = nil
	} else {
		out = make([]ValAddress, length)
	}
	for _0, length_0 := 0, length; _0 < length_0; _0++ { //slice of slice
		length = len(in[_0])
		if length == 0 {
			out[_0] = nil
		} else {
			out[_0] = make([]uint8, length)
		}
		copy(out[_0][:], in[_0][:])
	}
	return
} //End of DeepCopyValAddressList

// Interface
func DecodePubKey(bz []byte) (v PubKey, total int, err error) {

	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return v, total, err
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)
	switch magicNum {
	case 0xa5b24c72:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp PubKeyEd25519
		tmp, n, err = DecodePubKeyEd25519(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x7f05210e:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp PubKeyMultisigThreshold
		tmp, n, err = DecodePubKeyMultisigThreshold(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x60f9a133:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp PubKeySecp256k1
		tmp, n, err = DecodePubKeySecp256k1(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x461d2af7:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp StdSignature
		tmp, n, err = DecodeStdSignature(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	default:
		panic("Unknown type")
	} // end of switch
	return v, n, nil
} // end of DecodePubKey
func EncodePubKey(w *[]byte, x interface{}) {
	switch v := x.(type) {
	case PubKeyEd25519:
		codonEncodeByteSlice(int(getMagicNum("PubKeyEd25519")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeyEd25519(w, v)
			return wBuf
		}())
	case *PubKeyEd25519:
		codonEncodeByteSlice(int(getMagicNum("PubKeyEd25519")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeyEd25519(w, *v)
			return wBuf
		}())
	case PubKeyMultisigThreshold:
		codonEncodeByteSlice(int(getMagicNum("PubKeyMultisigThreshold")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeyMultisigThreshold(w, v)
			return wBuf
		}())
	case *PubKeyMultisigThreshold:
		codonEncodeByteSlice(int(getMagicNum("PubKeyMultisigThreshold")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeyMultisigThreshold(w, *v)
			return wBuf
		}())
	case PubKeySecp256k1:
		codonEncodeByteSlice(int(getMagicNum("PubKeySecp256k1")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeySecp256k1(w, v)
			return wBuf
		}())
	case *PubKeySecp256k1:
		codonEncodeByteSlice(int(getMagicNum("PubKeySecp256k1")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeySecp256k1(w, *v)
			return wBuf
		}())
	case StdSignature:
		codonEncodeByteSlice(int(getMagicNum("StdSignature")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeStdSignature(w, v)
			return wBuf
		}())
	case *StdSignature:
		codonEncodeByteSlice(int(getMagicNum("StdSignature")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeStdSignature(w, *v)
			return wBuf
		}())
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func RandPubKey(r RandSrc) PubKey {
	switch r.GetUint() % 2 {
	case 0:
		return RandPubKeyEd25519(r)
	case 1:
		return RandPubKeySecp256k1(r)
	default:
		panic("Unknown Type.")
	} // end of switch
} // end of func
func DeepCopyPubKey(x PubKey) PubKey {
	switch v := x.(type) {
	case PubKeyEd25519:
		res := DeepCopyPubKeyEd25519(v)
		return res
	case *PubKeyEd25519:
		res := DeepCopyPubKeyEd25519(*v)
		return &res
	case PubKeySecp256k1:
		res := DeepCopyPubKeySecp256k1(v)
		return res
	case *PubKeySecp256k1:
		res := DeepCopyPubKeySecp256k1(*v)
		return &res
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
// Interface
func DecodePrivKey(bz []byte) (v PrivKey, total int, err error) {

	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return v, total, err
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)
	switch magicNum {
	case 0x56ba5e9e:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp PrivKeyEd25519
		tmp, n, err = DecodePrivKeyEd25519(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x08171053:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp PrivKeySecp256k1
		tmp, n, err = DecodePrivKeySecp256k1(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	default:
		panic("Unknown type")
	} // end of switch
	return v, n, nil
} // end of DecodePrivKey
func EncodePrivKey(w *[]byte, x interface{}) {
	switch v := x.(type) {
	case PrivKeyEd25519:
		codonEncodeByteSlice(int(getMagicNum("PrivKeyEd25519")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePrivKeyEd25519(w, v)
			return wBuf
		}())
	case *PrivKeyEd25519:
		codonEncodeByteSlice(int(getMagicNum("PrivKeyEd25519")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePrivKeyEd25519(w, *v)
			return wBuf
		}())
	case PrivKeySecp256k1:
		codonEncodeByteSlice(int(getMagicNum("PrivKeySecp256k1")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePrivKeySecp256k1(w, v)
			return wBuf
		}())
	case *PrivKeySecp256k1:
		codonEncodeByteSlice(int(getMagicNum("PrivKeySecp256k1")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePrivKeySecp256k1(w, *v)
			return wBuf
		}())
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func RandPrivKey(r RandSrc) PrivKey {
	switch r.GetUint() % 2 {
	case 0:
		return RandPrivKeyEd25519(r)
	case 1:
		return RandPrivKeySecp256k1(r)
	default:
		panic("Unknown Type.")
	} // end of switch
} // end of func
func DeepCopyPrivKey(x PrivKey) PrivKey {
	switch v := x.(type) {
	case PrivKeyEd25519:
		res := DeepCopyPrivKeyEd25519(v)
		return res
	case *PrivKeyEd25519:
		res := DeepCopyPrivKeyEd25519(*v)
		return &res
	case PrivKeySecp256k1:
		res := DeepCopyPrivKeySecp256k1(v)
		return res
	case *PrivKeySecp256k1:
		res := DeepCopyPrivKeySecp256k1(*v)
		return &res
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
// Interface
func DecodeMsg(bz []byte) (v Msg, total int, err error) {

	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return v, total, err
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)
	switch magicNum {
	case 0x26f1078d:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgBeginRedelegate
		tmp, n, err = DecodeMsgBeginRedelegate(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x7ead4f18:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgCreateValidator
		tmp, n, err = DecodeMsgCreateValidator(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x9e3279b8:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgDelegate
		tmp, n, err = DecodeMsgDelegate(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xdb734cea:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgDeposit
		tmp, n, err = DecodeMsgDeposit(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xd7a9fe09:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgEditValidator
		tmp, n, err = DecodeMsgEditValidator(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xdfd37740:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgMultiSend
		tmp, n, err = DecodeMsgMultiSend(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xe22effd4:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgSend
		tmp, n, err = DecodeMsgSend(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x26a788d0:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgSetWithdrawAddress
		tmp, n, err = DecodeMsgSetWithdrawAddress(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x80c4f115:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgUndelegate
		tmp, n, err = DecodeMsgUndelegate(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xb5a46e8b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgUnjail
		tmp, n, err = DecodeMsgUnjail(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xf02dad6d:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgVerifyInvariant
		tmp, n, err = DecodeMsgVerifyInvariant(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xba0d79e9:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgVote
		tmp, n, err = DecodeMsgVote(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x53b7132b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgWithdrawDelegatorReward
		tmp, n, err = DecodeMsgWithdrawDelegatorReward(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x4eb35554:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgWithdrawValidatorCommission
		tmp, n, err = DecodeMsgWithdrawValidatorCommission(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	default:
		panic("Unknown type")
	} // end of switch
	return v, n, nil
} // end of DecodeMsg
func EncodeMsg(w *[]byte, x interface{}) {
	switch v := x.(type) {
	case MsgBeginRedelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgBeginRedelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgBeginRedelegate(w, v)
			return wBuf
		}())
	case *MsgBeginRedelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgBeginRedelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgBeginRedelegate(w, *v)
			return wBuf
		}())
	case MsgCreateValidator:
		codonEncodeByteSlice(int(getMagicNum("MsgCreateValidator")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgCreateValidator(w, v)
			return wBuf
		}())
	case *MsgCreateValidator:
		codonEncodeByteSlice(int(getMagicNum("MsgCreateValidator")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgCreateValidator(w, *v)
			return wBuf
		}())
	case MsgDelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgDelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgDelegate(w, v)
			return wBuf
		}())
	case *MsgDelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgDelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgDelegate(w, *v)
			return wBuf
		}())
	case MsgDeposit:
		codonEncodeByteSlice(int(getMagicNum("MsgDeposit")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgDeposit(w, v)
			return wBuf
		}())
	case *MsgDeposit:
		codonEncodeByteSlice(int(getMagicNum("MsgDeposit")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgDeposit(w, *v)
			return wBuf
		}())
	case MsgEditValidator:
		codonEncodeByteSlice(int(getMagicNum("MsgEditValidator")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgEditValidator(w, v)
			return wBuf
		}())
	case *MsgEditValidator:
		codonEncodeByteSlice(int(getMagicNum("MsgEditValidator")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgEditValidator(w, *v)
			return wBuf
		}())
	case MsgMultiSend:
		codonEncodeByteSlice(int(getMagicNum("MsgMultiSend")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgMultiSend(w, v)
			return wBuf
		}())
	case *MsgMultiSend:
		codonEncodeByteSlice(int(getMagicNum("MsgMultiSend")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgMultiSend(w, *v)
			return wBuf
		}())
	case MsgSend:
		codonEncodeByteSlice(int(getMagicNum("MsgSend")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgSend(w, v)
			return wBuf
		}())
	case *MsgSend:
		codonEncodeByteSlice(int(getMagicNum("MsgSend")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgSend(w, *v)
			return wBuf
		}())
	case MsgSetWithdrawAddress:
		codonEncodeByteSlice(int(getMagicNum("MsgSetWithdrawAddress")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgSetWithdrawAddress(w, v)
			return wBuf
		}())
	case *MsgSetWithdrawAddress:
		codonEncodeByteSlice(int(getMagicNum("MsgSetWithdrawAddress")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgSetWithdrawAddress(w, *v)
			return wBuf
		}())
	case MsgUndelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgUndelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgUndelegate(w, v)
			return wBuf
		}())
	case *MsgUndelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgUndelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgUndelegate(w, *v)
			return wBuf
		}())
	case MsgUnjail:
		codonEncodeByteSlice(int(getMagicNum("MsgUnjail")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgUnjail(w, v)
			return wBuf
		}())
	case *MsgUnjail:
		codonEncodeByteSlice(int(getMagicNum("MsgUnjail")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgUnjail(w, *v)
			return wBuf
		}())
	case MsgVerifyInvariant:
		codonEncodeByteSlice(int(getMagicNum("MsgVerifyInvariant")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgVerifyInvariant(w, v)
			return wBuf
		}())
	case *MsgVerifyInvariant:
		codonEncodeByteSlice(int(getMagicNum("MsgVerifyInvariant")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgVerifyInvariant(w, *v)
			return wBuf
		}())
	case MsgVote:
		codonEncodeByteSlice(int(getMagicNum("MsgVote")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgVote(w, v)
			return wBuf
		}())
	case *MsgVote:
		codonEncodeByteSlice(int(getMagicNum("MsgVote")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgVote(w, *v)
			return wBuf
		}())
	case MsgWithdrawDelegatorReward:
		codonEncodeByteSlice(int(getMagicNum("MsgWithdrawDelegatorReward")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgWithdrawDelegatorReward(w, v)
			return wBuf
		}())
	case *MsgWithdrawDelegatorReward:
		codonEncodeByteSlice(int(getMagicNum("MsgWithdrawDelegatorReward")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgWithdrawDelegatorReward(w, *v)
			return wBuf
		}())
	case MsgWithdrawValidatorCommission:
		codonEncodeByteSlice(int(getMagicNum("MsgWithdrawValidatorCommission")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgWithdrawValidatorCommission(w, v)
			return wBuf
		}())
	case *MsgWithdrawValidatorCommission:
		codonEncodeByteSlice(int(getMagicNum("MsgWithdrawValidatorCommission")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgWithdrawValidatorCommission(w, *v)
			return wBuf
		}())
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func RandMsg(r RandSrc) Msg {
	switch r.GetUint() % 14 {
	case 0:
		return RandMsgBeginRedelegate(r)
	case 1:
		return RandMsgCreateValidator(r)
	case 2:
		return RandMsgDelegate(r)
	case 3:
		return RandMsgDeposit(r)
	case 4:
		return RandMsgEditValidator(r)
	case 5:
		return RandMsgMultiSend(r)
	case 6:
		return RandMsgSend(r)
	case 7:
		return RandMsgSetWithdrawAddress(r)
	case 8:
		return RandMsgUndelegate(r)
	case 9:
		return RandMsgUnjail(r)
	case 10:
		return RandMsgVerifyInvariant(r)
	case 11:
		return RandMsgVote(r)
	case 12:
		return RandMsgWithdrawDelegatorReward(r)
	case 13:
		return RandMsgWithdrawValidatorCommission(r)
	default:
		panic("Unknown Type.")
	} // end of switch
} // end of func
func DeepCopyMsg(x Msg) Msg {
	switch v := x.(type) {
	case MsgBeginRedelegate:
		res := DeepCopyMsgBeginRedelegate(v)
		return res
	case *MsgBeginRedelegate:
		res := DeepCopyMsgBeginRedelegate(*v)
		return &res
	case MsgCreateValidator:
		res := DeepCopyMsgCreateValidator(v)
		return res
	case *MsgCreateValidator:
		res := DeepCopyMsgCreateValidator(*v)
		return &res
	case MsgDelegate:
		res := DeepCopyMsgDelegate(v)
		return res
	case *MsgDelegate:
		res := DeepCopyMsgDelegate(*v)
		return &res
	case MsgDeposit:
		res := DeepCopyMsgDeposit(v)
		return res
	case *MsgDeposit:
		res := DeepCopyMsgDeposit(*v)
		return &res
	case MsgEditValidator:
		res := DeepCopyMsgEditValidator(v)
		return res
	case *MsgEditValidator:
		res := DeepCopyMsgEditValidator(*v)
		return &res
	case MsgMultiSend:
		res := DeepCopyMsgMultiSend(v)
		return res
	case *MsgMultiSend:
		res := DeepCopyMsgMultiSend(*v)
		return &res
	case MsgSend:
		res := DeepCopyMsgSend(v)
		return res
	case *MsgSend:
		res := DeepCopyMsgSend(*v)
		return &res
	case MsgSetWithdrawAddress:
		res := DeepCopyMsgSetWithdrawAddress(v)
		return res
	case *MsgSetWithdrawAddress:
		res := DeepCopyMsgSetWithdrawAddress(*v)
		return &res
	case MsgUndelegate:
		res := DeepCopyMsgUndelegate(v)
		return res
	case *MsgUndelegate:
		res := DeepCopyMsgUndelegate(*v)
		return &res
	case MsgUnjail:
		res := DeepCopyMsgUnjail(v)
		return res
	case *MsgUnjail:
		res := DeepCopyMsgUnjail(*v)
		return &res
	case MsgVerifyInvariant:
		res := DeepCopyMsgVerifyInvariant(v)
		return res
	case *MsgVerifyInvariant:
		res := DeepCopyMsgVerifyInvariant(*v)
		return &res
	case MsgVote:
		res := DeepCopyMsgVote(v)
		return res
	case *MsgVote:
		res := DeepCopyMsgVote(*v)
		return &res
	case MsgWithdrawDelegatorReward:
		res := DeepCopyMsgWithdrawDelegatorReward(v)
		return res
	case *MsgWithdrawDelegatorReward:
		res := DeepCopyMsgWithdrawDelegatorReward(*v)
		return &res
	case MsgWithdrawValidatorCommission:
		res := DeepCopyMsgWithdrawValidatorCommission(v)
		return res
	case *MsgWithdrawValidatorCommission:
		res := DeepCopyMsgWithdrawValidatorCommission(*v)
		return &res
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
// Interface
func DecodeAccount(bz []byte) (v Account, total int, err error) {

	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return v, total, err
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)
	switch magicNum {
	case 0xf46b9d99:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp BaseAccount
		tmp, n, err = DecodeBaseAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = &tmp
		return
	case 0x0066f84e:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp BaseVestingAccount
		tmp, n, err = DecodeBaseVestingAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x033a454b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ContinuousVestingAccount
		tmp, n, err = DecodeContinuousVestingAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xcf68c13b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp DelayedVestingAccount
		tmp, n, err = DecodeDelayedVestingAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xae3b1d25:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ModuleAccount
		tmp, n, err = DecodeModuleAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	default:
		panic("Unknown type")
	} // end of switch
	return v, n, nil
} // end of DecodeAccount
func EncodeAccount(w *[]byte, x interface{}) {
	switch v := x.(type) {
	case BaseAccount:
		codonEncodeByteSlice(int(getMagicNum("BaseAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeBaseAccount(w, v)
			return wBuf
		}())
	case *BaseAccount:
		codonEncodeByteSlice(int(getMagicNum("BaseAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeBaseAccount(w, *v)
			return wBuf
		}())
	case BaseVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("BaseVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeBaseVestingAccount(w, v)
			return wBuf
		}())
	case *BaseVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("BaseVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeBaseVestingAccount(w, *v)
			return wBuf
		}())
	case ContinuousVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("ContinuousVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeContinuousVestingAccount(w, v)
			return wBuf
		}())
	case *ContinuousVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("ContinuousVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeContinuousVestingAccount(w, *v)
			return wBuf
		}())
	case DelayedVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("DelayedVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDelayedVestingAccount(w, v)
			return wBuf
		}())
	case *DelayedVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("DelayedVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDelayedVestingAccount(w, *v)
			return wBuf
		}())
	case ModuleAccount:
		codonEncodeByteSlice(int(getMagicNum("ModuleAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeModuleAccount(w, v)
			return wBuf
		}())
	case *ModuleAccount:
		codonEncodeByteSlice(int(getMagicNum("ModuleAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeModuleAccount(w, *v)
			return wBuf
		}())
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func RandAccount(r RandSrc) Account {
	switch r.GetUint() % 5 {
	case 0:
		tmp := RandBaseAccount(r)
		return &tmp
	case 1:
		return RandBaseVestingAccount(r)
	case 2:
		return RandContinuousVestingAccount(r)
	case 3:
		return RandDelayedVestingAccount(r)
	case 4:
		return RandModuleAccount(r)
	default:
		panic("Unknown Type.")
	} // end of switch
} // end of func
func DeepCopyAccount(x Account) Account {
	switch v := x.(type) {
	case *BaseAccount:
		res := DeepCopyBaseAccount(*v)
		return &res
	case BaseVestingAccount:
		res := DeepCopyBaseVestingAccount(v)
		return res
	case *BaseVestingAccount:
		res := DeepCopyBaseVestingAccount(*v)
		return &res
	case ContinuousVestingAccount:
		res := DeepCopyContinuousVestingAccount(v)
		return res
	case *ContinuousVestingAccount:
		res := DeepCopyContinuousVestingAccount(*v)
		return &res
	case DelayedVestingAccount:
		res := DeepCopyDelayedVestingAccount(v)
		return res
	case *DelayedVestingAccount:
		res := DeepCopyDelayedVestingAccount(*v)
		return &res
	case ModuleAccount:
		res := DeepCopyModuleAccount(v)
		return res
	case *ModuleAccount:
		res := DeepCopyModuleAccount(*v)
		return &res
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
// Interface
func DecodeVestingAccount(bz []byte) (v VestingAccount, total int, err error) {

	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return v, total, err
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)
	switch magicNum {
	case 0x033a454b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ContinuousVestingAccount
		tmp, n, err = DecodeContinuousVestingAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = &tmp
		return
	case 0xcf68c13b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp DelayedVestingAccount
		tmp, n, err = DecodeDelayedVestingAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = &tmp
		return
	default:
		panic("Unknown type")
	} // end of switch
	return v, n, nil
} // end of DecodeVestingAccount
func EncodeVestingAccount(w *[]byte, x interface{}) {
	switch v := x.(type) {
	case ContinuousVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("ContinuousVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeContinuousVestingAccount(w, v)
			return wBuf
		}())
	case *ContinuousVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("ContinuousVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeContinuousVestingAccount(w, *v)
			return wBuf
		}())
	case DelayedVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("DelayedVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDelayedVestingAccount(w, v)
			return wBuf
		}())
	case *DelayedVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("DelayedVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDelayedVestingAccount(w, *v)
			return wBuf
		}())
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func RandVestingAccount(r RandSrc) VestingAccount {
	switch r.GetUint() % 2 {
	case 0:
		tmp := RandContinuousVestingAccount(r)
		return &tmp
	case 1:
		tmp := RandDelayedVestingAccount(r)
		return &tmp
	default:
		panic("Unknown Type.")
	} // end of switch
} // end of func
func DeepCopyVestingAccount(x VestingAccount) VestingAccount {
	switch v := x.(type) {
	case *ContinuousVestingAccount:
		res := DeepCopyContinuousVestingAccount(*v)
		return &res
	case *DelayedVestingAccount:
		res := DeepCopyDelayedVestingAccount(*v)
		return &res
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
// Interface
func DecodeContent(bz []byte) (v Content, total int, err error) {

	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return v, total, err
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)
	switch magicNum {
	case 0x16ad5d1f:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp CommunityPoolSpendProposal
		tmp, n, err = DecodeCommunityPoolSpendProposal(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x53952531:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ParameterChangeProposal
		tmp, n, err = DecodeParameterChangeProposal(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x0dad94a2:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp SoftwareUpgradeProposal
		tmp, n, err = DecodeSoftwareUpgradeProposal(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xd0adb3cf:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp TextProposal
		tmp, n, err = DecodeTextProposal(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	default:
		panic("Unknown type")
	} // end of switch
	return v, n, nil
} // end of DecodeContent
func EncodeContent(w *[]byte, x interface{}) {
	switch v := x.(type) {
	case CommunityPoolSpendProposal:
		codonEncodeByteSlice(int(getMagicNum("CommunityPoolSpendProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeCommunityPoolSpendProposal(w, v)
			return wBuf
		}())
	case *CommunityPoolSpendProposal:
		codonEncodeByteSlice(int(getMagicNum("CommunityPoolSpendProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeCommunityPoolSpendProposal(w, *v)
			return wBuf
		}())
	case ParameterChangeProposal:
		codonEncodeByteSlice(int(getMagicNum("ParameterChangeProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeParameterChangeProposal(w, v)
			return wBuf
		}())
	case *ParameterChangeProposal:
		codonEncodeByteSlice(int(getMagicNum("ParameterChangeProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeParameterChangeProposal(w, *v)
			return wBuf
		}())
	case SoftwareUpgradeProposal:
		codonEncodeByteSlice(int(getMagicNum("SoftwareUpgradeProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSoftwareUpgradeProposal(w, v)
			return wBuf
		}())
	case *SoftwareUpgradeProposal:
		codonEncodeByteSlice(int(getMagicNum("SoftwareUpgradeProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSoftwareUpgradeProposal(w, *v)
			return wBuf
		}())
	case TextProposal:
		codonEncodeByteSlice(int(getMagicNum("TextProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeTextProposal(w, v)
			return wBuf
		}())
	case *TextProposal:
		codonEncodeByteSlice(int(getMagicNum("TextProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeTextProposal(w, *v)
			return wBuf
		}())
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func RandContent(r RandSrc) Content {
	switch r.GetUint() % 4 {
	case 0:
		return RandCommunityPoolSpendProposal(r)
	case 1:
		return RandParameterChangeProposal(r)
	case 2:
		return RandSoftwareUpgradeProposal(r)
	case 3:
		return RandTextProposal(r)
	default:
		panic("Unknown Type.")
	} // end of switch
} // end of func
func DeepCopyContent(x Content) Content {
	switch v := x.(type) {
	case CommunityPoolSpendProposal:
		res := DeepCopyCommunityPoolSpendProposal(v)
		return res
	case *CommunityPoolSpendProposal:
		res := DeepCopyCommunityPoolSpendProposal(*v)
		return &res
	case ParameterChangeProposal:
		res := DeepCopyParameterChangeProposal(v)
		return res
	case *ParameterChangeProposal:
		res := DeepCopyParameterChangeProposal(*v)
		return &res
	case SoftwareUpgradeProposal:
		res := DeepCopySoftwareUpgradeProposal(v)
		return res
	case *SoftwareUpgradeProposal:
		res := DeepCopySoftwareUpgradeProposal(*v)
		return &res
	case TextProposal:
		res := DeepCopyTextProposal(v)
		return res
	case *TextProposal:
		res := DeepCopyTextProposal(*v)
		return &res
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
// Interface
func DecodeTx(bz []byte) (v Tx, total int, err error) {

	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return v, total, err
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)
	switch magicNum {
	case 0x5133aaf7:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp StdTx
		tmp, n, err = DecodeStdTx(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	default:
		panic("Unknown type")
	} // end of switch
	return v, n, nil
} // end of DecodeTx
func EncodeTx(w *[]byte, x interface{}) {
	switch v := x.(type) {
	case StdTx:
		codonEncodeByteSlice(int(getMagicNum("StdTx")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeStdTx(w, v)
			return wBuf
		}())
	case *StdTx:
		codonEncodeByteSlice(int(getMagicNum("StdTx")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeStdTx(w, *v)
			return wBuf
		}())
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func RandTx(r RandSrc) Tx {
	switch r.GetUint() % 1 {
	case 0:
		return RandStdTx(r)
	default:
		panic("Unknown Type.")
	} // end of switch
} // end of func
func DeepCopyTx(x Tx) Tx {
	switch v := x.(type) {
	case StdTx:
		res := DeepCopyStdTx(v)
		return res
	case *StdTx:
		res := DeepCopyStdTx(*v)
		return &res
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
// Interface
func DecodeModuleAccountI(bz []byte) (v ModuleAccountI, total int, err error) {

	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return v, total, err
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)
	switch magicNum {
	case 0xae3b1d25:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ModuleAccount
		tmp, n, err = DecodeModuleAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	default:
		panic("Unknown type")
	} // end of switch
	return v, n, nil
} // end of DecodeModuleAccountI
func EncodeModuleAccountI(w *[]byte, x interface{}) {
	switch v := x.(type) {
	case ModuleAccount:
		codonEncodeByteSlice(int(getMagicNum("ModuleAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeModuleAccount(w, v)
			return wBuf
		}())
	case *ModuleAccount:
		codonEncodeByteSlice(int(getMagicNum("ModuleAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeModuleAccount(w, *v)
			return wBuf
		}())
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func RandModuleAccountI(r RandSrc) ModuleAccountI {
	switch r.GetUint() % 1 {
	case 0:
		return RandModuleAccount(r)
	default:
		panic("Unknown Type.")
	} // end of switch
} // end of func
func DeepCopyModuleAccountI(x ModuleAccountI) ModuleAccountI {
	switch v := x.(type) {
	case ModuleAccount:
		res := DeepCopyModuleAccount(v)
		return res
	case *ModuleAccount:
		res := DeepCopyModuleAccount(*v)
		return &res
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
// Interface
func DecodeSupplyI(bz []byte) (v SupplyI, total int, err error) {

	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return v, total, err
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)
	switch magicNum {
	case 0xe56842bf:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp Supply
		tmp, n, err = DecodeSupply(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	default:
		panic("Unknown type")
	} // end of switch
	return v, n, nil
} // end of DecodeSupplyI
func EncodeSupplyI(w *[]byte, x interface{}) {
	switch v := x.(type) {
	case Supply:
		codonEncodeByteSlice(int(getMagicNum("Supply")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSupply(w, v)
			return wBuf
		}())
	case *Supply:
		codonEncodeByteSlice(int(getMagicNum("Supply")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSupply(w, *v)
			return wBuf
		}())
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func RandSupplyI(r RandSrc) SupplyI {
	switch r.GetUint() % 1 {
	case 0:
		return RandSupply(r)
	default:
		panic("Unknown Type.")
	} // end of switch
} // end of func
func DeepCopySupplyI(x SupplyI) SupplyI {
	switch v := x.(type) {
	case Supply:
		res := DeepCopySupply(v)
		return res
	case *Supply:
		res := DeepCopySupply(*v)
		return &res
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func getMagicNum(name string) uint32 {
	switch name {
	case "AccAddress":
		return 0xe74c9d00
	case "AccAddressList":
		return 0x4d724825
	case "BaseAccount":
		return 0xf46b9d99
	case "BaseVestingAccount":
		return 0x0066f84e
	case "BondStatus":
		return 0xb9ec8316
	case "Coin":
		return 0xbaaa4102
	case "CommitInfo":
		return 0x39091a02
	case "CommunityPoolSpendProposal":
		return 0x16ad5d1f
	case "ConsAddress":
		return 0x8be3351c
	case "ContinuousVestingAccount":
		return 0x033a454b
	case "DecCoin":
		return 0x54c7a338
	case "DecCoins":
		return 0x22302239
	case "DelayedVestingAccount":
		return 0xcf68c13b
	case "Delegation":
		return 0x069ed9a6
	case "DelegatorStartingInfo":
		return 0x3a833fbf
	case "FeePool":
		return 0xc45cc112
	case "Input":
		return 0x5850ec36
	case "ModuleAccount":
		return 0xae3b1d25
	case "MsgBeginRedelegate":
		return 0x26f1078d
	case "MsgCreateValidator":
		return 0x7ead4f18
	case "MsgDelegate":
		return 0x9e3279b8
	case "MsgDeposit":
		return 0xdb734cea
	case "MsgEditValidator":
		return 0xd7a9fe09
	case "MsgMultiSend":
		return 0xdfd37740
	case "MsgSend":
		return 0xe22effd4
	case "MsgSetWithdrawAddress":
		return 0x26a788d0
	case "MsgUndelegate":
		return 0x80c4f115
	case "MsgUnjail":
		return 0xb5a46e8b
	case "MsgVerifyInvariant":
		return 0xf02dad6d
	case "MsgVote":
		return 0xba0d79e9
	case "MsgWithdrawDelegatorReward":
		return 0x53b7132b
	case "MsgWithdrawValidatorCommission":
		return 0x4eb35554
	case "Output":
		return 0x703f43b2
	case "ParamChange":
		return 0xbc52fa42
	case "ParameterChangeProposal":
		return 0x53952531
	case "PrivKeyEd25519":
		return 0x56ba5e9e
	case "PrivKeySecp256k1":
		return 0x08171053
	case "PubKeyEd25519":
		return 0xa5b24c72
	case "PubKeyMultisigThreshold":
		return 0x7f05210e
	case "PubKeySecp256k1":
		return 0x60f9a133
	case "SdkDec":
		return 0x0a456583
	case "SdkInt":
		return 0x8918d2bd
	case "SignedMsgType":
		return 0x90d93443
	case "SoftwareUpgradeProposal":
		return 0x0dad94a2
	case "StdSignature":
		return 0x461d2af7
	case "StdTx":
		return 0x5133aaf7
	case "StoreInfo":
		return 0xe96e31e0
	case "Supply":
		return 0xe56842bf
	case "TextProposal":
		return 0xd0adb3cf
	case "ValAddress":
		return 0x8b2f2762
	case "ValAddressList":
		return 0x4a608e0b
	case "Validator":
		return 0x02460a0b
	case "ValidatorCurrentRewards":
		return 0x1cfe14b2
	case "ValidatorHistoricalRewards":
		return 0x59012970
	case "ValidatorSigningInfo":
		return 0x2cfacfb0
	case "ValidatorSlashEvent":
		return 0x76e07d4c
	case "Vote":
		return 0xe6f455cd
	case "VoteOption":
		return 0xac41d0aa
	case "int64":
		return 0x7a9622bc
	case "uint64":
		return 0xd5f1d224
	} // end of switch
	panic("Should not reach here")
	return 0
} // end of getMagicNum
func getMagicNumOfVar(x interface{}) (uint32, bool) {
	switch x.(type) {
	case *AccAddress, AccAddress:
		return 0xe74c9d00, true
	case *AccAddressList, AccAddressList:
		return 0x4d724825, true
	case *BaseAccount, BaseAccount:
		return 0xf46b9d99, true
	case *BaseVestingAccount, BaseVestingAccount:
		return 0x0066f84e, true
	case *BondStatus, BondStatus:
		return 0xb9ec8316, true
	case *Coin, Coin:
		return 0xbaaa4102, true
	case *CommitInfo, CommitInfo:
		return 0x39091a02, true
	case *CommunityPoolSpendProposal, CommunityPoolSpendProposal:
		return 0x16ad5d1f, true
	case *ConsAddress, ConsAddress:
		return 0x8be3351c, true
	case *ContinuousVestingAccount, ContinuousVestingAccount:
		return 0x033a454b, true
	case *DecCoin, DecCoin:
		return 0x54c7a338, true
	case *DecCoins, DecCoins:
		return 0x22302239, true
	case *DelayedVestingAccount, DelayedVestingAccount:
		return 0xcf68c13b, true
	case *Delegation, Delegation:
		return 0x069ed9a6, true
	case *DelegatorStartingInfo, DelegatorStartingInfo:
		return 0x3a833fbf, true
	case *FeePool, FeePool:
		return 0xc45cc112, true
	case *Input, Input:
		return 0x5850ec36, true
	case *ModuleAccount, ModuleAccount:
		return 0xae3b1d25, true
	case *MsgBeginRedelegate, MsgBeginRedelegate:
		return 0x26f1078d, true
	case *MsgCreateValidator, MsgCreateValidator:
		return 0x7ead4f18, true
	case *MsgDelegate, MsgDelegate:
		return 0x9e3279b8, true
	case *MsgDeposit, MsgDeposit:
		return 0xdb734cea, true
	case *MsgEditValidator, MsgEditValidator:
		return 0xd7a9fe09, true
	case *MsgMultiSend, MsgMultiSend:
		return 0xdfd37740, true
	case *MsgSend, MsgSend:
		return 0xe22effd4, true
	case *MsgSetWithdrawAddress, MsgSetWithdrawAddress:
		return 0x26a788d0, true
	case *MsgUndelegate, MsgUndelegate:
		return 0x80c4f115, true
	case *MsgUnjail, MsgUnjail:
		return 0xb5a46e8b, true
	case *MsgVerifyInvariant, MsgVerifyInvariant:
		return 0xf02dad6d, true
	case *MsgVote, MsgVote:
		return 0xba0d79e9, true
	case *MsgWithdrawDelegatorReward, MsgWithdrawDelegatorReward:
		return 0x53b7132b, true
	case *MsgWithdrawValidatorCommission, MsgWithdrawValidatorCommission:
		return 0x4eb35554, true
	case *Output, Output:
		return 0x703f43b2, true
	case *ParamChange, ParamChange:
		return 0xbc52fa42, true
	case *ParameterChangeProposal, ParameterChangeProposal:
		return 0x53952531, true
	case *PrivKeyEd25519, PrivKeyEd25519:
		return 0x56ba5e9e, true
	case *PrivKeySecp256k1, PrivKeySecp256k1:
		return 0x08171053, true
	case *PubKeyEd25519, PubKeyEd25519:
		return 0xa5b24c72, true
	case *PubKeyMultisigThreshold, PubKeyMultisigThreshold:
		return 0x7f05210e, true
	case *PubKeySecp256k1, PubKeySecp256k1:
		return 0x60f9a133, true
	case *SdkDec, SdkDec:
		return 0x0a456583, true
	case *SdkInt, SdkInt:
		return 0x8918d2bd, true
	case *SignedMsgType, SignedMsgType:
		return 0x90d93443, true
	case *SoftwareUpgradeProposal, SoftwareUpgradeProposal:
		return 0x0dad94a2, true
	case *StdSignature, StdSignature:
		return 0x461d2af7, true
	case *StdTx, StdTx:
		return 0x5133aaf7, true
	case *StoreInfo, StoreInfo:
		return 0xe96e31e0, true
	case *Supply, Supply:
		return 0xe56842bf, true
	case *TextProposal, TextProposal:
		return 0xd0adb3cf, true
	case *ValAddress, ValAddress:
		return 0x8b2f2762, true
	case *ValAddressList, ValAddressList:
		return 0x4a608e0b, true
	case *Validator, Validator:
		return 0x02460a0b, true
	case *ValidatorCurrentRewards, ValidatorCurrentRewards:
		return 0x1cfe14b2, true
	case *ValidatorHistoricalRewards, ValidatorHistoricalRewards:
		return 0x59012970, true
	case *ValidatorSigningInfo, ValidatorSigningInfo:
		return 0x2cfacfb0, true
	case *ValidatorSlashEvent, ValidatorSlashEvent:
		return 0x76e07d4c, true
	case *Vote, Vote:
		return 0xe6f455cd, true
	case *VoteOption, VoteOption:
		return 0xac41d0aa, true
	case *int64, int64:
		return 0x7a9622bc, true
	case *uint64, uint64:
		return 0xd5f1d224, true
	default:
		return 0, false
	} // end of switch
} // end of func
func EncodeAny(w *[]byte, x interface{}) {
	switch v := x.(type) {
	case AccAddress:
		codonEncodeByteSlice(int(getMagicNum("AccAddress")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeAccAddress(w, v)
			return wBuf
		}())
	case *AccAddress:
		codonEncodeByteSlice(int(getMagicNum("AccAddress")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeAccAddress(w, *v)
			return wBuf
		}())
	case AccAddressList:
		codonEncodeByteSlice(int(getMagicNum("AccAddressList")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeAccAddressList(w, v)
			return wBuf
		}())
	case *AccAddressList:
		codonEncodeByteSlice(int(getMagicNum("AccAddressList")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeAccAddressList(w, *v)
			return wBuf
		}())
	case BaseAccount:
		codonEncodeByteSlice(int(getMagicNum("BaseAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeBaseAccount(w, v)
			return wBuf
		}())
	case *BaseAccount:
		codonEncodeByteSlice(int(getMagicNum("BaseAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeBaseAccount(w, *v)
			return wBuf
		}())
	case BaseVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("BaseVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeBaseVestingAccount(w, v)
			return wBuf
		}())
	case *BaseVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("BaseVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeBaseVestingAccount(w, *v)
			return wBuf
		}())
	case BondStatus:
		codonEncodeByteSlice(int(getMagicNum("BondStatus")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeBondStatus(w, v)
			return wBuf
		}())
	case *BondStatus:
		codonEncodeByteSlice(int(getMagicNum("BondStatus")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeBondStatus(w, *v)
			return wBuf
		}())
	case Coin:
		codonEncodeByteSlice(int(getMagicNum("Coin")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeCoin(w, v)
			return wBuf
		}())
	case *Coin:
		codonEncodeByteSlice(int(getMagicNum("Coin")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeCoin(w, *v)
			return wBuf
		}())
	case CommitInfo:
		codonEncodeByteSlice(int(getMagicNum("CommitInfo")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeCommitInfo(w, v)
			return wBuf
		}())
	case *CommitInfo:
		codonEncodeByteSlice(int(getMagicNum("CommitInfo")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeCommitInfo(w, *v)
			return wBuf
		}())
	case CommunityPoolSpendProposal:
		codonEncodeByteSlice(int(getMagicNum("CommunityPoolSpendProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeCommunityPoolSpendProposal(w, v)
			return wBuf
		}())
	case *CommunityPoolSpendProposal:
		codonEncodeByteSlice(int(getMagicNum("CommunityPoolSpendProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeCommunityPoolSpendProposal(w, *v)
			return wBuf
		}())
	case ConsAddress:
		codonEncodeByteSlice(int(getMagicNum("ConsAddress")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeConsAddress(w, v)
			return wBuf
		}())
	case *ConsAddress:
		codonEncodeByteSlice(int(getMagicNum("ConsAddress")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeConsAddress(w, *v)
			return wBuf
		}())
	case ContinuousVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("ContinuousVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeContinuousVestingAccount(w, v)
			return wBuf
		}())
	case *ContinuousVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("ContinuousVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeContinuousVestingAccount(w, *v)
			return wBuf
		}())
	case DecCoin:
		codonEncodeByteSlice(int(getMagicNum("DecCoin")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDecCoin(w, v)
			return wBuf
		}())
	case *DecCoin:
		codonEncodeByteSlice(int(getMagicNum("DecCoin")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDecCoin(w, *v)
			return wBuf
		}())
	case DecCoins:
		codonEncodeByteSlice(int(getMagicNum("DecCoins")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDecCoins(w, v)
			return wBuf
		}())
	case *DecCoins:
		codonEncodeByteSlice(int(getMagicNum("DecCoins")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDecCoins(w, *v)
			return wBuf
		}())
	case DelayedVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("DelayedVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDelayedVestingAccount(w, v)
			return wBuf
		}())
	case *DelayedVestingAccount:
		codonEncodeByteSlice(int(getMagicNum("DelayedVestingAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDelayedVestingAccount(w, *v)
			return wBuf
		}())
	case Delegation:
		codonEncodeByteSlice(int(getMagicNum("Delegation")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDelegation(w, v)
			return wBuf
		}())
	case *Delegation:
		codonEncodeByteSlice(int(getMagicNum("Delegation")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDelegation(w, *v)
			return wBuf
		}())
	case DelegatorStartingInfo:
		codonEncodeByteSlice(int(getMagicNum("DelegatorStartingInfo")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDelegatorStartingInfo(w, v)
			return wBuf
		}())
	case *DelegatorStartingInfo:
		codonEncodeByteSlice(int(getMagicNum("DelegatorStartingInfo")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeDelegatorStartingInfo(w, *v)
			return wBuf
		}())
	case FeePool:
		codonEncodeByteSlice(int(getMagicNum("FeePool")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeFeePool(w, v)
			return wBuf
		}())
	case *FeePool:
		codonEncodeByteSlice(int(getMagicNum("FeePool")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeFeePool(w, *v)
			return wBuf
		}())
	case Input:
		codonEncodeByteSlice(int(getMagicNum("Input")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeInput(w, v)
			return wBuf
		}())
	case *Input:
		codonEncodeByteSlice(int(getMagicNum("Input")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeInput(w, *v)
			return wBuf
		}())
	case ModuleAccount:
		codonEncodeByteSlice(int(getMagicNum("ModuleAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeModuleAccount(w, v)
			return wBuf
		}())
	case *ModuleAccount:
		codonEncodeByteSlice(int(getMagicNum("ModuleAccount")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeModuleAccount(w, *v)
			return wBuf
		}())
	case MsgBeginRedelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgBeginRedelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgBeginRedelegate(w, v)
			return wBuf
		}())
	case *MsgBeginRedelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgBeginRedelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgBeginRedelegate(w, *v)
			return wBuf
		}())
	case MsgCreateValidator:
		codonEncodeByteSlice(int(getMagicNum("MsgCreateValidator")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgCreateValidator(w, v)
			return wBuf
		}())
	case *MsgCreateValidator:
		codonEncodeByteSlice(int(getMagicNum("MsgCreateValidator")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgCreateValidator(w, *v)
			return wBuf
		}())
	case MsgDelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgDelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgDelegate(w, v)
			return wBuf
		}())
	case *MsgDelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgDelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgDelegate(w, *v)
			return wBuf
		}())
	case MsgDeposit:
		codonEncodeByteSlice(int(getMagicNum("MsgDeposit")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgDeposit(w, v)
			return wBuf
		}())
	case *MsgDeposit:
		codonEncodeByteSlice(int(getMagicNum("MsgDeposit")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgDeposit(w, *v)
			return wBuf
		}())
	case MsgEditValidator:
		codonEncodeByteSlice(int(getMagicNum("MsgEditValidator")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgEditValidator(w, v)
			return wBuf
		}())
	case *MsgEditValidator:
		codonEncodeByteSlice(int(getMagicNum("MsgEditValidator")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgEditValidator(w, *v)
			return wBuf
		}())
	case MsgMultiSend:
		codonEncodeByteSlice(int(getMagicNum("MsgMultiSend")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgMultiSend(w, v)
			return wBuf
		}())
	case *MsgMultiSend:
		codonEncodeByteSlice(int(getMagicNum("MsgMultiSend")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgMultiSend(w, *v)
			return wBuf
		}())
	case MsgSend:
		codonEncodeByteSlice(int(getMagicNum("MsgSend")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgSend(w, v)
			return wBuf
		}())
	case *MsgSend:
		codonEncodeByteSlice(int(getMagicNum("MsgSend")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgSend(w, *v)
			return wBuf
		}())
	case MsgSetWithdrawAddress:
		codonEncodeByteSlice(int(getMagicNum("MsgSetWithdrawAddress")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgSetWithdrawAddress(w, v)
			return wBuf
		}())
	case *MsgSetWithdrawAddress:
		codonEncodeByteSlice(int(getMagicNum("MsgSetWithdrawAddress")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgSetWithdrawAddress(w, *v)
			return wBuf
		}())
	case MsgUndelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgUndelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgUndelegate(w, v)
			return wBuf
		}())
	case *MsgUndelegate:
		codonEncodeByteSlice(int(getMagicNum("MsgUndelegate")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgUndelegate(w, *v)
			return wBuf
		}())
	case MsgUnjail:
		codonEncodeByteSlice(int(getMagicNum("MsgUnjail")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgUnjail(w, v)
			return wBuf
		}())
	case *MsgUnjail:
		codonEncodeByteSlice(int(getMagicNum("MsgUnjail")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgUnjail(w, *v)
			return wBuf
		}())
	case MsgVerifyInvariant:
		codonEncodeByteSlice(int(getMagicNum("MsgVerifyInvariant")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgVerifyInvariant(w, v)
			return wBuf
		}())
	case *MsgVerifyInvariant:
		codonEncodeByteSlice(int(getMagicNum("MsgVerifyInvariant")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgVerifyInvariant(w, *v)
			return wBuf
		}())
	case MsgVote:
		codonEncodeByteSlice(int(getMagicNum("MsgVote")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgVote(w, v)
			return wBuf
		}())
	case *MsgVote:
		codonEncodeByteSlice(int(getMagicNum("MsgVote")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgVote(w, *v)
			return wBuf
		}())
	case MsgWithdrawDelegatorReward:
		codonEncodeByteSlice(int(getMagicNum("MsgWithdrawDelegatorReward")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgWithdrawDelegatorReward(w, v)
			return wBuf
		}())
	case *MsgWithdrawDelegatorReward:
		codonEncodeByteSlice(int(getMagicNum("MsgWithdrawDelegatorReward")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgWithdrawDelegatorReward(w, *v)
			return wBuf
		}())
	case MsgWithdrawValidatorCommission:
		codonEncodeByteSlice(int(getMagicNum("MsgWithdrawValidatorCommission")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgWithdrawValidatorCommission(w, v)
			return wBuf
		}())
	case *MsgWithdrawValidatorCommission:
		codonEncodeByteSlice(int(getMagicNum("MsgWithdrawValidatorCommission")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeMsgWithdrawValidatorCommission(w, *v)
			return wBuf
		}())
	case Output:
		codonEncodeByteSlice(int(getMagicNum("Output")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeOutput(w, v)
			return wBuf
		}())
	case *Output:
		codonEncodeByteSlice(int(getMagicNum("Output")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeOutput(w, *v)
			return wBuf
		}())
	case ParamChange:
		codonEncodeByteSlice(int(getMagicNum("ParamChange")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeParamChange(w, v)
			return wBuf
		}())
	case *ParamChange:
		codonEncodeByteSlice(int(getMagicNum("ParamChange")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeParamChange(w, *v)
			return wBuf
		}())
	case ParameterChangeProposal:
		codonEncodeByteSlice(int(getMagicNum("ParameterChangeProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeParameterChangeProposal(w, v)
			return wBuf
		}())
	case *ParameterChangeProposal:
		codonEncodeByteSlice(int(getMagicNum("ParameterChangeProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeParameterChangeProposal(w, *v)
			return wBuf
		}())
	case PrivKeyEd25519:
		codonEncodeByteSlice(int(getMagicNum("PrivKeyEd25519")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePrivKeyEd25519(w, v)
			return wBuf
		}())
	case *PrivKeyEd25519:
		codonEncodeByteSlice(int(getMagicNum("PrivKeyEd25519")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePrivKeyEd25519(w, *v)
			return wBuf
		}())
	case PrivKeySecp256k1:
		codonEncodeByteSlice(int(getMagicNum("PrivKeySecp256k1")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePrivKeySecp256k1(w, v)
			return wBuf
		}())
	case *PrivKeySecp256k1:
		codonEncodeByteSlice(int(getMagicNum("PrivKeySecp256k1")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePrivKeySecp256k1(w, *v)
			return wBuf
		}())
	case PubKeyEd25519:
		codonEncodeByteSlice(int(getMagicNum("PubKeyEd25519")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeyEd25519(w, v)
			return wBuf
		}())
	case *PubKeyEd25519:
		codonEncodeByteSlice(int(getMagicNum("PubKeyEd25519")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeyEd25519(w, *v)
			return wBuf
		}())
	case PubKeyMultisigThreshold:
		codonEncodeByteSlice(int(getMagicNum("PubKeyMultisigThreshold")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeyMultisigThreshold(w, v)
			return wBuf
		}())
	case *PubKeyMultisigThreshold:
		codonEncodeByteSlice(int(getMagicNum("PubKeyMultisigThreshold")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeyMultisigThreshold(w, *v)
			return wBuf
		}())
	case PubKeySecp256k1:
		codonEncodeByteSlice(int(getMagicNum("PubKeySecp256k1")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeySecp256k1(w, v)
			return wBuf
		}())
	case *PubKeySecp256k1:
		codonEncodeByteSlice(int(getMagicNum("PubKeySecp256k1")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodePubKeySecp256k1(w, *v)
			return wBuf
		}())
	case SdkDec:
		codonEncodeByteSlice(int(getMagicNum("SdkDec")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSdkDec(w, v)
			return wBuf
		}())
	case *SdkDec:
		codonEncodeByteSlice(int(getMagicNum("SdkDec")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSdkDec(w, *v)
			return wBuf
		}())
	case SdkInt:
		codonEncodeByteSlice(int(getMagicNum("SdkInt")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSdkInt(w, v)
			return wBuf
		}())
	case *SdkInt:
		codonEncodeByteSlice(int(getMagicNum("SdkInt")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSdkInt(w, *v)
			return wBuf
		}())
	case SignedMsgType:
		codonEncodeByteSlice(int(getMagicNum("SignedMsgType")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSignedMsgType(w, v)
			return wBuf
		}())
	case *SignedMsgType:
		codonEncodeByteSlice(int(getMagicNum("SignedMsgType")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSignedMsgType(w, *v)
			return wBuf
		}())
	case SoftwareUpgradeProposal:
		codonEncodeByteSlice(int(getMagicNum("SoftwareUpgradeProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSoftwareUpgradeProposal(w, v)
			return wBuf
		}())
	case *SoftwareUpgradeProposal:
		codonEncodeByteSlice(int(getMagicNum("SoftwareUpgradeProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSoftwareUpgradeProposal(w, *v)
			return wBuf
		}())
	case StdSignature:
		codonEncodeByteSlice(int(getMagicNum("StdSignature")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeStdSignature(w, v)
			return wBuf
		}())
	case *StdSignature:
		codonEncodeByteSlice(int(getMagicNum("StdSignature")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeStdSignature(w, *v)
			return wBuf
		}())
	case StdTx:
		codonEncodeByteSlice(int(getMagicNum("StdTx")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeStdTx(w, v)
			return wBuf
		}())
	case *StdTx:
		codonEncodeByteSlice(int(getMagicNum("StdTx")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeStdTx(w, *v)
			return wBuf
		}())
	case StoreInfo:
		codonEncodeByteSlice(int(getMagicNum("StoreInfo")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeStoreInfo(w, v)
			return wBuf
		}())
	case *StoreInfo:
		codonEncodeByteSlice(int(getMagicNum("StoreInfo")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeStoreInfo(w, *v)
			return wBuf
		}())
	case Supply:
		codonEncodeByteSlice(int(getMagicNum("Supply")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSupply(w, v)
			return wBuf
		}())
	case *Supply:
		codonEncodeByteSlice(int(getMagicNum("Supply")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeSupply(w, *v)
			return wBuf
		}())
	case TextProposal:
		codonEncodeByteSlice(int(getMagicNum("TextProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeTextProposal(w, v)
			return wBuf
		}())
	case *TextProposal:
		codonEncodeByteSlice(int(getMagicNum("TextProposal")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeTextProposal(w, *v)
			return wBuf
		}())
	case ValAddress:
		codonEncodeByteSlice(int(getMagicNum("ValAddress")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValAddress(w, v)
			return wBuf
		}())
	case *ValAddress:
		codonEncodeByteSlice(int(getMagicNum("ValAddress")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValAddress(w, *v)
			return wBuf
		}())
	case ValAddressList:
		codonEncodeByteSlice(int(getMagicNum("ValAddressList")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValAddressList(w, v)
			return wBuf
		}())
	case *ValAddressList:
		codonEncodeByteSlice(int(getMagicNum("ValAddressList")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValAddressList(w, *v)
			return wBuf
		}())
	case Validator:
		codonEncodeByteSlice(int(getMagicNum("Validator")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValidator(w, v)
			return wBuf
		}())
	case *Validator:
		codonEncodeByteSlice(int(getMagicNum("Validator")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValidator(w, *v)
			return wBuf
		}())
	case ValidatorCurrentRewards:
		codonEncodeByteSlice(int(getMagicNum("ValidatorCurrentRewards")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValidatorCurrentRewards(w, v)
			return wBuf
		}())
	case *ValidatorCurrentRewards:
		codonEncodeByteSlice(int(getMagicNum("ValidatorCurrentRewards")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValidatorCurrentRewards(w, *v)
			return wBuf
		}())
	case ValidatorHistoricalRewards:
		codonEncodeByteSlice(int(getMagicNum("ValidatorHistoricalRewards")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValidatorHistoricalRewards(w, v)
			return wBuf
		}())
	case *ValidatorHistoricalRewards:
		codonEncodeByteSlice(int(getMagicNum("ValidatorHistoricalRewards")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValidatorHistoricalRewards(w, *v)
			return wBuf
		}())
	case ValidatorSigningInfo:
		codonEncodeByteSlice(int(getMagicNum("ValidatorSigningInfo")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValidatorSigningInfo(w, v)
			return wBuf
		}())
	case *ValidatorSigningInfo:
		codonEncodeByteSlice(int(getMagicNum("ValidatorSigningInfo")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValidatorSigningInfo(w, *v)
			return wBuf
		}())
	case ValidatorSlashEvent:
		codonEncodeByteSlice(int(getMagicNum("ValidatorSlashEvent")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValidatorSlashEvent(w, v)
			return wBuf
		}())
	case *ValidatorSlashEvent:
		codonEncodeByteSlice(int(getMagicNum("ValidatorSlashEvent")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeValidatorSlashEvent(w, *v)
			return wBuf
		}())
	case Vote:
		codonEncodeByteSlice(int(getMagicNum("Vote")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeVote(w, v)
			return wBuf
		}())
	case *Vote:
		codonEncodeByteSlice(int(getMagicNum("Vote")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeVote(w, *v)
			return wBuf
		}())
	case VoteOption:
		codonEncodeByteSlice(int(getMagicNum("VoteOption")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeVoteOption(w, v)
			return wBuf
		}())
	case *VoteOption:
		codonEncodeByteSlice(int(getMagicNum("VoteOption")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			EncodeVoteOption(w, *v)
			return wBuf
		}())
	case int64:
		codonEncodeByteSlice(int(getMagicNum("int64")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			Encodeint64(w, v)
			return wBuf
		}())
	case *int64:
		codonEncodeByteSlice(int(getMagicNum("int64")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			Encodeint64(w, *v)
			return wBuf
		}())
	case uint64:
		codonEncodeByteSlice(int(getMagicNum("uint64")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			Encodeuint64(w, v)
			return wBuf
		}())
	case *uint64:
		codonEncodeByteSlice(int(getMagicNum("uint64")), w, func() []byte {
			wBuf := make([]byte, 0, 64)
			w := &wBuf
			Encodeuint64(w, *v)
			return wBuf
		}())
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func DecodeAny(bz []byte) (v interface{}, total int, err error) {

	var n int
	tag := codonDecodeUint64(bz, &n, &err)
	if err != nil {
		return v, total, err
	}
	bz = bz[n:]
	total += n
	magicNum := uint32(tag >> 3)
	switch magicNum {
	case 0xe74c9d00:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp AccAddress
		tmp, n, err = DecodeAccAddress(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x4d724825:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp AccAddressList
		tmp, n, err = DecodeAccAddressList(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xf46b9d99:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp BaseAccount
		tmp, n, err = DecodeBaseAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x0066f84e:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp BaseVestingAccount
		tmp, n, err = DecodeBaseVestingAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xb9ec8316:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp BondStatus
		tmp, n, err = DecodeBondStatus(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xbaaa4102:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp Coin
		tmp, n, err = DecodeCoin(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x39091a02:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp CommitInfo
		tmp, n, err = DecodeCommitInfo(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x16ad5d1f:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp CommunityPoolSpendProposal
		tmp, n, err = DecodeCommunityPoolSpendProposal(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x8be3351c:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ConsAddress
		tmp, n, err = DecodeConsAddress(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x033a454b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ContinuousVestingAccount
		tmp, n, err = DecodeContinuousVestingAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x54c7a338:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp DecCoin
		tmp, n, err = DecodeDecCoin(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x22302239:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp DecCoins
		tmp, n, err = DecodeDecCoins(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xcf68c13b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp DelayedVestingAccount
		tmp, n, err = DecodeDelayedVestingAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x069ed9a6:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp Delegation
		tmp, n, err = DecodeDelegation(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x3a833fbf:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp DelegatorStartingInfo
		tmp, n, err = DecodeDelegatorStartingInfo(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xc45cc112:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp FeePool
		tmp, n, err = DecodeFeePool(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x5850ec36:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp Input
		tmp, n, err = DecodeInput(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xae3b1d25:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ModuleAccount
		tmp, n, err = DecodeModuleAccount(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x26f1078d:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgBeginRedelegate
		tmp, n, err = DecodeMsgBeginRedelegate(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x7ead4f18:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgCreateValidator
		tmp, n, err = DecodeMsgCreateValidator(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x9e3279b8:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgDelegate
		tmp, n, err = DecodeMsgDelegate(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xdb734cea:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgDeposit
		tmp, n, err = DecodeMsgDeposit(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xd7a9fe09:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgEditValidator
		tmp, n, err = DecodeMsgEditValidator(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xdfd37740:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgMultiSend
		tmp, n, err = DecodeMsgMultiSend(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xe22effd4:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgSend
		tmp, n, err = DecodeMsgSend(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x26a788d0:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgSetWithdrawAddress
		tmp, n, err = DecodeMsgSetWithdrawAddress(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x80c4f115:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgUndelegate
		tmp, n, err = DecodeMsgUndelegate(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xb5a46e8b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgUnjail
		tmp, n, err = DecodeMsgUnjail(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xf02dad6d:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgVerifyInvariant
		tmp, n, err = DecodeMsgVerifyInvariant(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xba0d79e9:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgVote
		tmp, n, err = DecodeMsgVote(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x53b7132b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgWithdrawDelegatorReward
		tmp, n, err = DecodeMsgWithdrawDelegatorReward(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x4eb35554:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp MsgWithdrawValidatorCommission
		tmp, n, err = DecodeMsgWithdrawValidatorCommission(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x703f43b2:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp Output
		tmp, n, err = DecodeOutput(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xbc52fa42:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ParamChange
		tmp, n, err = DecodeParamChange(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x53952531:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ParameterChangeProposal
		tmp, n, err = DecodeParameterChangeProposal(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x56ba5e9e:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp PrivKeyEd25519
		tmp, n, err = DecodePrivKeyEd25519(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x08171053:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp PrivKeySecp256k1
		tmp, n, err = DecodePrivKeySecp256k1(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xa5b24c72:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp PubKeyEd25519
		tmp, n, err = DecodePubKeyEd25519(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x7f05210e:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp PubKeyMultisigThreshold
		tmp, n, err = DecodePubKeyMultisigThreshold(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x60f9a133:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp PubKeySecp256k1
		tmp, n, err = DecodePubKeySecp256k1(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x0a456583:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp SdkDec
		tmp, n, err = DecodeSdkDec(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x8918d2bd:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp SdkInt
		tmp, n, err = DecodeSdkInt(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x90d93443:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp SignedMsgType
		tmp, n, err = DecodeSignedMsgType(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x0dad94a2:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp SoftwareUpgradeProposal
		tmp, n, err = DecodeSoftwareUpgradeProposal(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x461d2af7:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp StdSignature
		tmp, n, err = DecodeStdSignature(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x5133aaf7:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp StdTx
		tmp, n, err = DecodeStdTx(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xe96e31e0:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp StoreInfo
		tmp, n, err = DecodeStoreInfo(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xe56842bf:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp Supply
		tmp, n, err = DecodeSupply(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xd0adb3cf:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp TextProposal
		tmp, n, err = DecodeTextProposal(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x8b2f2762:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ValAddress
		tmp, n, err = DecodeValAddress(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x4a608e0b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ValAddressList
		tmp, n, err = DecodeValAddressList(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x02460a0b:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp Validator
		tmp, n, err = DecodeValidator(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x1cfe14b2:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ValidatorCurrentRewards
		tmp, n, err = DecodeValidatorCurrentRewards(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x59012970:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ValidatorHistoricalRewards
		tmp, n, err = DecodeValidatorHistoricalRewards(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x2cfacfb0:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ValidatorSigningInfo
		tmp, n, err = DecodeValidatorSigningInfo(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x76e07d4c:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp ValidatorSlashEvent
		tmp, n, err = DecodeValidatorSlashEvent(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xe6f455cd:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp Vote
		tmp, n, err = DecodeVote(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xac41d0aa:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp VoteOption
		tmp, n, err = DecodeVoteOption(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0x7a9622bc:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp int64
		tmp, n, err = Decodeint64(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	case 0xd5f1d224:
		l := codonDecodeUint64(bz, &n, &err)
		if err != nil {
			return
		}
		bz = bz[n:]
		total += n
		if int(l) > len(bz) {
			err = errors.New("Length Too Large")
			return
		}
		var tmp uint64
		tmp, n, err = Decodeuint64(bz[:l])
		if int(l) != n {
			err = errors.New("Length Mismatch")
			return
		}
		v = tmp
		return
	default:
		panic("Unknown type")
	} // end of switch
	return v, n, nil
} // end of DecodeAny
func AssignIfcPtrFromStruct(ifcPtrIn interface{}, structObjIn interface{}) {
	switch ifcPtr := ifcPtrIn.(type) {
	case *PubKey:
		switch structObj := structObjIn.(type) {
		case PubKeyMultisigThreshold:
			*ifcPtr = &structObj
		case PubKeySecp256k1:
			*ifcPtr = &structObj
		case StdSignature:
			*ifcPtr = &structObj
		case PubKeyEd25519:
			*ifcPtr = &structObj
		default:
			panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
		} // end switch of structs
	case *Content:
		switch structObj := structObjIn.(type) {
		case TextProposal:
			*ifcPtr = &structObj
		case ParameterChangeProposal:
			*ifcPtr = &structObj
		case SoftwareUpgradeProposal:
			*ifcPtr = &structObj
		case CommunityPoolSpendProposal:
			*ifcPtr = &structObj
		default:
			panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
		} // end switch of structs
	case *Tx:
		switch structObj := structObjIn.(type) {
		case StdTx:
			*ifcPtr = &structObj
		default:
			panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
		} // end switch of structs
	case *ModuleAccountI:
		switch structObj := structObjIn.(type) {
		case ModuleAccount:
			*ifcPtr = &structObj
		default:
			panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
		} // end switch of structs
	case *Account:
		switch structObj := structObjIn.(type) {
		case ContinuousVestingAccount:
			*ifcPtr = &structObj
		case ModuleAccount:
			*ifcPtr = &structObj
		case BaseVestingAccount:
			*ifcPtr = &structObj
		case BaseAccount:
			*ifcPtr = &structObj
		case DelayedVestingAccount:
			*ifcPtr = &structObj
		default:
			panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
		} // end switch of structs
	case *PrivKey:
		switch structObj := structObjIn.(type) {
		case PrivKeySecp256k1:
			*ifcPtr = &structObj
		case PrivKeyEd25519:
			*ifcPtr = &structObj
		default:
			panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
		} // end switch of structs
	case *SupplyI:
		switch structObj := structObjIn.(type) {
		case Supply:
			*ifcPtr = &structObj
		default:
			panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
		} // end switch of structs
	case *Msg:
		switch structObj := structObjIn.(type) {
		case MsgWithdrawDelegatorReward:
			*ifcPtr = &structObj
		case MsgVote:
			*ifcPtr = &structObj
		case MsgVerifyInvariant:
			*ifcPtr = &structObj
		case MsgSend:
			*ifcPtr = &structObj
		case MsgUndelegate:
			*ifcPtr = &structObj
		case MsgEditValidator:
			*ifcPtr = &structObj
		case MsgCreateValidator:
			*ifcPtr = &structObj
		case MsgDelegate:
			*ifcPtr = &structObj
		case MsgUnjail:
			*ifcPtr = &structObj
		case MsgBeginRedelegate:
			*ifcPtr = &structObj
		case MsgDeposit:
			*ifcPtr = &structObj
		case MsgWithdrawValidatorCommission:
			*ifcPtr = &structObj
		case MsgMultiSend:
			*ifcPtr = &structObj
		case MsgSetWithdrawAddress:
			*ifcPtr = &structObj
		default:
			panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
		} // end switch of structs
	case *VestingAccount:
		switch structObj := structObjIn.(type) {
		case ContinuousVestingAccount:
			*ifcPtr = &structObj
		case DelayedVestingAccount:
			*ifcPtr = &structObj
		default:
			panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
		} // end switch of structs
	default:
		panic(fmt.Sprintf("Unknown Type %v\n", reflect.TypeOf(ifcPtrIn)))
	} // end switch of interfaces
}
func RandAny(r RandSrc) interface{} {
	switch r.GetUint() % 60 {
	case 0:
		return RandAccAddress(r)
	case 1:
		return RandAccAddressList(r)
	case 2:
		return RandBaseAccount(r)
	case 3:
		return RandBaseVestingAccount(r)
	case 4:
		return RandBondStatus(r)
	case 5:
		return RandCoin(r)
	case 6:
		return RandCommitInfo(r)
	case 7:
		return RandCommunityPoolSpendProposal(r)
	case 8:
		return RandConsAddress(r)
	case 9:
		return RandContinuousVestingAccount(r)
	case 10:
		return RandDecCoin(r)
	case 11:
		return RandDecCoins(r)
	case 12:
		return RandDelayedVestingAccount(r)
	case 13:
		return RandDelegation(r)
	case 14:
		return RandDelegatorStartingInfo(r)
	case 15:
		return RandFeePool(r)
	case 16:
		return RandInput(r)
	case 17:
		return RandModuleAccount(r)
	case 18:
		return RandMsgBeginRedelegate(r)
	case 19:
		return RandMsgCreateValidator(r)
	case 20:
		return RandMsgDelegate(r)
	case 21:
		return RandMsgDeposit(r)
	case 22:
		return RandMsgEditValidator(r)
	case 23:
		return RandMsgMultiSend(r)
	case 24:
		return RandMsgSend(r)
	case 25:
		return RandMsgSetWithdrawAddress(r)
	case 26:
		return RandMsgUndelegate(r)
	case 27:
		return RandMsgUnjail(r)
	case 28:
		return RandMsgVerifyInvariant(r)
	case 29:
		return RandMsgVote(r)
	case 30:
		return RandMsgWithdrawDelegatorReward(r)
	case 31:
		return RandMsgWithdrawValidatorCommission(r)
	case 32:
		return RandOutput(r)
	case 33:
		return RandParamChange(r)
	case 34:
		return RandParameterChangeProposal(r)
	case 35:
		return RandPrivKeyEd25519(r)
	case 36:
		return RandPrivKeySecp256k1(r)
	case 37:
		return RandPubKeyEd25519(r)
	case 38:
		return RandPubKeyMultisigThreshold(r)
	case 39:
		return RandPubKeySecp256k1(r)
	case 40:
		return RandSdkDec(r)
	case 41:
		return RandSdkInt(r)
	case 42:
		return RandSignedMsgType(r)
	case 43:
		return RandSoftwareUpgradeProposal(r)
	case 44:
		return RandStdSignature(r)
	case 45:
		return RandStdTx(r)
	case 46:
		return RandStoreInfo(r)
	case 47:
		return RandSupply(r)
	case 48:
		return RandTextProposal(r)
	case 49:
		return RandValAddress(r)
	case 50:
		return RandValAddressList(r)
	case 51:
		return RandValidator(r)
	case 52:
		return RandValidatorCurrentRewards(r)
	case 53:
		return RandValidatorHistoricalRewards(r)
	case 54:
		return RandValidatorSigningInfo(r)
	case 55:
		return RandValidatorSlashEvent(r)
	case 56:
		return RandVote(r)
	case 57:
		return RandVoteOption(r)
	case 58:
		return Randint64(r)
	case 59:
		return Randuint64(r)
	default:
		panic("Unknown Type.")
	} // end of switch
} // end of func
func DeepCopyAny(x interface{}) interface{} {
	switch v := x.(type) {
	case AccAddress:
		res := DeepCopyAccAddress(v)
		return res
	case *AccAddress:
		res := DeepCopyAccAddress(*v)
		return &res
	case AccAddressList:
		res := DeepCopyAccAddressList(v)
		return res
	case *AccAddressList:
		res := DeepCopyAccAddressList(*v)
		return &res
	case BaseAccount:
		res := DeepCopyBaseAccount(v)
		return res
	case *BaseAccount:
		res := DeepCopyBaseAccount(*v)
		return &res
	case BaseVestingAccount:
		res := DeepCopyBaseVestingAccount(v)
		return res
	case *BaseVestingAccount:
		res := DeepCopyBaseVestingAccount(*v)
		return &res
	case BondStatus:
		res := DeepCopyBondStatus(v)
		return res
	case *BondStatus:
		res := DeepCopyBondStatus(*v)
		return &res
	case Coin:
		res := DeepCopyCoin(v)
		return res
	case *Coin:
		res := DeepCopyCoin(*v)
		return &res
	case CommitInfo:
		res := DeepCopyCommitInfo(v)
		return res
	case *CommitInfo:
		res := DeepCopyCommitInfo(*v)
		return &res
	case CommunityPoolSpendProposal:
		res := DeepCopyCommunityPoolSpendProposal(v)
		return res
	case *CommunityPoolSpendProposal:
		res := DeepCopyCommunityPoolSpendProposal(*v)
		return &res
	case ConsAddress:
		res := DeepCopyConsAddress(v)
		return res
	case *ConsAddress:
		res := DeepCopyConsAddress(*v)
		return &res
	case ContinuousVestingAccount:
		res := DeepCopyContinuousVestingAccount(v)
		return res
	case *ContinuousVestingAccount:
		res := DeepCopyContinuousVestingAccount(*v)
		return &res
	case DecCoin:
		res := DeepCopyDecCoin(v)
		return res
	case *DecCoin:
		res := DeepCopyDecCoin(*v)
		return &res
	case DecCoins:
		res := DeepCopyDecCoins(v)
		return res
	case *DecCoins:
		res := DeepCopyDecCoins(*v)
		return &res
	case DelayedVestingAccount:
		res := DeepCopyDelayedVestingAccount(v)
		return res
	case *DelayedVestingAccount:
		res := DeepCopyDelayedVestingAccount(*v)
		return &res
	case Delegation:
		res := DeepCopyDelegation(v)
		return res
	case *Delegation:
		res := DeepCopyDelegation(*v)
		return &res
	case DelegatorStartingInfo:
		res := DeepCopyDelegatorStartingInfo(v)
		return res
	case *DelegatorStartingInfo:
		res := DeepCopyDelegatorStartingInfo(*v)
		return &res
	case FeePool:
		res := DeepCopyFeePool(v)
		return res
	case *FeePool:
		res := DeepCopyFeePool(*v)
		return &res
	case Input:
		res := DeepCopyInput(v)
		return res
	case *Input:
		res := DeepCopyInput(*v)
		return &res
	case ModuleAccount:
		res := DeepCopyModuleAccount(v)
		return res
	case *ModuleAccount:
		res := DeepCopyModuleAccount(*v)
		return &res
	case MsgBeginRedelegate:
		res := DeepCopyMsgBeginRedelegate(v)
		return res
	case *MsgBeginRedelegate:
		res := DeepCopyMsgBeginRedelegate(*v)
		return &res
	case MsgCreateValidator:
		res := DeepCopyMsgCreateValidator(v)
		return res
	case *MsgCreateValidator:
		res := DeepCopyMsgCreateValidator(*v)
		return &res
	case MsgDelegate:
		res := DeepCopyMsgDelegate(v)
		return res
	case *MsgDelegate:
		res := DeepCopyMsgDelegate(*v)
		return &res
	case MsgDeposit:
		res := DeepCopyMsgDeposit(v)
		return res
	case *MsgDeposit:
		res := DeepCopyMsgDeposit(*v)
		return &res
	case MsgEditValidator:
		res := DeepCopyMsgEditValidator(v)
		return res
	case *MsgEditValidator:
		res := DeepCopyMsgEditValidator(*v)
		return &res
	case MsgMultiSend:
		res := DeepCopyMsgMultiSend(v)
		return res
	case *MsgMultiSend:
		res := DeepCopyMsgMultiSend(*v)
		return &res
	case MsgSend:
		res := DeepCopyMsgSend(v)
		return res
	case *MsgSend:
		res := DeepCopyMsgSend(*v)
		return &res
	case MsgSetWithdrawAddress:
		res := DeepCopyMsgSetWithdrawAddress(v)
		return res
	case *MsgSetWithdrawAddress:
		res := DeepCopyMsgSetWithdrawAddress(*v)
		return &res
	case MsgUndelegate:
		res := DeepCopyMsgUndelegate(v)
		return res
	case *MsgUndelegate:
		res := DeepCopyMsgUndelegate(*v)
		return &res
	case MsgUnjail:
		res := DeepCopyMsgUnjail(v)
		return res
	case *MsgUnjail:
		res := DeepCopyMsgUnjail(*v)
		return &res
	case MsgVerifyInvariant:
		res := DeepCopyMsgVerifyInvariant(v)
		return res
	case *MsgVerifyInvariant:
		res := DeepCopyMsgVerifyInvariant(*v)
		return &res
	case MsgVote:
		res := DeepCopyMsgVote(v)
		return res
	case *MsgVote:
		res := DeepCopyMsgVote(*v)
		return &res
	case MsgWithdrawDelegatorReward:
		res := DeepCopyMsgWithdrawDelegatorReward(v)
		return res
	case *MsgWithdrawDelegatorReward:
		res := DeepCopyMsgWithdrawDelegatorReward(*v)
		return &res
	case MsgWithdrawValidatorCommission:
		res := DeepCopyMsgWithdrawValidatorCommission(v)
		return res
	case *MsgWithdrawValidatorCommission:
		res := DeepCopyMsgWithdrawValidatorCommission(*v)
		return &res
	case Output:
		res := DeepCopyOutput(v)
		return res
	case *Output:
		res := DeepCopyOutput(*v)
		return &res
	case ParamChange:
		res := DeepCopyParamChange(v)
		return res
	case *ParamChange:
		res := DeepCopyParamChange(*v)
		return &res
	case ParameterChangeProposal:
		res := DeepCopyParameterChangeProposal(v)
		return res
	case *ParameterChangeProposal:
		res := DeepCopyParameterChangeProposal(*v)
		return &res
	case PrivKeyEd25519:
		res := DeepCopyPrivKeyEd25519(v)
		return res
	case *PrivKeyEd25519:
		res := DeepCopyPrivKeyEd25519(*v)
		return &res
	case PrivKeySecp256k1:
		res := DeepCopyPrivKeySecp256k1(v)
		return res
	case *PrivKeySecp256k1:
		res := DeepCopyPrivKeySecp256k1(*v)
		return &res
	case PubKeyEd25519:
		res := DeepCopyPubKeyEd25519(v)
		return res
	case *PubKeyEd25519:
		res := DeepCopyPubKeyEd25519(*v)
		return &res
	case PubKeyMultisigThreshold:
		res := DeepCopyPubKeyMultisigThreshold(v)
		return res
	case *PubKeyMultisigThreshold:
		res := DeepCopyPubKeyMultisigThreshold(*v)
		return &res
	case PubKeySecp256k1:
		res := DeepCopyPubKeySecp256k1(v)
		return res
	case *PubKeySecp256k1:
		res := DeepCopyPubKeySecp256k1(*v)
		return &res
	case SdkDec:
		res := DeepCopySdkDec(v)
		return res
	case *SdkDec:
		res := DeepCopySdkDec(*v)
		return &res
	case SdkInt:
		res := DeepCopySdkInt(v)
		return res
	case *SdkInt:
		res := DeepCopySdkInt(*v)
		return &res
	case SignedMsgType:
		res := DeepCopySignedMsgType(v)
		return res
	case *SignedMsgType:
		res := DeepCopySignedMsgType(*v)
		return &res
	case SoftwareUpgradeProposal:
		res := DeepCopySoftwareUpgradeProposal(v)
		return res
	case *SoftwareUpgradeProposal:
		res := DeepCopySoftwareUpgradeProposal(*v)
		return &res
	case StdSignature:
		res := DeepCopyStdSignature(v)
		return res
	case *StdSignature:
		res := DeepCopyStdSignature(*v)
		return &res
	case StdTx:
		res := DeepCopyStdTx(v)
		return res
	case *StdTx:
		res := DeepCopyStdTx(*v)
		return &res
	case StoreInfo:
		res := DeepCopyStoreInfo(v)
		return res
	case *StoreInfo:
		res := DeepCopyStoreInfo(*v)
		return &res
	case Supply:
		res := DeepCopySupply(v)
		return res
	case *Supply:
		res := DeepCopySupply(*v)
		return &res
	case TextProposal:
		res := DeepCopyTextProposal(v)
		return res
	case *TextProposal:
		res := DeepCopyTextProposal(*v)
		return &res
	case ValAddress:
		res := DeepCopyValAddress(v)
		return res
	case *ValAddress:
		res := DeepCopyValAddress(*v)
		return &res
	case ValAddressList:
		res := DeepCopyValAddressList(v)
		return res
	case *ValAddressList:
		res := DeepCopyValAddressList(*v)
		return &res
	case Validator:
		res := DeepCopyValidator(v)
		return res
	case *Validator:
		res := DeepCopyValidator(*v)
		return &res
	case ValidatorCurrentRewards:
		res := DeepCopyValidatorCurrentRewards(v)
		return res
	case *ValidatorCurrentRewards:
		res := DeepCopyValidatorCurrentRewards(*v)
		return &res
	case ValidatorHistoricalRewards:
		res := DeepCopyValidatorHistoricalRewards(v)
		return res
	case *ValidatorHistoricalRewards:
		res := DeepCopyValidatorHistoricalRewards(*v)
		return &res
	case ValidatorSigningInfo:
		res := DeepCopyValidatorSigningInfo(v)
		return res
	case *ValidatorSigningInfo:
		res := DeepCopyValidatorSigningInfo(*v)
		return &res
	case ValidatorSlashEvent:
		res := DeepCopyValidatorSlashEvent(v)
		return res
	case *ValidatorSlashEvent:
		res := DeepCopyValidatorSlashEvent(*v)
		return &res
	case Vote:
		res := DeepCopyVote(v)
		return res
	case *Vote:
		res := DeepCopyVote(*v)
		return &res
	case VoteOption:
		res := DeepCopyVoteOption(v)
		return res
	case *VoteOption:
		res := DeepCopyVoteOption(*v)
		return &res
	case int64:
		res := DeepCopyint64(v)
		return res
	case *int64:
		res := DeepCopyint64(*v)
		return &res
	case uint64:
		res := DeepCopyuint64(v)
		return res
	case *uint64:
		res := DeepCopyuint64(*v)
		return &res
	default:
		panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
	} // end of switch
} // end of func
func GetSupportList() []string {
	return []string{
		".int64",
		".uint64",
		"AccAddressList",
		"ValAddressList",
		"github.com/cosmos/cosmos-sdk/store/rootmulti.commitInfo",
		"github.com/cosmos/cosmos-sdk/store/rootmulti.storeInfo",
		"github.com/cosmos/cosmos-sdk/types.AccAddress",
		"github.com/cosmos/cosmos-sdk/types.BondStatus",
		"github.com/cosmos/cosmos-sdk/types.Coin",
		"github.com/cosmos/cosmos-sdk/types.ConsAddress",
		"github.com/cosmos/cosmos-sdk/types.Dec",
		"github.com/cosmos/cosmos-sdk/types.DecCoin",
		"github.com/cosmos/cosmos-sdk/types.DecCoins",
		"github.com/cosmos/cosmos-sdk/types.Int",
		"github.com/cosmos/cosmos-sdk/types.Msg",
		"github.com/cosmos/cosmos-sdk/types.Tx",
		"github.com/cosmos/cosmos-sdk/types.ValAddress",
		"github.com/cosmos/cosmos-sdk/x/auth/exported.Account",
		"github.com/cosmos/cosmos-sdk/x/auth/exported.VestingAccount",
		"github.com/cosmos/cosmos-sdk/x/auth/types.BaseAccount",
		"github.com/cosmos/cosmos-sdk/x/auth/types.BaseVestingAccount",
		"github.com/cosmos/cosmos-sdk/x/auth/types.ContinuousVestingAccount",
		"github.com/cosmos/cosmos-sdk/x/auth/types.DelayedVestingAccount",
		"github.com/cosmos/cosmos-sdk/x/auth/types.StdSignature",
		"github.com/cosmos/cosmos-sdk/x/auth/types.StdTx",
		"github.com/cosmos/cosmos-sdk/x/bank/internal/types.Input",
		"github.com/cosmos/cosmos-sdk/x/bank/internal/types.MsgMultiSend",
		"github.com/cosmos/cosmos-sdk/x/bank/internal/types.MsgSend",
		"github.com/cosmos/cosmos-sdk/x/bank/internal/types.Output",
		"github.com/cosmos/cosmos-sdk/x/crisis/internal/types.MsgVerifyInvariant",
		"github.com/cosmos/cosmos-sdk/x/distribution/types.CommunityPoolSpendProposal",
		"github.com/cosmos/cosmos-sdk/x/distribution/types.DelegatorStartingInfo",
		"github.com/cosmos/cosmos-sdk/x/distribution/types.FeePool",
		"github.com/cosmos/cosmos-sdk/x/distribution/types.MsgSetWithdrawAddress",
		"github.com/cosmos/cosmos-sdk/x/distribution/types.MsgWithdrawDelegatorReward",
		"github.com/cosmos/cosmos-sdk/x/distribution/types.MsgWithdrawValidatorCommission",
		"github.com/cosmos/cosmos-sdk/x/distribution/types.ValidatorCurrentRewards",
		"github.com/cosmos/cosmos-sdk/x/distribution/types.ValidatorHistoricalRewards",
		"github.com/cosmos/cosmos-sdk/x/distribution/types.ValidatorSlashEvent",
		"github.com/cosmos/cosmos-sdk/x/gov/types.Content",
		"github.com/cosmos/cosmos-sdk/x/gov/types.MsgDeposit",
		"github.com/cosmos/cosmos-sdk/x/gov/types.MsgVote",
		"github.com/cosmos/cosmos-sdk/x/gov/types.SoftwareUpgradeProposal",
		"github.com/cosmos/cosmos-sdk/x/gov/types.TextProposal",
		"github.com/cosmos/cosmos-sdk/x/gov/types.VoteOption",
		"github.com/cosmos/cosmos-sdk/x/params/types.ParamChange",
		"github.com/cosmos/cosmos-sdk/x/params/types.ParameterChangeProposal",
		"github.com/cosmos/cosmos-sdk/x/slashing/types.MsgUnjail",
		"github.com/cosmos/cosmos-sdk/x/slashing/types.ValidatorSigningInfo",
		"github.com/cosmos/cosmos-sdk/x/staking/types.Delegation",
		"github.com/cosmos/cosmos-sdk/x/staking/types.MsgBeginRedelegate",
		"github.com/cosmos/cosmos-sdk/x/staking/types.MsgCreateValidator",
		"github.com/cosmos/cosmos-sdk/x/staking/types.MsgDelegate",
		"github.com/cosmos/cosmos-sdk/x/staking/types.MsgEditValidator",
		"github.com/cosmos/cosmos-sdk/x/staking/types.MsgUndelegate",
		"github.com/cosmos/cosmos-sdk/x/staking/types.Validator",
		"github.com/cosmos/cosmos-sdk/x/supply/exported.ModuleAccountI",
		"github.com/cosmos/cosmos-sdk/x/supply/exported.SupplyI",
		"github.com/cosmos/cosmos-sdk/x/supply/internal/types.ModuleAccount",
		"github.com/cosmos/cosmos-sdk/x/supply/internal/types.Supply",
		"github.com/tendermint/tendermint/crypto.PrivKey",
		"github.com/tendermint/tendermint/crypto.PubKey",
		"github.com/tendermint/tendermint/crypto/ed25519.PrivKeyEd25519",
		"github.com/tendermint/tendermint/crypto/ed25519.PubKeyEd25519",
		"github.com/tendermint/tendermint/crypto/multisig.PubKeyMultisigThreshold",
		"github.com/tendermint/tendermint/crypto/secp256k1.PrivKeySecp256k1",
		"github.com/tendermint/tendermint/crypto/secp256k1.PubKeySecp256k1",
		"github.com/tendermint/tendermint/types.SignedMsgType",
		"github.com/tendermint/tendermint/types.Vote",
	}
} // end of GetSupportList
