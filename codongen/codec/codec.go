//nolint
package codec
import (
"time"
"math/big"
sdk "github.com/cosmos/cosmos-sdk/types"
"github.com/cosmos/cosmos-sdk/codec"
"io"
"fmt"
"reflect"
amino "github.com/coinexchain/codon/wrap-amino"
"encoding/binary"
"math"
"errors"
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

func codonEncodeBool(w *[]byte, v bool) {
	if v {
		*w = append(*w, byte(1))
	} else {
		*w = append(*w, byte(0))
	}
}
func codonEncodeVarint(w *[]byte, v int64) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], v)
	*w = append(*w, buf[0:n]...)
}
func codonEncodeInt8(w *[]byte, v int8) {
	*w = append(*w, byte(v))
}
func codonEncodeInt16(w *[]byte, v int16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], uint16(v))
	*w = append(*w, buf[:]...)
}
func codonEncodeUvarint(w *[]byte, v uint64) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], v)
	*w = append(*w, buf[0:n]...)
}
func codonEncodeUint8(w *[]byte, v uint8) {
	*w = append(*w, byte(v))
}
func codonEncodeUint16(w *[]byte, v uint16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], v)
	*w = append(*w, buf[:]...)
}
func codonEncodeFloat32(w *[]byte, v float32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
	*w = append(*w, buf[:]...)
}
func codonEncodeFloat64(w *[]byte, v float64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(v))
	*w = append(*w, buf[:]...)
}
func codonEncodeByteSlice(w *[]byte, v []byte) {
	codonEncodeVarint(w, int64(len(v)))
	*w = append(*w, v...)
}
func codonEncodeString(w *[]byte, v string) {
	codonEncodeByteSlice(w, []byte(v))
}
func codonDecodeBool(bz []byte, n *int, err *error) bool {
	if len(bz) < 1 {
		*err = errors.New("Not enough bytes to read")
		return false
	}
	*n = 1
	*err = nil
	return bz[0]!=0
}
func codonDecodeInt(bz []byte, m *int, err *error) int {
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
	return int(i)
}
func codonDecodeInt8(bz []byte, n *int, err *error) int8 {
	if len(bz) < 1 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*err = nil
	*n = 1
	return int8(bz[0])
}
func codonDecodeInt16(bz []byte, n *int, err *error) int16 {
	if len(bz) < 2 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*n = 2
	*err = nil
	return int16(binary.LittleEndian.Uint16(bz[:2]))
}
func codonDecodeInt32(bz []byte, n *int, err *error) int32 {
	i := codonDecodeInt64(bz, n, err)
	return int32(i)
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
	i := codonDecodeUint64(bz, n, err)
	return uint(i)
}
func codonDecodeUint8(bz []byte, n *int, err *error) uint8 {
	if len(bz) < 1 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*n = 1
	*err = nil
	return uint8(bz[0])
}
func codonDecodeUint16(bz []byte, n *int, err *error) uint16 {
	if len(bz) < 2 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*n = 2
	*err = nil
	return uint16(binary.LittleEndian.Uint16(bz[:2]))
}
func codonDecodeUint32(bz []byte, n *int, err *error) uint32 {
	i := codonDecodeUint64(bz, n, err)
	return uint32(i)
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
func codonDecodeFloat64(bz []byte, n *int, err *error) float64 {
	if len(bz) < 8 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*n = 8
	*err = nil
	i := binary.LittleEndian.Uint64(bz[:8])
	return math.Float64frombits(i)
}
func codonDecodeFloat32(bz []byte, n *int, err *error) float32 {
	if len(bz) < 4 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*n = 4
	*err = nil
	i := binary.LittleEndian.Uint32(bz[:4])
	return math.Float32frombits(i)
}
func codonGetByteSlice(bz []byte, length int) ([]byte, int, error) {
	if length == 0 {
		return nil, 0, nil
	}
	if len(bz) < length {
		return nil, 0, errors.New("Not enough bytes to read")
	}
	res := make([]byte, length)
	copy(res, bz[:length])
	return res, length, nil
}
func codonDecodeString(bz []byte, n *int, err *error) string {
	var m int
	length := codonDecodeInt64(bz, &m, err)
	if *err != nil {
		return ""
	}
	var bs []byte
	var l int
	bs, l, *err = codonGetByteSlice(bz[m:], int(length))
	*n = m + l
	return string(bs)
}


func init() {
	codec.SetFirstInitFunc(func() {
		amino.Stub = &CodonStub{}
	})
}
func EncodeTime(w *[]byte, t time.Time) {
	t = t.UTC()
	sec := t.Unix()
	var buf [10]byte
	n := binary.PutVarint(buf[:], sec)
	*w = append(*w, buf[0:n]...)

	nanosec := t.Nanosecond()
	n = binary.PutVarint(buf[:], int64(nanosec))
	*w = append(*w, buf[0:n]...)
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

func T(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

var maxSec = T("9999-09-29T08:02:06.647266Z").Unix()

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

func EncodeInt(w *[]byte, v sdk.Int) {
	codonEncodeByteSlice(w, v.BigInt().Bytes())
	codonEncodeBool(w, v.BigInt().Sign() < 0)
}

func DecodeInt(bz []byte) (v sdk.Int, n int, err error) {
	var m int
	length := codonDecodeInt64(bz, &m, &err)
	if err != nil {
		return
	}
	var bs []byte
	var l int
	bs, l, err = codonGetByteSlice(bz[m:], int(length))
	n = m + l
	if err != nil {
		return
	}
	var k int
	isNeg := codonDecodeBool(bz[n:], &k, &err)
	n = n + 1
	if err != nil {
		return
	}
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

func EncodeDec(w *[]byte, v sdk.Dec) {
	codonEncodeByteSlice(w, v.Int.Bytes())
	codonEncodeBool(w, v.Int.Sign() < 0)
}

func DecodeDec(bz []byte) (v sdk.Dec, n int, err error) {
	var m int
	length := codonDecodeInt64(bz, &m, &err)
	if err != nil {
		return
	}
	var bs []byte
	var l int
	bs, l, err = codonGetByteSlice(bz[m:], int(length))
	n = m + l
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


// ========= BridgeBegin ============
type CodecImp struct {
	sealed          bool
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
	if err!=nil {
		panic(err)
	}
	return bz
}
func (cdc *CodecImp) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryLengthPrefixed(o)
	if err!=nil {
		panic(err)
	}
	return bz
}
func (cdc *CodecImp) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryBare(bz, ptr)
	if err!=nil {
		panic(err)
	}
}
func (cdc *CodecImp) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err!=nil {
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
	if _, err := getMagicBytesOfVar(o); err!=nil {
		return nil, err
	}
	buf := make([]byte, 0, 1024)
	EncodeAny(&buf, o)
	return buf, nil
}
func (s *CodonStub) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	if _, err := getMagicBytesOfVar(o); err!=nil {
		return nil, err
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
	if rv.Elem().Kind() != reflect.Interface {
		magicBytes, err := getMagicBytesOfVar(ptr)
		if err!=nil {
			return err
		}
		if bz[0]!=magicBytes[0] || bz[1]!=magicBytes[1] || bz[2]!=magicBytes[2] || bz[3]!=magicBytes[3] {
			return fmt.Errorf("MagicBytes Missmatch %v vs %v", bz[0:4], magicBytes[:])
		}
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
	if err!=nil {
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
	buf := make([]byte, 0, binary.MaxVarintLen64+len(bz))
	codonEncodeByteSlice(&buf, bz)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeUvarint(w io.Writer, u uint64) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeUvarint(&buf, u)
	_, err := w.Write(buf)
	return err
}
func (s *CodonStub) ByteSliceSize(bz []byte) int {
	return s.UvarintSize(uint64(len(bz))) + len(bz)
}
func (_ *CodonStub) EncodeInt8(w io.Writer, i int8) error {
	_, err := w.Write([]byte{byte(i)})
	return err
}
func (_ *CodonStub) EncodeInt16(w io.Writer, i int16) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeInt16(&buf, i)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeInt32(w io.Writer, i int32) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeVarint(&buf, int64(i))
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeInt64(w io.Writer, i int64) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeVarint(&buf, i)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeVarint(w io.Writer, i int64) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeVarint(&buf, i)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeByte(w io.Writer, b byte) error {
	_, err := w.Write([]byte{b})
	return err
}
func (_ *CodonStub) EncodeUint8(w io.Writer, u uint8) error {
	_, err := w.Write([]byte{u})
	return err
}
func (_ *CodonStub) EncodeUint16(w io.Writer, u uint16) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeUint16(&buf, u)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeUint32(w io.Writer, u uint32) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeUvarint(&buf, uint64(u))
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeUint64(w io.Writer, u uint64) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeUvarint(&buf, u)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeBool(w io.Writer, b bool) error {
	u := byte(0)
	if b {
		u = byte(1)
	}
	_, err := w.Write([]byte{u})
	return err
}
func (_ *CodonStub) EncodeFloat32(w io.Writer, f float32) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeFloat32(&buf, f)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeFloat64(w io.Writer, f float64) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeFloat64(&buf, f)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeString(w io.Writer, s string) error {
	buf := make([]byte, 0, binary.MaxVarintLen64+len(s))
	codonEncodeString(&buf, s)
	_, err := w.Write(buf)
	return err
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
func (_ *CodonStub) DecodeFloat32(bz []byte) (f float32, n int, err error) {
	f = codonDecodeFloat32(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeFloat64(bz []byte) (f float64, n int, err error) {
	f = codonDecodeFloat64(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error) {
	length := codonDecodeInt(bz, &n, &err)
	if err != nil {
		return
	}
	bz = bz[n:]
	n += length
	bz2, m, err := codonGetByteSlice(bz, length)
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
codonEncodeByteSlice(w, v[:])
} //End of EncodePrivKeyEd25519

func DecodePrivKeyEd25519(bz []byte) (PrivKeyEd25519, int, error) {
var err error
var length int
var v PrivKeyEd25519
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
v[_0] = uint8(codonDecodeUint8(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodePrivKeyEd25519

func RandPrivKeyEd25519(r RandSrc) PrivKeyEd25519 {
var length int
var v PrivKeyEd25519
length = 64
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
v[_0] = r.GetUint8()
}
return v
} //End of RandPrivKeyEd25519

func DeepCopyPrivKeyEd25519(in PrivKeyEd25519) (out PrivKeyEd25519) {
var length int
length = len(in)
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
out[_0] = in[_0]
}
return
} //End of DeepCopyPrivKeyEd25519

// Non-Interface
func EncodePrivKeySecp256k1(w *[]byte, v PrivKeySecp256k1) {
codonEncodeByteSlice(w, v[:])
} //End of EncodePrivKeySecp256k1

func DecodePrivKeySecp256k1(bz []byte) (PrivKeySecp256k1, int, error) {
var err error
var length int
var v PrivKeySecp256k1
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
v[_0] = uint8(codonDecodeUint8(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodePrivKeySecp256k1

func RandPrivKeySecp256k1(r RandSrc) PrivKeySecp256k1 {
var length int
var v PrivKeySecp256k1
length = 32
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
v[_0] = r.GetUint8()
}
return v
} //End of RandPrivKeySecp256k1

func DeepCopyPrivKeySecp256k1(in PrivKeySecp256k1) (out PrivKeySecp256k1) {
var length int
length = len(in)
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
out[_0] = in[_0]
}
return
} //End of DeepCopyPrivKeySecp256k1

// Non-Interface
func EncodePubKeyEd25519(w *[]byte, v PubKeyEd25519) {
codonEncodeByteSlice(w, v[:])
} //End of EncodePubKeyEd25519

func DecodePubKeyEd25519(bz []byte) (PubKeyEd25519, int, error) {
var err error
var length int
var v PubKeyEd25519
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
v[_0] = uint8(codonDecodeUint8(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodePubKeyEd25519

func RandPubKeyEd25519(r RandSrc) PubKeyEd25519 {
var length int
var v PubKeyEd25519
length = 32
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
v[_0] = r.GetUint8()
}
return v
} //End of RandPubKeyEd25519

func DeepCopyPubKeyEd25519(in PubKeyEd25519) (out PubKeyEd25519) {
var length int
length = len(in)
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
out[_0] = in[_0]
}
return
} //End of DeepCopyPubKeyEd25519

// Non-Interface
func EncodePubKeySecp256k1(w *[]byte, v PubKeySecp256k1) {
codonEncodeByteSlice(w, v[:])
} //End of EncodePubKeySecp256k1

func DecodePubKeySecp256k1(bz []byte) (PubKeySecp256k1, int, error) {
var err error
var length int
var v PubKeySecp256k1
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
v[_0] = uint8(codonDecodeUint8(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodePubKeySecp256k1

func RandPubKeySecp256k1(r RandSrc) PubKeySecp256k1 {
var length int
var v PubKeySecp256k1
length = 33
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
v[_0] = r.GetUint8()
}
return v
} //End of RandPubKeySecp256k1

func DeepCopyPubKeySecp256k1(in PubKeySecp256k1) (out PubKeySecp256k1) {
var length int
length = len(in)
for _0, length_0 := 0, length; _0<length_0; _0++ { //array of uint8
out[_0] = in[_0]
}
return
} //End of DeepCopyPubKeySecp256k1

// Non-Interface
func EncodePubKeyMultisigThreshold(w *[]byte, v PubKeyMultisigThreshold) {
codonEncodeUvarint(w, uint64(v.K))
codonEncodeVarint(w, int64(len(v.PubKeys)))
for _0:=0; _0<len(v.PubKeys); _0++ {
EncodePubKey(w, v.PubKeys[_0])// interface_encode
}
} //End of EncodePubKeyMultisigThreshold

func DecodePubKeyMultisigThreshold(bz []byte) (PubKeyMultisigThreshold, int, error) {
var err error
var length int
var v PubKeyMultisigThreshold
var n int
var total int
v.K = uint(codonDecodeUint(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.PubKeys = nil
} else {
v.PubKeys = make([]PubKey, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of interface
v.PubKeys[_0], n, err = DecodePubKey(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodePubKeyMultisigThreshold

func RandPubKeyMultisigThreshold(r RandSrc) PubKeyMultisigThreshold {
var length int
var v PubKeyMultisigThreshold
v.K = r.GetUint()
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.PubKeys = nil
} else {
v.PubKeys = make([]PubKey, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of interface
v.PubKeys[_0] = RandPubKey(r)
}
return v
} //End of RandPubKeyMultisigThreshold

func DeepCopyPubKeyMultisigThreshold(in PubKeyMultisigThreshold) (out PubKeyMultisigThreshold) {
var length int
out.K = in.K
length = len(in.PubKeys)
if length==0 {out.PubKeys = nil
} else {
out.PubKeys = make([]PubKey, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of interface
out.PubKeys[_0] = DeepCopyPubKey(in.PubKeys[_0])
}
return
} //End of DeepCopyPubKeyMultisigThreshold

// Non-Interface
func EncodeSignedMsgType(w *[]byte, v SignedMsgType) {
codonEncodeUint8(w, uint8(v))
} //End of EncodeSignedMsgType

func DecodeSignedMsgType(bz []byte) (SignedMsgType, int, error) {
var err error
var v SignedMsgType
var n int
var total int
v = SignedMsgType(codonDecodeUint8(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
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
codonEncodeUint8(w, uint8(v))
} //End of EncodeVoteOption

func DecodeVoteOption(bz []byte) (VoteOption, int, error) {
var err error
var v VoteOption
var n int
var total int
v = VoteOption(codonDecodeUint8(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
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
codonEncodeUint8(w, uint8(v.Type))
codonEncodeVarint(w, int64(v.Height))
codonEncodeVarint(w, int64(v.Round))
codonEncodeByteSlice(w, v.BlockID.Hash[:])
codonEncodeVarint(w, int64(v.BlockID.PartsHeader.Total))
codonEncodeByteSlice(w, v.BlockID.PartsHeader.Hash[:])
// end of v.BlockID.PartsHeader
// end of v.BlockID
EncodeTime(w, v.Timestamp)
codonEncodeByteSlice(w, v.ValidatorAddress[:])
codonEncodeVarint(w, int64(v.ValidatorIndex))
codonEncodeByteSlice(w, v.Signature[:])
} //End of EncodeVote

func DecodeVote(bz []byte) (Vote, int, error) {
var err error
var length int
var v Vote
var n int
var total int
v.Type = SignedMsgType(codonDecodeUint8(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Height = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Round = int(codonDecodeInt(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BlockID.Hash, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BlockID.PartsHeader.Total = int(codonDecodeInt(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BlockID.PartsHeader.Hash, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.BlockID.PartsHeader
// end of v.BlockID
v.Timestamp, n, err = DecodeTime(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorIndex = int(codonDecodeInt(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Signature, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeVote

func RandVote(r RandSrc) Vote {
var length int
var v Vote
v.Type = SignedMsgType(r.GetUint8())
v.Height = r.GetInt64()
v.Round = r.GetInt()
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.BlockID.Hash = r.GetBytes(length)
v.BlockID.PartsHeader.Total = r.GetInt()
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.BlockID.PartsHeader.Hash = r.GetBytes(length)
// end of v.BlockID.PartsHeader
// end of v.BlockID
v.Timestamp = RandTime(r)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ValidatorAddress = r.GetBytes(length)
v.ValidatorIndex = r.GetInt()
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.Signature = r.GetBytes(length)
return v
} //End of RandVote

func DeepCopyVote(in Vote) (out Vote) {
var length int
out.Type = in.Type
out.Height = in.Height
out.Round = in.Round
length = len(in.BlockID.Hash)
if length==0 {out.BlockID.Hash = nil
} else {
out.BlockID.Hash = make([]uint8, length)
}
copy(out.BlockID.Hash[:], in.BlockID.Hash[:])
out.BlockID.PartsHeader.Total = in.BlockID.PartsHeader.Total
length = len(in.BlockID.PartsHeader.Hash)
if length==0 {out.BlockID.PartsHeader.Hash = nil
} else {
out.BlockID.PartsHeader.Hash = make([]uint8, length)
}
copy(out.BlockID.PartsHeader.Hash[:], in.BlockID.PartsHeader.Hash[:])
// end of .BlockID.PartsHeader
// end of .BlockID
out.Timestamp = DeepCopyTime(in.Timestamp)
length = len(in.ValidatorAddress)
if length==0 {out.ValidatorAddress = nil
} else {
out.ValidatorAddress = make([]uint8, length)
}
copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
out.ValidatorIndex = in.ValidatorIndex
length = len(in.Signature)
if length==0 {out.Signature = nil
} else {
out.Signature = make([]uint8, length)
}
copy(out.Signature[:], in.Signature[:])
return
} //End of DeepCopyVote

// Non-Interface
func EncodeSdkInt(w *[]byte, v SdkInt) {
EncodeInt(w, v)
} //End of EncodeSdkInt

func DecodeSdkInt(bz []byte) (SdkInt, int, error) {
var err error
var v SdkInt
var n int
var total int
v, n, err = DecodeInt(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
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
EncodeDec(w, v)
} //End of EncodeSdkDec

func DecodeSdkDec(bz []byte) (SdkDec, int, error) {
var err error
var v SdkDec
var n int
var total int
v, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
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
codonEncodeUvarint(w, uint64(v))
} //End of Encodeuint64

func Decodeuint64(bz []byte) (uint64, int, error) {
var err error
var v uint64
var n int
var total int
v = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
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
codonEncodeVarint(w, int64(v))
} //End of Encodeint64

func Decodeint64(bz []byte) (int64, int, error) {
var err error
var v int64
var n int
var total int
v = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
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
codonEncodeByteSlice(w, v[:])
} //End of EncodeConsAddress

func DecodeConsAddress(bz []byte) (ConsAddress, int, error) {
var err error
var length int
var v ConsAddress
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeConsAddress

func RandConsAddress(r RandSrc) ConsAddress {
var length int
var v ConsAddress
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v = r.GetBytes(length)
return v
} //End of RandConsAddress

func DeepCopyConsAddress(in ConsAddress) (out ConsAddress) {
var length int
length = len(in)
if length==0 {out = nil
} else {
out = make([]uint8, length)
}
copy(out[:], in[:])
return
} //End of DeepCopyConsAddress

// Non-Interface
func EncodeCoin(w *[]byte, v Coin) {
codonEncodeString(w, v.Denom)
EncodeInt(w, v.Amount)
} //End of EncodeCoin

func DecodeCoin(bz []byte) (Coin, int, error) {
var err error
var v Coin
var n int
var total int
v.Denom = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Amount, n, err = DecodeInt(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeCoin

func RandCoin(r RandSrc) Coin {
var v Coin
v.Denom = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
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
codonEncodeString(w, v.Denom)
EncodeDec(w, v.Amount)
} //End of EncodeDecCoin

func DecodeDecCoin(bz []byte) (DecCoin, int, error) {
var err error
var v DecCoin
var n int
var total int
v.Denom = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Amount, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeDecCoin

func RandDecCoin(r RandSrc) DecCoin {
var v DecCoin
v.Denom = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
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
EncodePubKey(w, v.PubKey)// interface_encode
codonEncodeByteSlice(w, v.Signature[:])
} //End of EncodeStdSignature

func DecodeStdSignature(bz []byte) (StdSignature, int, error) {
var err error
var length int
var v StdSignature
var n int
var total int
v.PubKey, n, err = DecodePubKey(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n // interface_decode
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Signature, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeStdSignature

func RandStdSignature(r RandSrc) StdSignature {
var length int
var v StdSignature
v.PubKey = RandPubKey(r) // interface_decode
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.Signature = r.GetBytes(length)
return v
} //End of RandStdSignature

func DeepCopyStdSignature(in StdSignature) (out StdSignature) {
var length int
out.PubKey = DeepCopyPubKey(in.PubKey)
length = len(in.Signature)
if length==0 {out.Signature = nil
} else {
out.Signature = make([]uint8, length)
}
copy(out.Signature[:], in.Signature[:])
return
} //End of DeepCopyStdSignature

// Non-Interface
func EncodeParamChange(w *[]byte, v ParamChange) {
codonEncodeString(w, v.Subspace)
codonEncodeString(w, v.Key)
codonEncodeString(w, v.Subkey)
codonEncodeString(w, v.Value)
} //End of EncodeParamChange

func DecodeParamChange(bz []byte) (ParamChange, int, error) {
var err error
var v ParamChange
var n int
var total int
v.Subspace = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Key = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Subkey = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Value = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeParamChange

func RandParamChange(r RandSrc) ParamChange {
var v ParamChange
v.Subspace = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Key = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Subkey = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Value = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
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
codonEncodeByteSlice(w, v.Address[:])
codonEncodeVarint(w, int64(len(v.Coins)))
for _0:=0; _0<len(v.Coins); _0++ {
codonEncodeString(w, v.Coins[_0].Denom)
EncodeInt(w, v.Coins[_0].Amount)
// end of v.Coins[_0]
}
} //End of EncodeInput

func DecodeInput(bz []byte) (Input, int, error) {
var err error
var length int
var v Input
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Address, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Coins = nil
} else {
v.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Coins[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeInput

func RandInput(r RandSrc) Input {
var length int
var v Input
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.Address = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Coins = nil
} else {
v.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Coins[_0] = RandCoin(r)
}
return v
} //End of RandInput

func DeepCopyInput(in Input) (out Input) {
var length int
length = len(in.Address)
if length==0 {out.Address = nil
} else {
out.Address = make([]uint8, length)
}
copy(out.Address[:], in.Address[:])
length = len(in.Coins)
if length==0 {out.Coins = nil
} else {
out.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Coins[_0] = DeepCopyCoin(in.Coins[_0])
}
return
} //End of DeepCopyInput

// Non-Interface
func EncodeOutput(w *[]byte, v Output) {
codonEncodeByteSlice(w, v.Address[:])
codonEncodeVarint(w, int64(len(v.Coins)))
for _0:=0; _0<len(v.Coins); _0++ {
codonEncodeString(w, v.Coins[_0].Denom)
EncodeInt(w, v.Coins[_0].Amount)
// end of v.Coins[_0]
}
} //End of EncodeOutput

func DecodeOutput(bz []byte) (Output, int, error) {
var err error
var length int
var v Output
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Address, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Coins = nil
} else {
v.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Coins[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeOutput

func RandOutput(r RandSrc) Output {
var length int
var v Output
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.Address = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Coins = nil
} else {
v.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Coins[_0] = RandCoin(r)
}
return v
} //End of RandOutput

func DeepCopyOutput(in Output) (out Output) {
var length int
length = len(in.Address)
if length==0 {out.Address = nil
} else {
out.Address = make([]uint8, length)
}
copy(out.Address[:], in.Address[:])
length = len(in.Coins)
if length==0 {out.Coins = nil
} else {
out.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Coins[_0] = DeepCopyCoin(in.Coins[_0])
}
return
} //End of DeepCopyOutput

// Non-Interface
func EncodeAccAddress(w *[]byte, v AccAddress) {
codonEncodeByteSlice(w, v[:])
} //End of EncodeAccAddress

func DecodeAccAddress(bz []byte) (AccAddress, int, error) {
var err error
var length int
var v AccAddress
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeAccAddress

func RandAccAddress(r RandSrc) AccAddress {
var length int
var v AccAddress
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v = r.GetBytes(length)
return v
} //End of RandAccAddress

func DeepCopyAccAddress(in AccAddress) (out AccAddress) {
var length int
length = len(in)
if length==0 {out = nil
} else {
out = make([]uint8, length)
}
copy(out[:], in[:])
return
} //End of DeepCopyAccAddress

// Non-Interface
func EncodeBaseAccount(w *[]byte, v BaseAccount) {
codonEncodeByteSlice(w, v.Address[:])
codonEncodeVarint(w, int64(len(v.Coins)))
for _0:=0; _0<len(v.Coins); _0++ {
codonEncodeString(w, v.Coins[_0].Denom)
EncodeInt(w, v.Coins[_0].Amount)
// end of v.Coins[_0]
}
EncodePubKey(w, v.PubKey)// interface_encode
codonEncodeUvarint(w, uint64(v.AccountNumber))
codonEncodeUvarint(w, uint64(v.Sequence))
} //End of EncodeBaseAccount

func DecodeBaseAccount(bz []byte) (BaseAccount, int, error) {
var err error
var length int
var v BaseAccount
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Address, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Coins = nil
} else {
v.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Coins[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.PubKey, n, err = DecodePubKey(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n // interface_decode
v.AccountNumber = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Sequence = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeBaseAccount

func RandBaseAccount(r RandSrc) BaseAccount {
var length int
var v BaseAccount
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.Address = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Coins = nil
} else {
v.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
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
if length==0 {out.Address = nil
} else {
out.Address = make([]uint8, length)
}
copy(out.Address[:], in.Address[:])
length = len(in.Coins)
if length==0 {out.Coins = nil
} else {
out.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Coins[_0] = DeepCopyCoin(in.Coins[_0])
}
out.PubKey = DeepCopyPubKey(in.PubKey)
out.AccountNumber = in.AccountNumber
out.Sequence = in.Sequence
return
} //End of DeepCopyBaseAccount

// Non-Interface
func EncodeBaseVestingAccount(w *[]byte, v BaseVestingAccount) {
codonEncodeByteSlice(w, v.BaseAccount.Address[:])
codonEncodeVarint(w, int64(len(v.BaseAccount.Coins)))
for _0:=0; _0<len(v.BaseAccount.Coins); _0++ {
codonEncodeString(w, v.BaseAccount.Coins[_0].Denom)
EncodeInt(w, v.BaseAccount.Coins[_0].Amount)
// end of v.BaseAccount.Coins[_0]
}
EncodePubKey(w, v.BaseAccount.PubKey)// interface_encode
codonEncodeUvarint(w, uint64(v.BaseAccount.AccountNumber))
codonEncodeUvarint(w, uint64(v.BaseAccount.Sequence))
// end of v.BaseAccount
codonEncodeVarint(w, int64(len(v.OriginalVesting)))
for _0:=0; _0<len(v.OriginalVesting); _0++ {
codonEncodeString(w, v.OriginalVesting[_0].Denom)
EncodeInt(w, v.OriginalVesting[_0].Amount)
// end of v.OriginalVesting[_0]
}
codonEncodeVarint(w, int64(len(v.DelegatedFree)))
for _0:=0; _0<len(v.DelegatedFree); _0++ {
codonEncodeString(w, v.DelegatedFree[_0].Denom)
EncodeInt(w, v.DelegatedFree[_0].Amount)
// end of v.DelegatedFree[_0]
}
codonEncodeVarint(w, int64(len(v.DelegatedVesting)))
for _0:=0; _0<len(v.DelegatedVesting); _0++ {
codonEncodeString(w, v.DelegatedVesting[_0].Denom)
EncodeInt(w, v.DelegatedVesting[_0].Amount)
// end of v.DelegatedVesting[_0]
}
codonEncodeVarint(w, int64(v.EndTime))
} //End of EncodeBaseVestingAccount

func DecodeBaseVestingAccount(bz []byte) (BaseVestingAccount, int, error) {
var err error
var length int
var v BaseVestingAccount
var n int
var total int
v.BaseAccount = &BaseAccount{}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BaseAccount.Address, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.BaseAccount.Coins = nil
} else {
v.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseAccount.Coins[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.BaseAccount.PubKey, n, err = DecodePubKey(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n // interface_decode
v.BaseAccount.AccountNumber = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BaseAccount.Sequence = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.BaseAccount
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.OriginalVesting = nil
} else {
v.OriginalVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.OriginalVesting[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.DelegatedFree = nil
} else {
v.DelegatedFree = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.DelegatedFree[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.DelegatedVesting = nil
} else {
v.DelegatedVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.DelegatedVesting[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.EndTime = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeBaseVestingAccount

func RandBaseVestingAccount(r RandSrc) BaseVestingAccount {
var length int
var v BaseVestingAccount
v.BaseAccount = &BaseAccount{}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.BaseAccount.Address = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.BaseAccount.Coins = nil
} else {
v.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseAccount.Coins[_0] = RandCoin(r)
}
v.BaseAccount.PubKey = RandPubKey(r) // interface_decode
v.BaseAccount.AccountNumber = r.GetUint64()
v.BaseAccount.Sequence = r.GetUint64()
// end of v.BaseAccount
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.OriginalVesting = nil
} else {
v.OriginalVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.OriginalVesting[_0] = RandCoin(r)
}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.DelegatedFree = nil
} else {
v.DelegatedFree = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.DelegatedFree[_0] = RandCoin(r)
}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.DelegatedVesting = nil
} else {
v.DelegatedVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.DelegatedVesting[_0] = RandCoin(r)
}
v.EndTime = r.GetInt64()
return v
} //End of RandBaseVestingAccount

func DeepCopyBaseVestingAccount(in BaseVestingAccount) (out BaseVestingAccount) {
var length int
out.BaseAccount = &BaseAccount{}
length = len(in.BaseAccount.Address)
if length==0 {out.BaseAccount.Address = nil
} else {
out.BaseAccount.Address = make([]uint8, length)
}
copy(out.BaseAccount.Address[:], in.BaseAccount.Address[:])
length = len(in.BaseAccount.Coins)
if length==0 {out.BaseAccount.Coins = nil
} else {
out.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.BaseAccount.Coins[_0] = DeepCopyCoin(in.BaseAccount.Coins[_0])
}
out.BaseAccount.PubKey = DeepCopyPubKey(in.BaseAccount.PubKey)
out.BaseAccount.AccountNumber = in.BaseAccount.AccountNumber
out.BaseAccount.Sequence = in.BaseAccount.Sequence
// end of .BaseAccount
length = len(in.OriginalVesting)
if length==0 {out.OriginalVesting = nil
} else {
out.OriginalVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.OriginalVesting[_0] = DeepCopyCoin(in.OriginalVesting[_0])
}
length = len(in.DelegatedFree)
if length==0 {out.DelegatedFree = nil
} else {
out.DelegatedFree = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.DelegatedFree[_0] = DeepCopyCoin(in.DelegatedFree[_0])
}
length = len(in.DelegatedVesting)
if length==0 {out.DelegatedVesting = nil
} else {
out.DelegatedVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.DelegatedVesting[_0] = DeepCopyCoin(in.DelegatedVesting[_0])
}
out.EndTime = in.EndTime
return
} //End of DeepCopyBaseVestingAccount

// Non-Interface
func EncodeContinuousVestingAccount(w *[]byte, v ContinuousVestingAccount) {
codonEncodeByteSlice(w, v.BaseVestingAccount.BaseAccount.Address[:])
codonEncodeVarint(w, int64(len(v.BaseVestingAccount.BaseAccount.Coins)))
for _0:=0; _0<len(v.BaseVestingAccount.BaseAccount.Coins); _0++ {
codonEncodeString(w, v.BaseVestingAccount.BaseAccount.Coins[_0].Denom)
EncodeInt(w, v.BaseVestingAccount.BaseAccount.Coins[_0].Amount)
// end of v.BaseVestingAccount.BaseAccount.Coins[_0]
}
EncodePubKey(w, v.BaseVestingAccount.BaseAccount.PubKey)// interface_encode
codonEncodeUvarint(w, uint64(v.BaseVestingAccount.BaseAccount.AccountNumber))
codonEncodeUvarint(w, uint64(v.BaseVestingAccount.BaseAccount.Sequence))
// end of v.BaseVestingAccount.BaseAccount
codonEncodeVarint(w, int64(len(v.BaseVestingAccount.OriginalVesting)))
for _0:=0; _0<len(v.BaseVestingAccount.OriginalVesting); _0++ {
codonEncodeString(w, v.BaseVestingAccount.OriginalVesting[_0].Denom)
EncodeInt(w, v.BaseVestingAccount.OriginalVesting[_0].Amount)
// end of v.BaseVestingAccount.OriginalVesting[_0]
}
codonEncodeVarint(w, int64(len(v.BaseVestingAccount.DelegatedFree)))
for _0:=0; _0<len(v.BaseVestingAccount.DelegatedFree); _0++ {
codonEncodeString(w, v.BaseVestingAccount.DelegatedFree[_0].Denom)
EncodeInt(w, v.BaseVestingAccount.DelegatedFree[_0].Amount)
// end of v.BaseVestingAccount.DelegatedFree[_0]
}
codonEncodeVarint(w, int64(len(v.BaseVestingAccount.DelegatedVesting)))
for _0:=0; _0<len(v.BaseVestingAccount.DelegatedVesting); _0++ {
codonEncodeString(w, v.BaseVestingAccount.DelegatedVesting[_0].Denom)
EncodeInt(w, v.BaseVestingAccount.DelegatedVesting[_0].Amount)
// end of v.BaseVestingAccount.DelegatedVesting[_0]
}
codonEncodeVarint(w, int64(v.BaseVestingAccount.EndTime))
// end of v.BaseVestingAccount
codonEncodeVarint(w, int64(v.StartTime))
} //End of EncodeContinuousVestingAccount

func DecodeContinuousVestingAccount(bz []byte) (ContinuousVestingAccount, int, error) {
var err error
var length int
var v ContinuousVestingAccount
var n int
var total int
v.BaseVestingAccount = &BaseVestingAccount{}
v.BaseVestingAccount.BaseAccount = &BaseAccount{}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BaseVestingAccount.BaseAccount.Address, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.BaseVestingAccount.BaseAccount.Coins = nil
} else {
v.BaseVestingAccount.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.BaseAccount.Coins[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.BaseVestingAccount.BaseAccount.PubKey, n, err = DecodePubKey(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n // interface_decode
v.BaseVestingAccount.BaseAccount.AccountNumber = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BaseVestingAccount.BaseAccount.Sequence = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.BaseVestingAccount.BaseAccount
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.BaseVestingAccount.OriginalVesting = nil
} else {
v.BaseVestingAccount.OriginalVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.OriginalVesting[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.BaseVestingAccount.DelegatedFree = nil
} else {
v.BaseVestingAccount.DelegatedFree = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.DelegatedFree[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.BaseVestingAccount.DelegatedVesting = nil
} else {
v.BaseVestingAccount.DelegatedVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.DelegatedVesting[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.BaseVestingAccount.EndTime = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.BaseVestingAccount
v.StartTime = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeContinuousVestingAccount

func RandContinuousVestingAccount(r RandSrc) ContinuousVestingAccount {
var length int
var v ContinuousVestingAccount
v.BaseVestingAccount = &BaseVestingAccount{}
v.BaseVestingAccount.BaseAccount = &BaseAccount{}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.BaseVestingAccount.BaseAccount.Address = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.BaseVestingAccount.BaseAccount.Coins = nil
} else {
v.BaseVestingAccount.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.BaseAccount.Coins[_0] = RandCoin(r)
}
v.BaseVestingAccount.BaseAccount.PubKey = RandPubKey(r) // interface_decode
v.BaseVestingAccount.BaseAccount.AccountNumber = r.GetUint64()
v.BaseVestingAccount.BaseAccount.Sequence = r.GetUint64()
// end of v.BaseVestingAccount.BaseAccount
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.BaseVestingAccount.OriginalVesting = nil
} else {
v.BaseVestingAccount.OriginalVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.OriginalVesting[_0] = RandCoin(r)
}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.BaseVestingAccount.DelegatedFree = nil
} else {
v.BaseVestingAccount.DelegatedFree = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.DelegatedFree[_0] = RandCoin(r)
}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.BaseVestingAccount.DelegatedVesting = nil
} else {
v.BaseVestingAccount.DelegatedVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
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
if length==0 {out.BaseVestingAccount.BaseAccount.Address = nil
} else {
out.BaseVestingAccount.BaseAccount.Address = make([]uint8, length)
}
copy(out.BaseVestingAccount.BaseAccount.Address[:], in.BaseVestingAccount.BaseAccount.Address[:])
length = len(in.BaseVestingAccount.BaseAccount.Coins)
if length==0 {out.BaseVestingAccount.BaseAccount.Coins = nil
} else {
out.BaseVestingAccount.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.BaseVestingAccount.BaseAccount.Coins[_0] = DeepCopyCoin(in.BaseVestingAccount.BaseAccount.Coins[_0])
}
out.BaseVestingAccount.BaseAccount.PubKey = DeepCopyPubKey(in.BaseVestingAccount.BaseAccount.PubKey)
out.BaseVestingAccount.BaseAccount.AccountNumber = in.BaseVestingAccount.BaseAccount.AccountNumber
out.BaseVestingAccount.BaseAccount.Sequence = in.BaseVestingAccount.BaseAccount.Sequence
// end of .BaseVestingAccount.BaseAccount
length = len(in.BaseVestingAccount.OriginalVesting)
if length==0 {out.BaseVestingAccount.OriginalVesting = nil
} else {
out.BaseVestingAccount.OriginalVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.BaseVestingAccount.OriginalVesting[_0] = DeepCopyCoin(in.BaseVestingAccount.OriginalVesting[_0])
}
length = len(in.BaseVestingAccount.DelegatedFree)
if length==0 {out.BaseVestingAccount.DelegatedFree = nil
} else {
out.BaseVestingAccount.DelegatedFree = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.BaseVestingAccount.DelegatedFree[_0] = DeepCopyCoin(in.BaseVestingAccount.DelegatedFree[_0])
}
length = len(in.BaseVestingAccount.DelegatedVesting)
if length==0 {out.BaseVestingAccount.DelegatedVesting = nil
} else {
out.BaseVestingAccount.DelegatedVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.BaseVestingAccount.DelegatedVesting[_0] = DeepCopyCoin(in.BaseVestingAccount.DelegatedVesting[_0])
}
out.BaseVestingAccount.EndTime = in.BaseVestingAccount.EndTime
// end of .BaseVestingAccount
out.StartTime = in.StartTime
return
} //End of DeepCopyContinuousVestingAccount

// Non-Interface
func EncodeDelayedVestingAccount(w *[]byte, v DelayedVestingAccount) {
codonEncodeByteSlice(w, v.BaseVestingAccount.BaseAccount.Address[:])
codonEncodeVarint(w, int64(len(v.BaseVestingAccount.BaseAccount.Coins)))
for _0:=0; _0<len(v.BaseVestingAccount.BaseAccount.Coins); _0++ {
codonEncodeString(w, v.BaseVestingAccount.BaseAccount.Coins[_0].Denom)
EncodeInt(w, v.BaseVestingAccount.BaseAccount.Coins[_0].Amount)
// end of v.BaseVestingAccount.BaseAccount.Coins[_0]
}
EncodePubKey(w, v.BaseVestingAccount.BaseAccount.PubKey)// interface_encode
codonEncodeUvarint(w, uint64(v.BaseVestingAccount.BaseAccount.AccountNumber))
codonEncodeUvarint(w, uint64(v.BaseVestingAccount.BaseAccount.Sequence))
// end of v.BaseVestingAccount.BaseAccount
codonEncodeVarint(w, int64(len(v.BaseVestingAccount.OriginalVesting)))
for _0:=0; _0<len(v.BaseVestingAccount.OriginalVesting); _0++ {
codonEncodeString(w, v.BaseVestingAccount.OriginalVesting[_0].Denom)
EncodeInt(w, v.BaseVestingAccount.OriginalVesting[_0].Amount)
// end of v.BaseVestingAccount.OriginalVesting[_0]
}
codonEncodeVarint(w, int64(len(v.BaseVestingAccount.DelegatedFree)))
for _0:=0; _0<len(v.BaseVestingAccount.DelegatedFree); _0++ {
codonEncodeString(w, v.BaseVestingAccount.DelegatedFree[_0].Denom)
EncodeInt(w, v.BaseVestingAccount.DelegatedFree[_0].Amount)
// end of v.BaseVestingAccount.DelegatedFree[_0]
}
codonEncodeVarint(w, int64(len(v.BaseVestingAccount.DelegatedVesting)))
for _0:=0; _0<len(v.BaseVestingAccount.DelegatedVesting); _0++ {
codonEncodeString(w, v.BaseVestingAccount.DelegatedVesting[_0].Denom)
EncodeInt(w, v.BaseVestingAccount.DelegatedVesting[_0].Amount)
// end of v.BaseVestingAccount.DelegatedVesting[_0]
}
codonEncodeVarint(w, int64(v.BaseVestingAccount.EndTime))
// end of v.BaseVestingAccount
} //End of EncodeDelayedVestingAccount

func DecodeDelayedVestingAccount(bz []byte) (DelayedVestingAccount, int, error) {
var err error
var length int
var v DelayedVestingAccount
var n int
var total int
v.BaseVestingAccount = &BaseVestingAccount{}
v.BaseVestingAccount.BaseAccount = &BaseAccount{}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BaseVestingAccount.BaseAccount.Address, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.BaseVestingAccount.BaseAccount.Coins = nil
} else {
v.BaseVestingAccount.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.BaseAccount.Coins[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.BaseVestingAccount.BaseAccount.PubKey, n, err = DecodePubKey(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n // interface_decode
v.BaseVestingAccount.BaseAccount.AccountNumber = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BaseVestingAccount.BaseAccount.Sequence = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.BaseVestingAccount.BaseAccount
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.BaseVestingAccount.OriginalVesting = nil
} else {
v.BaseVestingAccount.OriginalVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.OriginalVesting[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.BaseVestingAccount.DelegatedFree = nil
} else {
v.BaseVestingAccount.DelegatedFree = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.DelegatedFree[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.BaseVestingAccount.DelegatedVesting = nil
} else {
v.BaseVestingAccount.DelegatedVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.DelegatedVesting[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.BaseVestingAccount.EndTime = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.BaseVestingAccount
return v, total, nil
} //End of DecodeDelayedVestingAccount

func RandDelayedVestingAccount(r RandSrc) DelayedVestingAccount {
var length int
var v DelayedVestingAccount
v.BaseVestingAccount = &BaseVestingAccount{}
v.BaseVestingAccount.BaseAccount = &BaseAccount{}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.BaseVestingAccount.BaseAccount.Address = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.BaseVestingAccount.BaseAccount.Coins = nil
} else {
v.BaseVestingAccount.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.BaseAccount.Coins[_0] = RandCoin(r)
}
v.BaseVestingAccount.BaseAccount.PubKey = RandPubKey(r) // interface_decode
v.BaseVestingAccount.BaseAccount.AccountNumber = r.GetUint64()
v.BaseVestingAccount.BaseAccount.Sequence = r.GetUint64()
// end of v.BaseVestingAccount.BaseAccount
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.BaseVestingAccount.OriginalVesting = nil
} else {
v.BaseVestingAccount.OriginalVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.OriginalVesting[_0] = RandCoin(r)
}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.BaseVestingAccount.DelegatedFree = nil
} else {
v.BaseVestingAccount.DelegatedFree = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseVestingAccount.DelegatedFree[_0] = RandCoin(r)
}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.BaseVestingAccount.DelegatedVesting = nil
} else {
v.BaseVestingAccount.DelegatedVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
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
if length==0 {out.BaseVestingAccount.BaseAccount.Address = nil
} else {
out.BaseVestingAccount.BaseAccount.Address = make([]uint8, length)
}
copy(out.BaseVestingAccount.BaseAccount.Address[:], in.BaseVestingAccount.BaseAccount.Address[:])
length = len(in.BaseVestingAccount.BaseAccount.Coins)
if length==0 {out.BaseVestingAccount.BaseAccount.Coins = nil
} else {
out.BaseVestingAccount.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.BaseVestingAccount.BaseAccount.Coins[_0] = DeepCopyCoin(in.BaseVestingAccount.BaseAccount.Coins[_0])
}
out.BaseVestingAccount.BaseAccount.PubKey = DeepCopyPubKey(in.BaseVestingAccount.BaseAccount.PubKey)
out.BaseVestingAccount.BaseAccount.AccountNumber = in.BaseVestingAccount.BaseAccount.AccountNumber
out.BaseVestingAccount.BaseAccount.Sequence = in.BaseVestingAccount.BaseAccount.Sequence
// end of .BaseVestingAccount.BaseAccount
length = len(in.BaseVestingAccount.OriginalVesting)
if length==0 {out.BaseVestingAccount.OriginalVesting = nil
} else {
out.BaseVestingAccount.OriginalVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.BaseVestingAccount.OriginalVesting[_0] = DeepCopyCoin(in.BaseVestingAccount.OriginalVesting[_0])
}
length = len(in.BaseVestingAccount.DelegatedFree)
if length==0 {out.BaseVestingAccount.DelegatedFree = nil
} else {
out.BaseVestingAccount.DelegatedFree = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.BaseVestingAccount.DelegatedFree[_0] = DeepCopyCoin(in.BaseVestingAccount.DelegatedFree[_0])
}
length = len(in.BaseVestingAccount.DelegatedVesting)
if length==0 {out.BaseVestingAccount.DelegatedVesting = nil
} else {
out.BaseVestingAccount.DelegatedVesting = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.BaseVestingAccount.DelegatedVesting[_0] = DeepCopyCoin(in.BaseVestingAccount.DelegatedVesting[_0])
}
out.BaseVestingAccount.EndTime = in.BaseVestingAccount.EndTime
// end of .BaseVestingAccount
return
} //End of DeepCopyDelayedVestingAccount

// Non-Interface
func EncodeModuleAccount(w *[]byte, v ModuleAccount) {
codonEncodeByteSlice(w, v.BaseAccount.Address[:])
codonEncodeVarint(w, int64(len(v.BaseAccount.Coins)))
for _0:=0; _0<len(v.BaseAccount.Coins); _0++ {
codonEncodeString(w, v.BaseAccount.Coins[_0].Denom)
EncodeInt(w, v.BaseAccount.Coins[_0].Amount)
// end of v.BaseAccount.Coins[_0]
}
EncodePubKey(w, v.BaseAccount.PubKey)// interface_encode
codonEncodeUvarint(w, uint64(v.BaseAccount.AccountNumber))
codonEncodeUvarint(w, uint64(v.BaseAccount.Sequence))
// end of v.BaseAccount
codonEncodeString(w, v.Name)
codonEncodeVarint(w, int64(len(v.Permissions)))
for _0:=0; _0<len(v.Permissions); _0++ {
codonEncodeString(w, v.Permissions[_0])
}
} //End of EncodeModuleAccount

func DecodeModuleAccount(bz []byte) (ModuleAccount, int, error) {
var err error
var length int
var v ModuleAccount
var n int
var total int
v.BaseAccount = &BaseAccount{}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BaseAccount.Address, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.BaseAccount.Coins = nil
} else {
v.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseAccount.Coins[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.BaseAccount.PubKey, n, err = DecodePubKey(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n // interface_decode
v.BaseAccount.AccountNumber = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.BaseAccount.Sequence = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.BaseAccount
v.Name = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Permissions = nil
} else {
v.Permissions = make([]string, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of string
v.Permissions[_0] = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeModuleAccount

func RandModuleAccount(r RandSrc) ModuleAccount {
var length int
var v ModuleAccount
v.BaseAccount = &BaseAccount{}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.BaseAccount.Address = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.BaseAccount.Coins = nil
} else {
v.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.BaseAccount.Coins[_0] = RandCoin(r)
}
v.BaseAccount.PubKey = RandPubKey(r) // interface_decode
v.BaseAccount.AccountNumber = r.GetUint64()
v.BaseAccount.Sequence = r.GetUint64()
// end of v.BaseAccount
v.Name = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Permissions = nil
} else {
v.Permissions = make([]string, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of string
v.Permissions[_0] = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
}
return v
} //End of RandModuleAccount

func DeepCopyModuleAccount(in ModuleAccount) (out ModuleAccount) {
var length int
out.BaseAccount = &BaseAccount{}
length = len(in.BaseAccount.Address)
if length==0 {out.BaseAccount.Address = nil
} else {
out.BaseAccount.Address = make([]uint8, length)
}
copy(out.BaseAccount.Address[:], in.BaseAccount.Address[:])
length = len(in.BaseAccount.Coins)
if length==0 {out.BaseAccount.Coins = nil
} else {
out.BaseAccount.Coins = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.BaseAccount.Coins[_0] = DeepCopyCoin(in.BaseAccount.Coins[_0])
}
out.BaseAccount.PubKey = DeepCopyPubKey(in.BaseAccount.PubKey)
out.BaseAccount.AccountNumber = in.BaseAccount.AccountNumber
out.BaseAccount.Sequence = in.BaseAccount.Sequence
// end of .BaseAccount
out.Name = in.Name
length = len(in.Permissions)
if length==0 {out.Permissions = nil
} else {
out.Permissions = make([]string, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of string
out.Permissions[_0] = in.Permissions[_0]
}
return
} //End of DeepCopyModuleAccount

// Non-Interface
func EncodeStdTx(w *[]byte, v StdTx) {
codonEncodeVarint(w, int64(len(v.Msgs)))
for _0:=0; _0<len(v.Msgs); _0++ {
EncodeMsg(w, v.Msgs[_0])// interface_encode
}
codonEncodeVarint(w, int64(len(v.Fee.Amount)))
for _0:=0; _0<len(v.Fee.Amount); _0++ {
codonEncodeString(w, v.Fee.Amount[_0].Denom)
EncodeInt(w, v.Fee.Amount[_0].Amount)
// end of v.Fee.Amount[_0]
}
codonEncodeUvarint(w, uint64(v.Fee.Gas))
// end of v.Fee
codonEncodeVarint(w, int64(len(v.Signatures)))
for _0:=0; _0<len(v.Signatures); _0++ {
EncodePubKey(w, v.Signatures[_0].PubKey)// interface_encode
codonEncodeByteSlice(w, v.Signatures[_0].Signature[:])
// end of v.Signatures[_0]
}
codonEncodeString(w, v.Memo)
} //End of EncodeStdTx

func DecodeStdTx(bz []byte) (StdTx, int, error) {
var err error
var length int
var v StdTx
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Msgs = nil
} else {
v.Msgs = make([]Msg, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of interface
v.Msgs[_0], n, err = DecodeMsg(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Fee.Amount = nil
} else {
v.Fee.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Fee.Amount[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.Fee.Gas = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Fee
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Signatures = nil
} else {
v.Signatures = make([]StdSignature, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Signatures[_0], n, err = DecodeStdSignature(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.Memo = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeStdTx

func RandStdTx(r RandSrc) StdTx {
var length int
var v StdTx
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Msgs = nil
} else {
v.Msgs = make([]Msg, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of interface
v.Msgs[_0] = RandMsg(r)
}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Fee.Amount = nil
} else {
v.Fee.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Fee.Amount[_0] = RandCoin(r)
}
v.Fee.Gas = r.GetUint64()
// end of v.Fee
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Signatures = nil
} else {
v.Signatures = make([]StdSignature, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Signatures[_0] = RandStdSignature(r)
}
v.Memo = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
return v
} //End of RandStdTx

func DeepCopyStdTx(in StdTx) (out StdTx) {
var length int
length = len(in.Msgs)
if length==0 {out.Msgs = nil
} else {
out.Msgs = make([]Msg, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of interface
out.Msgs[_0] = DeepCopyMsg(in.Msgs[_0])
}
length = len(in.Fee.Amount)
if length==0 {out.Fee.Amount = nil
} else {
out.Fee.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Fee.Amount[_0] = DeepCopyCoin(in.Fee.Amount[_0])
}
out.Fee.Gas = in.Fee.Gas
// end of .Fee
length = len(in.Signatures)
if length==0 {out.Signatures = nil
} else {
out.Signatures = make([]StdSignature, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Signatures[_0] = DeepCopyStdSignature(in.Signatures[_0])
}
out.Memo = in.Memo
return
} //End of DeepCopyStdTx

// Non-Interface
func EncodeMsgBeginRedelegate(w *[]byte, v MsgBeginRedelegate) {
codonEncodeByteSlice(w, v.DelegatorAddress[:])
codonEncodeByteSlice(w, v.ValidatorSrcAddress[:])
codonEncodeByteSlice(w, v.ValidatorDstAddress[:])
codonEncodeString(w, v.Amount.Denom)
EncodeInt(w, v.Amount.Amount)
// end of v.Amount
} //End of EncodeMsgBeginRedelegate

func DecodeMsgBeginRedelegate(bz []byte) (MsgBeginRedelegate, int, error) {
var err error
var length int
var v MsgBeginRedelegate
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.DelegatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorSrcAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorDstAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Amount.Denom = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Amount.Amount, n, err = DecodeInt(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Amount
return v, total, nil
} //End of DecodeMsgBeginRedelegate

func RandMsgBeginRedelegate(r RandSrc) MsgBeginRedelegate {
var length int
var v MsgBeginRedelegate
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.DelegatorAddress = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ValidatorSrcAddress = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ValidatorDstAddress = r.GetBytes(length)
v.Amount.Denom = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Amount.Amount = RandInt(r)
// end of v.Amount
return v
} //End of RandMsgBeginRedelegate

func DeepCopyMsgBeginRedelegate(in MsgBeginRedelegate) (out MsgBeginRedelegate) {
var length int
length = len(in.DelegatorAddress)
if length==0 {out.DelegatorAddress = nil
} else {
out.DelegatorAddress = make([]uint8, length)
}
copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
length = len(in.ValidatorSrcAddress)
if length==0 {out.ValidatorSrcAddress = nil
} else {
out.ValidatorSrcAddress = make([]uint8, length)
}
copy(out.ValidatorSrcAddress[:], in.ValidatorSrcAddress[:])
length = len(in.ValidatorDstAddress)
if length==0 {out.ValidatorDstAddress = nil
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
codonEncodeString(w, v.Description.Moniker)
codonEncodeString(w, v.Description.Identity)
codonEncodeString(w, v.Description.Website)
codonEncodeString(w, v.Description.Details)
// end of v.Description
EncodeDec(w, v.Commission.Rate)
EncodeDec(w, v.Commission.MaxRate)
EncodeDec(w, v.Commission.MaxChangeRate)
// end of v.Commission
EncodeInt(w, v.MinSelfDelegation)
codonEncodeByteSlice(w, v.DelegatorAddress[:])
codonEncodeByteSlice(w, v.ValidatorAddress[:])
EncodePubKey(w, v.PubKey)// interface_encode
codonEncodeString(w, v.Value.Denom)
EncodeInt(w, v.Value.Amount)
// end of v.Value
} //End of EncodeMsgCreateValidator

func DecodeMsgCreateValidator(bz []byte) (MsgCreateValidator, int, error) {
var err error
var length int
var v MsgCreateValidator
var n int
var total int
v.Description.Moniker = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description.Identity = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description.Website = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description.Details = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Description
v.Commission.Rate, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Commission.MaxRate, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Commission.MaxChangeRate, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Commission
v.MinSelfDelegation, n, err = DecodeInt(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.DelegatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.PubKey, n, err = DecodePubKey(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n // interface_decode
v.Value.Denom = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Value.Amount, n, err = DecodeInt(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Value
return v, total, nil
} //End of DecodeMsgCreateValidator

func RandMsgCreateValidator(r RandSrc) MsgCreateValidator {
var length int
var v MsgCreateValidator
v.Description.Moniker = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description.Identity = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description.Website = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description.Details = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
// end of v.Description
v.Commission.Rate = RandDec(r)
v.Commission.MaxRate = RandDec(r)
v.Commission.MaxChangeRate = RandDec(r)
// end of v.Commission
v.MinSelfDelegation = RandInt(r)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.DelegatorAddress = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ValidatorAddress = r.GetBytes(length)
v.PubKey = RandPubKey(r) // interface_decode
v.Value.Denom = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
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
if length==0 {out.DelegatorAddress = nil
} else {
out.DelegatorAddress = make([]uint8, length)
}
copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
length = len(in.ValidatorAddress)
if length==0 {out.ValidatorAddress = nil
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
codonEncodeByteSlice(w, v.DelegatorAddress[:])
codonEncodeByteSlice(w, v.ValidatorAddress[:])
codonEncodeString(w, v.Amount.Denom)
EncodeInt(w, v.Amount.Amount)
// end of v.Amount
} //End of EncodeMsgDelegate

func DecodeMsgDelegate(bz []byte) (MsgDelegate, int, error) {
var err error
var length int
var v MsgDelegate
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.DelegatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Amount.Denom = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Amount.Amount, n, err = DecodeInt(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Amount
return v, total, nil
} //End of DecodeMsgDelegate

func RandMsgDelegate(r RandSrc) MsgDelegate {
var length int
var v MsgDelegate
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.DelegatorAddress = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ValidatorAddress = r.GetBytes(length)
v.Amount.Denom = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Amount.Amount = RandInt(r)
// end of v.Amount
return v
} //End of RandMsgDelegate

func DeepCopyMsgDelegate(in MsgDelegate) (out MsgDelegate) {
var length int
length = len(in.DelegatorAddress)
if length==0 {out.DelegatorAddress = nil
} else {
out.DelegatorAddress = make([]uint8, length)
}
copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
length = len(in.ValidatorAddress)
if length==0 {out.ValidatorAddress = nil
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
codonEncodeString(w, v.Description.Moniker)
codonEncodeString(w, v.Description.Identity)
codonEncodeString(w, v.Description.Website)
codonEncodeString(w, v.Description.Details)
// end of v.Description
codonEncodeByteSlice(w, v.ValidatorAddress[:])
EncodeDec(w, *(v.CommissionRate))
EncodeInt(w, *(v.MinSelfDelegation))
} //End of EncodeMsgEditValidator

func DecodeMsgEditValidator(bz []byte) (MsgEditValidator, int, error) {
var err error
var length int
var v MsgEditValidator
var n int
var total int
v.Description.Moniker = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description.Identity = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description.Website = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description.Details = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Description
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.CommissionRate = &SdkDec{}
*(v.CommissionRate), n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.MinSelfDelegation = &SdkInt{}
*(v.MinSelfDelegation), n, err = DecodeInt(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeMsgEditValidator

func RandMsgEditValidator(r RandSrc) MsgEditValidator {
var length int
var v MsgEditValidator
v.Description.Moniker = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description.Identity = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description.Website = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description.Details = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
// end of v.Description
length = 1+int(r.GetUint()%(MaxSliceLength-1))
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
if length==0 {out.ValidatorAddress = nil
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
codonEncodeByteSlice(w, v.DelegatorAddress[:])
codonEncodeByteSlice(w, v.WithdrawAddress[:])
} //End of EncodeMsgSetWithdrawAddress

func DecodeMsgSetWithdrawAddress(bz []byte) (MsgSetWithdrawAddress, int, error) {
var err error
var length int
var v MsgSetWithdrawAddress
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.DelegatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.WithdrawAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeMsgSetWithdrawAddress

func RandMsgSetWithdrawAddress(r RandSrc) MsgSetWithdrawAddress {
var length int
var v MsgSetWithdrawAddress
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.DelegatorAddress = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.WithdrawAddress = r.GetBytes(length)
return v
} //End of RandMsgSetWithdrawAddress

func DeepCopyMsgSetWithdrawAddress(in MsgSetWithdrawAddress) (out MsgSetWithdrawAddress) {
var length int
length = len(in.DelegatorAddress)
if length==0 {out.DelegatorAddress = nil
} else {
out.DelegatorAddress = make([]uint8, length)
}
copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
length = len(in.WithdrawAddress)
if length==0 {out.WithdrawAddress = nil
} else {
out.WithdrawAddress = make([]uint8, length)
}
copy(out.WithdrawAddress[:], in.WithdrawAddress[:])
return
} //End of DeepCopyMsgSetWithdrawAddress

// Non-Interface
func EncodeMsgUndelegate(w *[]byte, v MsgUndelegate) {
codonEncodeByteSlice(w, v.DelegatorAddress[:])
codonEncodeByteSlice(w, v.ValidatorAddress[:])
codonEncodeString(w, v.Amount.Denom)
EncodeInt(w, v.Amount.Amount)
// end of v.Amount
} //End of EncodeMsgUndelegate

func DecodeMsgUndelegate(bz []byte) (MsgUndelegate, int, error) {
var err error
var length int
var v MsgUndelegate
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.DelegatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Amount.Denom = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Amount.Amount, n, err = DecodeInt(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Amount
return v, total, nil
} //End of DecodeMsgUndelegate

func RandMsgUndelegate(r RandSrc) MsgUndelegate {
var length int
var v MsgUndelegate
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.DelegatorAddress = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ValidatorAddress = r.GetBytes(length)
v.Amount.Denom = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Amount.Amount = RandInt(r)
// end of v.Amount
return v
} //End of RandMsgUndelegate

func DeepCopyMsgUndelegate(in MsgUndelegate) (out MsgUndelegate) {
var length int
length = len(in.DelegatorAddress)
if length==0 {out.DelegatorAddress = nil
} else {
out.DelegatorAddress = make([]uint8, length)
}
copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
length = len(in.ValidatorAddress)
if length==0 {out.ValidatorAddress = nil
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
codonEncodeByteSlice(w, v.ValidatorAddr[:])
} //End of EncodeMsgUnjail

func DecodeMsgUnjail(bz []byte) (MsgUnjail, int, error) {
var err error
var length int
var v MsgUnjail
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorAddr, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeMsgUnjail

func RandMsgUnjail(r RandSrc) MsgUnjail {
var length int
var v MsgUnjail
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ValidatorAddr = r.GetBytes(length)
return v
} //End of RandMsgUnjail

func DeepCopyMsgUnjail(in MsgUnjail) (out MsgUnjail) {
var length int
length = len(in.ValidatorAddr)
if length==0 {out.ValidatorAddr = nil
} else {
out.ValidatorAddr = make([]uint8, length)
}
copy(out.ValidatorAddr[:], in.ValidatorAddr[:])
return
} //End of DeepCopyMsgUnjail

// Non-Interface
func EncodeMsgWithdrawDelegatorReward(w *[]byte, v MsgWithdrawDelegatorReward) {
codonEncodeByteSlice(w, v.DelegatorAddress[:])
codonEncodeByteSlice(w, v.ValidatorAddress[:])
} //End of EncodeMsgWithdrawDelegatorReward

func DecodeMsgWithdrawDelegatorReward(bz []byte) (MsgWithdrawDelegatorReward, int, error) {
var err error
var length int
var v MsgWithdrawDelegatorReward
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.DelegatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeMsgWithdrawDelegatorReward

func RandMsgWithdrawDelegatorReward(r RandSrc) MsgWithdrawDelegatorReward {
var length int
var v MsgWithdrawDelegatorReward
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.DelegatorAddress = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ValidatorAddress = r.GetBytes(length)
return v
} //End of RandMsgWithdrawDelegatorReward

func DeepCopyMsgWithdrawDelegatorReward(in MsgWithdrawDelegatorReward) (out MsgWithdrawDelegatorReward) {
var length int
length = len(in.DelegatorAddress)
if length==0 {out.DelegatorAddress = nil
} else {
out.DelegatorAddress = make([]uint8, length)
}
copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
length = len(in.ValidatorAddress)
if length==0 {out.ValidatorAddress = nil
} else {
out.ValidatorAddress = make([]uint8, length)
}
copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
return
} //End of DeepCopyMsgWithdrawDelegatorReward

// Non-Interface
func EncodeMsgWithdrawValidatorCommission(w *[]byte, v MsgWithdrawValidatorCommission) {
codonEncodeByteSlice(w, v.ValidatorAddress[:])
} //End of EncodeMsgWithdrawValidatorCommission

func DecodeMsgWithdrawValidatorCommission(bz []byte) (MsgWithdrawValidatorCommission, int, error) {
var err error
var length int
var v MsgWithdrawValidatorCommission
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeMsgWithdrawValidatorCommission

func RandMsgWithdrawValidatorCommission(r RandSrc) MsgWithdrawValidatorCommission {
var length int
var v MsgWithdrawValidatorCommission
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ValidatorAddress = r.GetBytes(length)
return v
} //End of RandMsgWithdrawValidatorCommission

func DeepCopyMsgWithdrawValidatorCommission(in MsgWithdrawValidatorCommission) (out MsgWithdrawValidatorCommission) {
var length int
length = len(in.ValidatorAddress)
if length==0 {out.ValidatorAddress = nil
} else {
out.ValidatorAddress = make([]uint8, length)
}
copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
return
} //End of DeepCopyMsgWithdrawValidatorCommission

// Non-Interface
func EncodeMsgDeposit(w *[]byte, v MsgDeposit) {
codonEncodeUvarint(w, uint64(v.ProposalID))
codonEncodeByteSlice(w, v.Depositor[:])
codonEncodeVarint(w, int64(len(v.Amount)))
for _0:=0; _0<len(v.Amount); _0++ {
codonEncodeString(w, v.Amount[_0].Denom)
EncodeInt(w, v.Amount[_0].Amount)
// end of v.Amount[_0]
}
} //End of EncodeMsgDeposit

func DecodeMsgDeposit(bz []byte) (MsgDeposit, int, error) {
var err error
var length int
var v MsgDeposit
var n int
var total int
v.ProposalID = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Depositor, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Amount = nil
} else {
v.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Amount[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeMsgDeposit

func RandMsgDeposit(r RandSrc) MsgDeposit {
var length int
var v MsgDeposit
v.ProposalID = r.GetUint64()
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.Depositor = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Amount = nil
} else {
v.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Amount[_0] = RandCoin(r)
}
return v
} //End of RandMsgDeposit

func DeepCopyMsgDeposit(in MsgDeposit) (out MsgDeposit) {
var length int
out.ProposalID = in.ProposalID
length = len(in.Depositor)
if length==0 {out.Depositor = nil
} else {
out.Depositor = make([]uint8, length)
}
copy(out.Depositor[:], in.Depositor[:])
length = len(in.Amount)
if length==0 {out.Amount = nil
} else {
out.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Amount[_0] = DeepCopyCoin(in.Amount[_0])
}
return
} //End of DeepCopyMsgDeposit

// Non-Interface
func EncodeMsgVote(w *[]byte, v MsgVote) {
codonEncodeUvarint(w, uint64(v.ProposalID))
codonEncodeByteSlice(w, v.Voter[:])
codonEncodeUint8(w, uint8(v.Option))
} //End of EncodeMsgVote

func DecodeMsgVote(bz []byte) (MsgVote, int, error) {
var err error
var length int
var v MsgVote
var n int
var total int
v.ProposalID = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Voter, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Option = VoteOption(codonDecodeUint8(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeMsgVote

func RandMsgVote(r RandSrc) MsgVote {
var length int
var v MsgVote
v.ProposalID = r.GetUint64()
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.Voter = r.GetBytes(length)
v.Option = VoteOption(r.GetUint8())
return v
} //End of RandMsgVote

func DeepCopyMsgVote(in MsgVote) (out MsgVote) {
var length int
out.ProposalID = in.ProposalID
length = len(in.Voter)
if length==0 {out.Voter = nil
} else {
out.Voter = make([]uint8, length)
}
copy(out.Voter[:], in.Voter[:])
out.Option = in.Option
return
} //End of DeepCopyMsgVote

// Non-Interface
func EncodeParameterChangeProposal(w *[]byte, v ParameterChangeProposal) {
codonEncodeString(w, v.Title)
codonEncodeString(w, v.Description)
codonEncodeVarint(w, int64(len(v.Changes)))
for _0:=0; _0<len(v.Changes); _0++ {
codonEncodeString(w, v.Changes[_0].Subspace)
codonEncodeString(w, v.Changes[_0].Key)
codonEncodeString(w, v.Changes[_0].Subkey)
codonEncodeString(w, v.Changes[_0].Value)
// end of v.Changes[_0]
}
} //End of EncodeParameterChangeProposal

func DecodeParameterChangeProposal(bz []byte) (ParameterChangeProposal, int, error) {
var err error
var length int
var v ParameterChangeProposal
var n int
var total int
v.Title = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Changes = nil
} else {
v.Changes = make([]ParamChange, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Changes[_0], n, err = DecodeParamChange(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeParameterChangeProposal

func RandParameterChangeProposal(r RandSrc) ParameterChangeProposal {
var length int
var v ParameterChangeProposal
v.Title = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Changes = nil
} else {
v.Changes = make([]ParamChange, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Changes[_0] = RandParamChange(r)
}
return v
} //End of RandParameterChangeProposal

func DeepCopyParameterChangeProposal(in ParameterChangeProposal) (out ParameterChangeProposal) {
var length int
out.Title = in.Title
out.Description = in.Description
length = len(in.Changes)
if length==0 {out.Changes = nil
} else {
out.Changes = make([]ParamChange, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Changes[_0] = DeepCopyParamChange(in.Changes[_0])
}
return
} //End of DeepCopyParameterChangeProposal

// Non-Interface
func EncodeSoftwareUpgradeProposal(w *[]byte, v SoftwareUpgradeProposal) {
codonEncodeString(w, v.Title)
codonEncodeString(w, v.Description)
} //End of EncodeSoftwareUpgradeProposal

func DecodeSoftwareUpgradeProposal(bz []byte) (SoftwareUpgradeProposal, int, error) {
var err error
var v SoftwareUpgradeProposal
var n int
var total int
v.Title = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeSoftwareUpgradeProposal

func RandSoftwareUpgradeProposal(r RandSrc) SoftwareUpgradeProposal {
var v SoftwareUpgradeProposal
v.Title = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
return v
} //End of RandSoftwareUpgradeProposal

func DeepCopySoftwareUpgradeProposal(in SoftwareUpgradeProposal) (out SoftwareUpgradeProposal) {
out.Title = in.Title
out.Description = in.Description
return
} //End of DeepCopySoftwareUpgradeProposal

// Non-Interface
func EncodeTextProposal(w *[]byte, v TextProposal) {
codonEncodeString(w, v.Title)
codonEncodeString(w, v.Description)
} //End of EncodeTextProposal

func DecodeTextProposal(bz []byte) (TextProposal, int, error) {
var err error
var v TextProposal
var n int
var total int
v.Title = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeTextProposal

func RandTextProposal(r RandSrc) TextProposal {
var v TextProposal
v.Title = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
return v
} //End of RandTextProposal

func DeepCopyTextProposal(in TextProposal) (out TextProposal) {
out.Title = in.Title
out.Description = in.Description
return
} //End of DeepCopyTextProposal

// Non-Interface
func EncodeCommunityPoolSpendProposal(w *[]byte, v CommunityPoolSpendProposal) {
codonEncodeString(w, v.Title)
codonEncodeString(w, v.Description)
codonEncodeByteSlice(w, v.Recipient[:])
codonEncodeVarint(w, int64(len(v.Amount)))
for _0:=0; _0<len(v.Amount); _0++ {
codonEncodeString(w, v.Amount[_0].Denom)
EncodeInt(w, v.Amount[_0].Amount)
// end of v.Amount[_0]
}
} //End of EncodeCommunityPoolSpendProposal

func DecodeCommunityPoolSpendProposal(bz []byte) (CommunityPoolSpendProposal, int, error) {
var err error
var length int
var v CommunityPoolSpendProposal
var n int
var total int
v.Title = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Recipient, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Amount = nil
} else {
v.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Amount[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeCommunityPoolSpendProposal

func RandCommunityPoolSpendProposal(r RandSrc) CommunityPoolSpendProposal {
var length int
var v CommunityPoolSpendProposal
v.Title = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.Recipient = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Amount = nil
} else {
v.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Amount[_0] = RandCoin(r)
}
return v
} //End of RandCommunityPoolSpendProposal

func DeepCopyCommunityPoolSpendProposal(in CommunityPoolSpendProposal) (out CommunityPoolSpendProposal) {
var length int
out.Title = in.Title
out.Description = in.Description
length = len(in.Recipient)
if length==0 {out.Recipient = nil
} else {
out.Recipient = make([]uint8, length)
}
copy(out.Recipient[:], in.Recipient[:])
length = len(in.Amount)
if length==0 {out.Amount = nil
} else {
out.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Amount[_0] = DeepCopyCoin(in.Amount[_0])
}
return
} //End of DeepCopyCommunityPoolSpendProposal

// Non-Interface
func EncodeMsgMultiSend(w *[]byte, v MsgMultiSend) {
codonEncodeVarint(w, int64(len(v.Inputs)))
for _0:=0; _0<len(v.Inputs); _0++ {
codonEncodeByteSlice(w, v.Inputs[_0].Address[:])
codonEncodeVarint(w, int64(len(v.Inputs[_0].Coins)))
for _1:=0; _1<len(v.Inputs[_0].Coins); _1++ {
codonEncodeString(w, v.Inputs[_0].Coins[_1].Denom)
EncodeInt(w, v.Inputs[_0].Coins[_1].Amount)
// end of v.Inputs[_0].Coins[_1]
}
// end of v.Inputs[_0]
}
codonEncodeVarint(w, int64(len(v.Outputs)))
for _0:=0; _0<len(v.Outputs); _0++ {
codonEncodeByteSlice(w, v.Outputs[_0].Address[:])
codonEncodeVarint(w, int64(len(v.Outputs[_0].Coins)))
for _1:=0; _1<len(v.Outputs[_0].Coins); _1++ {
codonEncodeString(w, v.Outputs[_0].Coins[_1].Denom)
EncodeInt(w, v.Outputs[_0].Coins[_1].Amount)
// end of v.Outputs[_0].Coins[_1]
}
// end of v.Outputs[_0]
}
} //End of EncodeMsgMultiSend

func DecodeMsgMultiSend(bz []byte) (MsgMultiSend, int, error) {
var err error
var length int
var v MsgMultiSend
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Inputs = nil
} else {
v.Inputs = make([]Input, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Inputs[_0], n, err = DecodeInput(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Outputs = nil
} else {
v.Outputs = make([]Output, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Outputs[_0], n, err = DecodeOutput(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeMsgMultiSend

func RandMsgMultiSend(r RandSrc) MsgMultiSend {
var length int
var v MsgMultiSend
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Inputs = nil
} else {
v.Inputs = make([]Input, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Inputs[_0] = RandInput(r)
}
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Outputs = nil
} else {
v.Outputs = make([]Output, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Outputs[_0] = RandOutput(r)
}
return v
} //End of RandMsgMultiSend

func DeepCopyMsgMultiSend(in MsgMultiSend) (out MsgMultiSend) {
var length int
length = len(in.Inputs)
if length==0 {out.Inputs = nil
} else {
out.Inputs = make([]Input, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Inputs[_0] = DeepCopyInput(in.Inputs[_0])
}
length = len(in.Outputs)
if length==0 {out.Outputs = nil
} else {
out.Outputs = make([]Output, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Outputs[_0] = DeepCopyOutput(in.Outputs[_0])
}
return
} //End of DeepCopyMsgMultiSend

// Non-Interface
func EncodeFeePool(w *[]byte, v FeePool) {
codonEncodeVarint(w, int64(len(v.CommunityPool)))
for _0:=0; _0<len(v.CommunityPool); _0++ {
codonEncodeString(w, v.CommunityPool[_0].Denom)
EncodeDec(w, v.CommunityPool[_0].Amount)
// end of v.CommunityPool[_0]
}
} //End of EncodeFeePool

func DecodeFeePool(bz []byte) (FeePool, int, error) {
var err error
var length int
var v FeePool
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.CommunityPool = nil
} else {
v.CommunityPool = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.CommunityPool[_0], n, err = DecodeDecCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeFeePool

func RandFeePool(r RandSrc) FeePool {
var length int
var v FeePool
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.CommunityPool = nil
} else {
v.CommunityPool = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.CommunityPool[_0] = RandDecCoin(r)
}
return v
} //End of RandFeePool

func DeepCopyFeePool(in FeePool) (out FeePool) {
var length int
length = len(in.CommunityPool)
if length==0 {out.CommunityPool = nil
} else {
out.CommunityPool = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.CommunityPool[_0] = DeepCopyDecCoin(in.CommunityPool[_0])
}
return
} //End of DeepCopyFeePool

// Non-Interface
func EncodeMsgSend(w *[]byte, v MsgSend) {
codonEncodeByteSlice(w, v.FromAddress[:])
codonEncodeByteSlice(w, v.ToAddress[:])
codonEncodeVarint(w, int64(len(v.Amount)))
for _0:=0; _0<len(v.Amount); _0++ {
codonEncodeString(w, v.Amount[_0].Denom)
EncodeInt(w, v.Amount[_0].Amount)
// end of v.Amount[_0]
}
} //End of EncodeMsgSend

func DecodeMsgSend(bz []byte) (MsgSend, int, error) {
var err error
var length int
var v MsgSend
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.FromAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ToAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Amount = nil
} else {
v.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Amount[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeMsgSend

func RandMsgSend(r RandSrc) MsgSend {
var length int
var v MsgSend
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.FromAddress = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ToAddress = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Amount = nil
} else {
v.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Amount[_0] = RandCoin(r)
}
return v
} //End of RandMsgSend

func DeepCopyMsgSend(in MsgSend) (out MsgSend) {
var length int
length = len(in.FromAddress)
if length==0 {out.FromAddress = nil
} else {
out.FromAddress = make([]uint8, length)
}
copy(out.FromAddress[:], in.FromAddress[:])
length = len(in.ToAddress)
if length==0 {out.ToAddress = nil
} else {
out.ToAddress = make([]uint8, length)
}
copy(out.ToAddress[:], in.ToAddress[:])
length = len(in.Amount)
if length==0 {out.Amount = nil
} else {
out.Amount = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Amount[_0] = DeepCopyCoin(in.Amount[_0])
}
return
} //End of DeepCopyMsgSend

// Non-Interface
func EncodeMsgVerifyInvariant(w *[]byte, v MsgVerifyInvariant) {
codonEncodeByteSlice(w, v.Sender[:])
codonEncodeString(w, v.InvariantModuleName)
codonEncodeString(w, v.InvariantRoute)
} //End of EncodeMsgVerifyInvariant

func DecodeMsgVerifyInvariant(bz []byte) (MsgVerifyInvariant, int, error) {
var err error
var length int
var v MsgVerifyInvariant
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Sender, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.InvariantModuleName = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.InvariantRoute = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeMsgVerifyInvariant

func RandMsgVerifyInvariant(r RandSrc) MsgVerifyInvariant {
var length int
var v MsgVerifyInvariant
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.Sender = r.GetBytes(length)
v.InvariantModuleName = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.InvariantRoute = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
return v
} //End of RandMsgVerifyInvariant

func DeepCopyMsgVerifyInvariant(in MsgVerifyInvariant) (out MsgVerifyInvariant) {
var length int
length = len(in.Sender)
if length==0 {out.Sender = nil
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
codonEncodeVarint(w, int64(len(v.Total)))
for _0:=0; _0<len(v.Total); _0++ {
codonEncodeString(w, v.Total[_0].Denom)
EncodeInt(w, v.Total[_0].Amount)
// end of v.Total[_0]
}
} //End of EncodeSupply

func DecodeSupply(bz []byte) (Supply, int, error) {
var err error
var length int
var v Supply
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Total = nil
} else {
v.Total = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Total[_0], n, err = DecodeCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeSupply

func RandSupply(r RandSrc) Supply {
var length int
var v Supply
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Total = nil
} else {
v.Total = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Total[_0] = RandCoin(r)
}
return v
} //End of RandSupply

func DeepCopySupply(in Supply) (out Supply) {
var length int
length = len(in.Total)
if length==0 {out.Total = nil
} else {
out.Total = make([]Coin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Total[_0] = DeepCopyCoin(in.Total[_0])
}
return
} //End of DeepCopySupply

// Non-Interface
func EncodeAccAddressList(w *[]byte, v AccAddressList) {
codonEncodeVarint(w, int64(len(v)))
for _0:=0; _0<len(v); _0++ {
codonEncodeByteSlice(w, v[_0][:])
}
} //End of EncodeAccAddressList

func DecodeAccAddressList(bz []byte) (AccAddressList, int, error) {
var err error
var length int
var v AccAddressList
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v = nil
} else {
v = make([]AccAddress, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of slice
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v[_0], n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeAccAddressList

func RandAccAddressList(r RandSrc) AccAddressList {
var length int
var v AccAddressList
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v = nil
} else {
v = make([]AccAddress, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of slice
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v[_0] = r.GetBytes(length)
}
return v
} //End of RandAccAddressList

func DeepCopyAccAddressList(in AccAddressList) (out AccAddressList) {
var length int
length = len(in)
if length==0 {out = nil
} else {
out = make([]AccAddress, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of slice
length = len(in[_0])
if length==0 {out[_0] = nil
} else {
out[_0] = make([]uint8, length)
}
copy(out[_0][:], in[_0][:])
}
return
} //End of DeepCopyAccAddressList

// Non-Interface
func EncodeCommitInfo(w *[]byte, v CommitInfo) {
codonEncodeVarint(w, int64(v.Version))
codonEncodeVarint(w, int64(len(v.StoreInfos)))
for _0:=0; _0<len(v.StoreInfos); _0++ {
codonEncodeString(w, v.StoreInfos[_0].Name)
codonEncodeVarint(w, int64(v.StoreInfos[_0].Core.CommitID.Version))
codonEncodeByteSlice(w, v.StoreInfos[_0].Core.CommitID.Hash[:])
// end of v.StoreInfos[_0].Core.CommitID
// end of v.StoreInfos[_0].Core
// end of v.StoreInfos[_0]
}
} //End of EncodeCommitInfo

func DecodeCommitInfo(bz []byte) (CommitInfo, int, error) {
var err error
var length int
var v CommitInfo
var n int
var total int
v.Version = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.StoreInfos = nil
} else {
v.StoreInfos = make([]StoreInfo, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.StoreInfos[_0], n, err = DecodeStoreInfo(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeCommitInfo

func RandCommitInfo(r RandSrc) CommitInfo {
var length int
var v CommitInfo
v.Version = r.GetInt64()
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.StoreInfos = nil
} else {
v.StoreInfos = make([]StoreInfo, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.StoreInfos[_0] = RandStoreInfo(r)
}
return v
} //End of RandCommitInfo

func DeepCopyCommitInfo(in CommitInfo) (out CommitInfo) {
var length int
out.Version = in.Version
length = len(in.StoreInfos)
if length==0 {out.StoreInfos = nil
} else {
out.StoreInfos = make([]StoreInfo, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.StoreInfos[_0] = DeepCopyStoreInfo(in.StoreInfos[_0])
}
return
} //End of DeepCopyCommitInfo

// Non-Interface
func EncodeStoreInfo(w *[]byte, v StoreInfo) {
codonEncodeString(w, v.Name)
codonEncodeVarint(w, int64(v.Core.CommitID.Version))
codonEncodeByteSlice(w, v.Core.CommitID.Hash[:])
// end of v.Core.CommitID
// end of v.Core
} //End of EncodeStoreInfo

func DecodeStoreInfo(bz []byte) (StoreInfo, int, error) {
var err error
var length int
var v StoreInfo
var n int
var total int
v.Name = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Core.CommitID.Version = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Core.CommitID.Hash, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Core.CommitID
// end of v.Core
return v, total, nil
} //End of DecodeStoreInfo

func RandStoreInfo(r RandSrc) StoreInfo {
var length int
var v StoreInfo
v.Name = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Core.CommitID.Version = r.GetInt64()
length = 1+int(r.GetUint()%(MaxSliceLength-1))
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
if length==0 {out.Core.CommitID.Hash = nil
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
codonEncodeByteSlice(w, v.OperatorAddress[:])
EncodePubKey(w, v.ConsPubKey)// interface_encode
codonEncodeBool(w, v.Jailed)
codonEncodeUint8(w, uint8(v.Status))
EncodeInt(w, v.Tokens)
EncodeDec(w, v.DelegatorShares)
codonEncodeString(w, v.Description.Moniker)
codonEncodeString(w, v.Description.Identity)
codonEncodeString(w, v.Description.Website)
codonEncodeString(w, v.Description.Details)
// end of v.Description
codonEncodeVarint(w, int64(v.UnbondingHeight))
EncodeTime(w, v.UnbondingCompletionTime)
EncodeDec(w, v.Commission.CommissionRates.Rate)
EncodeDec(w, v.Commission.CommissionRates.MaxRate)
EncodeDec(w, v.Commission.CommissionRates.MaxChangeRate)
// end of v.Commission.CommissionRates
EncodeTime(w, v.Commission.UpdateTime)
// end of v.Commission
EncodeInt(w, v.MinSelfDelegation)
} //End of EncodeValidator

func DecodeValidator(bz []byte) (Validator, int, error) {
var err error
var length int
var v Validator
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.OperatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ConsPubKey, n, err = DecodePubKey(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n // interface_decode
v.Jailed = bool(codonDecodeBool(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Status = BondStatus(codonDecodeUint8(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Tokens, n, err = DecodeInt(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.DelegatorShares, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description.Moniker = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description.Identity = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description.Website = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Description.Details = string(codonDecodeString(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Description
v.UnbondingHeight = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.UnbondingCompletionTime, n, err = DecodeTime(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Commission.CommissionRates.Rate, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Commission.CommissionRates.MaxRate, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Commission.CommissionRates.MaxChangeRate, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Commission.CommissionRates
v.Commission.UpdateTime, n, err = DecodeTime(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
// end of v.Commission
v.MinSelfDelegation, n, err = DecodeInt(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeValidator

func RandValidator(r RandSrc) Validator {
var length int
var v Validator
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.OperatorAddress = r.GetBytes(length)
v.ConsPubKey = RandPubKey(r) // interface_decode
v.Jailed = r.GetBool()
v.Status = BondStatus(r.GetUint8())
v.Tokens = RandInt(r)
v.DelegatorShares = RandDec(r)
v.Description.Moniker = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description.Identity = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description.Website = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
v.Description.Details = r.GetString(1+int(r.GetUint()%(MaxStringLength-1)))
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
if length==0 {out.OperatorAddress = nil
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
codonEncodeByteSlice(w, v.DelegatorAddress[:])
codonEncodeByteSlice(w, v.ValidatorAddress[:])
EncodeDec(w, v.Shares)
} //End of EncodeDelegation

func DecodeDelegation(bz []byte) (Delegation, int, error) {
var err error
var length int
var v Delegation
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.DelegatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.ValidatorAddress, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Shares, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeDelegation

func RandDelegation(r RandSrc) Delegation {
var length int
var v Delegation
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.DelegatorAddress = r.GetBytes(length)
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v.ValidatorAddress = r.GetBytes(length)
v.Shares = RandDec(r)
return v
} //End of RandDelegation

func DeepCopyDelegation(in Delegation) (out Delegation) {
var length int
length = len(in.DelegatorAddress)
if length==0 {out.DelegatorAddress = nil
} else {
out.DelegatorAddress = make([]uint8, length)
}
copy(out.DelegatorAddress[:], in.DelegatorAddress[:])
length = len(in.ValidatorAddress)
if length==0 {out.ValidatorAddress = nil
} else {
out.ValidatorAddress = make([]uint8, length)
}
copy(out.ValidatorAddress[:], in.ValidatorAddress[:])
out.Shares = DeepCopyDec(in.Shares)
return
} //End of DeepCopyDelegation

// Non-Interface
func EncodeBondStatus(w *[]byte, v BondStatus) {
codonEncodeUint8(w, uint8(v))
} //End of EncodeBondStatus

func DecodeBondStatus(bz []byte) (BondStatus, int, error) {
var err error
var v BondStatus
var n int
var total int
v = BondStatus(codonDecodeUint8(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
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
codonEncodeUvarint(w, uint64(v.PreviousPeriod))
EncodeDec(w, v.Stake)
codonEncodeUvarint(w, uint64(v.Height))
} //End of EncodeDelegatorStartingInfo

func DecodeDelegatorStartingInfo(bz []byte) (DelegatorStartingInfo, int, error) {
var err error
var v DelegatorStartingInfo
var n int
var total int
v.PreviousPeriod = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Stake, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Height = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
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
codonEncodeVarint(w, int64(len(v.CumulativeRewardRatio)))
for _0:=0; _0<len(v.CumulativeRewardRatio); _0++ {
codonEncodeString(w, v.CumulativeRewardRatio[_0].Denom)
EncodeDec(w, v.CumulativeRewardRatio[_0].Amount)
// end of v.CumulativeRewardRatio[_0]
}
codonEncodeUint16(w, v.ReferenceCount)
} //End of EncodeValidatorHistoricalRewards

func DecodeValidatorHistoricalRewards(bz []byte) (ValidatorHistoricalRewards, int, error) {
var err error
var length int
var v ValidatorHistoricalRewards
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.CumulativeRewardRatio = nil
} else {
v.CumulativeRewardRatio = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.CumulativeRewardRatio[_0], n, err = DecodeDecCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.ReferenceCount = uint16(codonDecodeUint16(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeValidatorHistoricalRewards

func RandValidatorHistoricalRewards(r RandSrc) ValidatorHistoricalRewards {
var length int
var v ValidatorHistoricalRewards
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.CumulativeRewardRatio = nil
} else {
v.CumulativeRewardRatio = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.CumulativeRewardRatio[_0] = RandDecCoin(r)
}
v.ReferenceCount = r.GetUint16()
return v
} //End of RandValidatorHistoricalRewards

func DeepCopyValidatorHistoricalRewards(in ValidatorHistoricalRewards) (out ValidatorHistoricalRewards) {
var length int
length = len(in.CumulativeRewardRatio)
if length==0 {out.CumulativeRewardRatio = nil
} else {
out.CumulativeRewardRatio = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.CumulativeRewardRatio[_0] = DeepCopyDecCoin(in.CumulativeRewardRatio[_0])
}
out.ReferenceCount = in.ReferenceCount
return
} //End of DeepCopyValidatorHistoricalRewards

// Non-Interface
func EncodeValidatorCurrentRewards(w *[]byte, v ValidatorCurrentRewards) {
codonEncodeVarint(w, int64(len(v.Rewards)))
for _0:=0; _0<len(v.Rewards); _0++ {
codonEncodeString(w, v.Rewards[_0].Denom)
EncodeDec(w, v.Rewards[_0].Amount)
// end of v.Rewards[_0]
}
codonEncodeUvarint(w, uint64(v.Period))
} //End of EncodeValidatorCurrentRewards

func DecodeValidatorCurrentRewards(bz []byte) (ValidatorCurrentRewards, int, error) {
var err error
var length int
var v ValidatorCurrentRewards
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v.Rewards = nil
} else {
v.Rewards = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Rewards[_0], n, err = DecodeDecCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
v.Period = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeValidatorCurrentRewards

func RandValidatorCurrentRewards(r RandSrc) ValidatorCurrentRewards {
var length int
var v ValidatorCurrentRewards
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v.Rewards = nil
} else {
v.Rewards = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v.Rewards[_0] = RandDecCoin(r)
}
v.Period = r.GetUint64()
return v
} //End of RandValidatorCurrentRewards

func DeepCopyValidatorCurrentRewards(in ValidatorCurrentRewards) (out ValidatorCurrentRewards) {
var length int
length = len(in.Rewards)
if length==0 {out.Rewards = nil
} else {
out.Rewards = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out.Rewards[_0] = DeepCopyDecCoin(in.Rewards[_0])
}
out.Period = in.Period
return
} //End of DeepCopyValidatorCurrentRewards

// Non-Interface
func EncodeValidatorSigningInfo(w *[]byte, v ValidatorSigningInfo) {
codonEncodeByteSlice(w, v.Address[:])
codonEncodeVarint(w, int64(v.StartHeight))
codonEncodeVarint(w, int64(v.IndexOffset))
EncodeTime(w, v.JailedUntil)
codonEncodeBool(w, v.Tombstoned)
codonEncodeVarint(w, int64(v.MissedBlocksCounter))
} //End of EncodeValidatorSigningInfo

func DecodeValidatorSigningInfo(bz []byte) (ValidatorSigningInfo, int, error) {
var err error
var length int
var v ValidatorSigningInfo
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Address, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.StartHeight = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.IndexOffset = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.JailedUntil, n, err = DecodeTime(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Tombstoned = bool(codonDecodeBool(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.MissedBlocksCounter = int64(codonDecodeInt64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeValidatorSigningInfo

func RandValidatorSigningInfo(r RandSrc) ValidatorSigningInfo {
var length int
var v ValidatorSigningInfo
length = 1+int(r.GetUint()%(MaxSliceLength-1))
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
if length==0 {out.Address = nil
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
codonEncodeUvarint(w, uint64(v.ValidatorPeriod))
EncodeDec(w, v.Fraction)
} //End of EncodeValidatorSlashEvent

func DecodeValidatorSlashEvent(bz []byte) (ValidatorSlashEvent, int, error) {
var err error
var v ValidatorSlashEvent
var n int
var total int
v.ValidatorPeriod = uint64(codonDecodeUint64(bz, &n, &err))
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v.Fraction, n, err = DecodeDec(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
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
codonEncodeVarint(w, int64(len(v)))
for _0:=0; _0<len(v); _0++ {
codonEncodeString(w, v[_0].Denom)
EncodeDec(w, v[_0].Amount)
// end of v[_0]
}
} //End of EncodeDecCoins

func DecodeDecCoins(bz []byte) (DecCoins, int, error) {
var err error
var length int
var v DecCoins
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v = nil
} else {
v = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v[_0], n, err = DecodeDecCoin(bz)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeDecCoins

func RandDecCoins(r RandSrc) DecCoins {
var length int
var v DecCoins
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v = nil
} else {
v = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
v[_0] = RandDecCoin(r)
}
return v
} //End of RandDecCoins

func DeepCopyDecCoins(in DecCoins) (out DecCoins) {
var length int
length = len(in)
if length==0 {out = nil
} else {
out = make([]DecCoin, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of struct
out[_0] = DeepCopyDecCoin(in[_0])
}
return
} //End of DeepCopyDecCoins

// Non-Interface
func EncodeValAddress(w *[]byte, v ValAddress) {
codonEncodeByteSlice(w, v[:])
} //End of EncodeValAddress

func DecodeValAddress(bz []byte) (ValAddress, int, error) {
var err error
var length int
var v ValAddress
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v, n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
return v, total, nil
} //End of DecodeValAddress

func RandValAddress(r RandSrc) ValAddress {
var length int
var v ValAddress
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v = r.GetBytes(length)
return v
} //End of RandValAddress

func DeepCopyValAddress(in ValAddress) (out ValAddress) {
var length int
length = len(in)
if length==0 {out = nil
} else {
out = make([]uint8, length)
}
copy(out[:], in[:])
return
} //End of DeepCopyValAddress

// Non-Interface
func EncodeValAddressList(w *[]byte, v ValAddressList) {
codonEncodeVarint(w, int64(len(v)))
for _0:=0; _0<len(v); _0++ {
codonEncodeByteSlice(w, v[_0][:])
}
} //End of EncodeValAddressList

func DecodeValAddressList(bz []byte) (ValAddressList, int, error) {
var err error
var length int
var v ValAddressList
var n int
var total int
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
if length==0 {v = nil
} else {
v = make([]ValAddress, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of slice
length = codonDecodeInt(bz, &n, &err)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
v[_0], n, err = codonGetByteSlice(bz, length)
if err != nil {return v, total, err}
bz = bz[n:]
total+=n
}
return v, total, nil
} //End of DecodeValAddressList

func RandValAddressList(r RandSrc) ValAddressList {
var length int
var v ValAddressList
length = 1+int(r.GetUint()%(MaxSliceLength-1))
if length==0 {v = nil
} else {
v = make([]ValAddress, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of slice
length = 1+int(r.GetUint()%(MaxSliceLength-1))
v[_0] = r.GetBytes(length)
}
return v
} //End of RandValAddressList

func DeepCopyValAddressList(in ValAddressList) (out ValAddressList) {
var length int
length = len(in)
if length==0 {out = nil
} else {
out = make([]ValAddress, length)
}
for _0, length_0 := 0, length; _0<length_0; _0++ { //slice of slice
length = len(in[_0])
if length==0 {out[_0] = nil
} else {
out[_0] = make([]uint8, length)
}
copy(out[_0][:], in[_0][:])
}
return
} //End of DeepCopyValAddressList

// Interface
func DecodePubKey(bz []byte) (PubKey, int, error) {
var v PubKey
var magicBytes [4]byte
var n int
var err error
notNil := codonDecodeBool(bz, &n, &err)
if err != nil {return v, n, err}
bz = bz[n:]
if !notNil {
return nil, n, nil
}
for i:=0; i<4; i++ {magicBytes[i] = bz[i]}
switch magicBytes {
case [4]byte{114,76,37,23}:
v, n, err := DecodePubKeyEd25519(bz[4:])
return v, n+5, err
case [4]byte{14,33,23,141}:
v, n, err := DecodePubKeyMultisigThreshold(bz[4:])
return v, n+5, err
case [4]byte{51,161,20,197}:
v, n, err := DecodePubKeySecp256k1(bz[4:])
return v, n+5, err
case [4]byte{247,42,43,179}:
v, n, err := DecodeStdSignature(bz[4:])
return v, n+5, err
default:
panic("Unknown type")
} // end of switch
return v, n, nil
} // end of DecodePubKey
func EncodePubKey(w *[]byte, x interface{}) {
codonEncodeBool(w, x != nil)
if x == nil {return}

switch v := x.(type) {
case PubKeyEd25519:
*w = append(*w, getMagicBytes("PubKeyEd25519")...)
EncodePubKeyEd25519(w, v)
case *PubKeyEd25519:
*w = append(*w, getMagicBytes("PubKeyEd25519")...)
EncodePubKeyEd25519(w, *v)
case PubKeyMultisigThreshold:
*w = append(*w, getMagicBytes("PubKeyMultisigThreshold")...)
EncodePubKeyMultisigThreshold(w, v)
case *PubKeyMultisigThreshold:
*w = append(*w, getMagicBytes("PubKeyMultisigThreshold")...)
EncodePubKeyMultisigThreshold(w, *v)
case PubKeySecp256k1:
*w = append(*w, getMagicBytes("PubKeySecp256k1")...)
EncodePubKeySecp256k1(w, v)
case *PubKeySecp256k1:
*w = append(*w, getMagicBytes("PubKeySecp256k1")...)
EncodePubKeySecp256k1(w, *v)
case StdSignature:
*w = append(*w, getMagicBytes("StdSignature")...)
EncodeStdSignature(w, v)
case *StdSignature:
*w = append(*w, getMagicBytes("StdSignature")...)
EncodeStdSignature(w, *v)
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
func DecodePrivKey(bz []byte) (PrivKey, int, error) {
var v PrivKey
var magicBytes [4]byte
var n int
var err error
notNil := codonDecodeBool(bz, &n, &err)
if err != nil {return v, n, err}
bz = bz[n:]
if !notNil {
return nil, n, nil
}
for i:=0; i<4; i++ {magicBytes[i] = bz[i]}
switch magicBytes {
case [4]byte{158,94,112,161}:
v, n, err := DecodePrivKeyEd25519(bz[4:])
return v, n+5, err
case [4]byte{83,16,177,42}:
v, n, err := DecodePrivKeySecp256k1(bz[4:])
return v, n+5, err
default:
panic("Unknown type")
} // end of switch
return v, n, nil
} // end of DecodePrivKey
func EncodePrivKey(w *[]byte, x interface{}) {
codonEncodeBool(w, x != nil)
if x == nil {return}

switch v := x.(type) {
case PrivKeyEd25519:
*w = append(*w, getMagicBytes("PrivKeyEd25519")...)
EncodePrivKeyEd25519(w, v)
case *PrivKeyEd25519:
*w = append(*w, getMagicBytes("PrivKeyEd25519")...)
EncodePrivKeyEd25519(w, *v)
case PrivKeySecp256k1:
*w = append(*w, getMagicBytes("PrivKeySecp256k1")...)
EncodePrivKeySecp256k1(w, v)
case *PrivKeySecp256k1:
*w = append(*w, getMagicBytes("PrivKeySecp256k1")...)
EncodePrivKeySecp256k1(w, *v)
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
func DecodeMsg(bz []byte) (Msg, int, error) {
var v Msg
var magicBytes [4]byte
var n int
var err error
notNil := codonDecodeBool(bz, &n, &err)
if err != nil {return v, n, err}
bz = bz[n:]
if !notNil {
return nil, n, nil
}
for i:=0; i<4; i++ {magicBytes[i] = bz[i]}
switch magicBytes {
case [4]byte{141,7,107,68}:
v, n, err := DecodeMsgBeginRedelegate(bz[4:])
return v, n+5, err
case [4]byte{24,79,66,107}:
v, n, err := DecodeMsgCreateValidator(bz[4:])
return v, n+5, err
case [4]byte{184,121,196,185}:
v, n, err := DecodeMsgDelegate(bz[4:])
return v, n+5, err
case [4]byte{234,76,240,151}:
v, n, err := DecodeMsgDeposit(bz[4:])
return v, n+5, err
case [4]byte{9,254,168,109}:
v, n, err := DecodeMsgEditValidator(bz[4:])
return v, n+5, err
case [4]byte{64,119,59,163}:
v, n, err := DecodeMsgMultiSend(bz[4:])
return v, n+5, err
case [4]byte{212,255,125,220}:
v, n, err := DecodeMsgSend(bz[4:])
return v, n+5, err
case [4]byte{208,136,199,77}:
v, n, err := DecodeMsgSetWithdrawAddress(bz[4:])
return v, n+5, err
case [4]byte{21,241,6,56}:
v, n, err := DecodeMsgUndelegate(bz[4:])
return v, n+5, err
case [4]byte{139,110,39,159}:
v, n, err := DecodeMsgUnjail(bz[4:])
return v, n+5, err
case [4]byte{109,173,240,7}:
v, n, err := DecodeMsgVerifyInvariant(bz[4:])
return v, n+5, err
case [4]byte{233,121,28,250}:
v, n, err := DecodeMsgVote(bz[4:])
return v, n+5, err
case [4]byte{43,19,183,111}:
v, n, err := DecodeMsgWithdrawDelegatorReward(bz[4:])
return v, n+5, err
case [4]byte{84,85,236,88}:
v, n, err := DecodeMsgWithdrawValidatorCommission(bz[4:])
return v, n+5, err
default:
panic("Unknown type")
} // end of switch
return v, n, nil
} // end of DecodeMsg
func EncodeMsg(w *[]byte, x interface{}) {
codonEncodeBool(w, x != nil)
if x == nil {return}

switch v := x.(type) {
case MsgBeginRedelegate:
*w = append(*w, getMagicBytes("MsgBeginRedelegate")...)
EncodeMsgBeginRedelegate(w, v)
case *MsgBeginRedelegate:
*w = append(*w, getMagicBytes("MsgBeginRedelegate")...)
EncodeMsgBeginRedelegate(w, *v)
case MsgCreateValidator:
*w = append(*w, getMagicBytes("MsgCreateValidator")...)
EncodeMsgCreateValidator(w, v)
case *MsgCreateValidator:
*w = append(*w, getMagicBytes("MsgCreateValidator")...)
EncodeMsgCreateValidator(w, *v)
case MsgDelegate:
*w = append(*w, getMagicBytes("MsgDelegate")...)
EncodeMsgDelegate(w, v)
case *MsgDelegate:
*w = append(*w, getMagicBytes("MsgDelegate")...)
EncodeMsgDelegate(w, *v)
case MsgDeposit:
*w = append(*w, getMagicBytes("MsgDeposit")...)
EncodeMsgDeposit(w, v)
case *MsgDeposit:
*w = append(*w, getMagicBytes("MsgDeposit")...)
EncodeMsgDeposit(w, *v)
case MsgEditValidator:
*w = append(*w, getMagicBytes("MsgEditValidator")...)
EncodeMsgEditValidator(w, v)
case *MsgEditValidator:
*w = append(*w, getMagicBytes("MsgEditValidator")...)
EncodeMsgEditValidator(w, *v)
case MsgMultiSend:
*w = append(*w, getMagicBytes("MsgMultiSend")...)
EncodeMsgMultiSend(w, v)
case *MsgMultiSend:
*w = append(*w, getMagicBytes("MsgMultiSend")...)
EncodeMsgMultiSend(w, *v)
case MsgSend:
*w = append(*w, getMagicBytes("MsgSend")...)
EncodeMsgSend(w, v)
case *MsgSend:
*w = append(*w, getMagicBytes("MsgSend")...)
EncodeMsgSend(w, *v)
case MsgSetWithdrawAddress:
*w = append(*w, getMagicBytes("MsgSetWithdrawAddress")...)
EncodeMsgSetWithdrawAddress(w, v)
case *MsgSetWithdrawAddress:
*w = append(*w, getMagicBytes("MsgSetWithdrawAddress")...)
EncodeMsgSetWithdrawAddress(w, *v)
case MsgUndelegate:
*w = append(*w, getMagicBytes("MsgUndelegate")...)
EncodeMsgUndelegate(w, v)
case *MsgUndelegate:
*w = append(*w, getMagicBytes("MsgUndelegate")...)
EncodeMsgUndelegate(w, *v)
case MsgUnjail:
*w = append(*w, getMagicBytes("MsgUnjail")...)
EncodeMsgUnjail(w, v)
case *MsgUnjail:
*w = append(*w, getMagicBytes("MsgUnjail")...)
EncodeMsgUnjail(w, *v)
case MsgVerifyInvariant:
*w = append(*w, getMagicBytes("MsgVerifyInvariant")...)
EncodeMsgVerifyInvariant(w, v)
case *MsgVerifyInvariant:
*w = append(*w, getMagicBytes("MsgVerifyInvariant")...)
EncodeMsgVerifyInvariant(w, *v)
case MsgVote:
*w = append(*w, getMagicBytes("MsgVote")...)
EncodeMsgVote(w, v)
case *MsgVote:
*w = append(*w, getMagicBytes("MsgVote")...)
EncodeMsgVote(w, *v)
case MsgWithdrawDelegatorReward:
*w = append(*w, getMagicBytes("MsgWithdrawDelegatorReward")...)
EncodeMsgWithdrawDelegatorReward(w, v)
case *MsgWithdrawDelegatorReward:
*w = append(*w, getMagicBytes("MsgWithdrawDelegatorReward")...)
EncodeMsgWithdrawDelegatorReward(w, *v)
case MsgWithdrawValidatorCommission:
*w = append(*w, getMagicBytes("MsgWithdrawValidatorCommission")...)
EncodeMsgWithdrawValidatorCommission(w, v)
case *MsgWithdrawValidatorCommission:
*w = append(*w, getMagicBytes("MsgWithdrawValidatorCommission")...)
EncodeMsgWithdrawValidatorCommission(w, *v)
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
func DecodeAccount(bz []byte) (Account, int, error) {
var v Account
var magicBytes [4]byte
var n int
var err error
notNil := codonDecodeBool(bz, &n, &err)
if err != nil {return v, n, err}
bz = bz[n:]
if !notNil {
return nil, n, nil
}
for i:=0; i<4; i++ {magicBytes[i] = bz[i]}
switch magicBytes {
case [4]byte{153,157,134,34}:
v, n, err := DecodeBaseAccount(bz[4:])
return &v, n+5, err
case [4]byte{78,248,144,54}:
v, n, err := DecodeBaseVestingAccount(bz[4:])
return v, n+5, err
case [4]byte{75,69,41,151}:
v, n, err := DecodeContinuousVestingAccount(bz[4:])
return v, n+5, err
case [4]byte{59,193,203,230}:
v, n, err := DecodeDelayedVestingAccount(bz[4:])
return v, n+5, err
case [4]byte{37,29,227,212}:
v, n, err := DecodeModuleAccount(bz[4:])
return v, n+5, err
default:
panic("Unknown type")
} // end of switch
return v, n, nil
} // end of DecodeAccount
func EncodeAccount(w *[]byte, x interface{}) {
codonEncodeBool(w, x != nil)
if x == nil {return}

switch v := x.(type) {
case BaseAccount:
*w = append(*w, getMagicBytes("BaseAccount")...)
EncodeBaseAccount(w, v)
case *BaseAccount:
*w = append(*w, getMagicBytes("BaseAccount")...)
EncodeBaseAccount(w, *v)
case BaseVestingAccount:
*w = append(*w, getMagicBytes("BaseVestingAccount")...)
EncodeBaseVestingAccount(w, v)
case *BaseVestingAccount:
*w = append(*w, getMagicBytes("BaseVestingAccount")...)
EncodeBaseVestingAccount(w, *v)
case ContinuousVestingAccount:
*w = append(*w, getMagicBytes("ContinuousVestingAccount")...)
EncodeContinuousVestingAccount(w, v)
case *ContinuousVestingAccount:
*w = append(*w, getMagicBytes("ContinuousVestingAccount")...)
EncodeContinuousVestingAccount(w, *v)
case DelayedVestingAccount:
*w = append(*w, getMagicBytes("DelayedVestingAccount")...)
EncodeDelayedVestingAccount(w, v)
case *DelayedVestingAccount:
*w = append(*w, getMagicBytes("DelayedVestingAccount")...)
EncodeDelayedVestingAccount(w, *v)
case ModuleAccount:
*w = append(*w, getMagicBytes("ModuleAccount")...)
EncodeModuleAccount(w, v)
case *ModuleAccount:
*w = append(*w, getMagicBytes("ModuleAccount")...)
EncodeModuleAccount(w, *v)
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
func DecodeVestingAccount(bz []byte) (VestingAccount, int, error) {
var v VestingAccount
var magicBytes [4]byte
var n int
var err error
notNil := codonDecodeBool(bz, &n, &err)
if err != nil {return v, n, err}
bz = bz[n:]
if !notNil {
return nil, n, nil
}
for i:=0; i<4; i++ {magicBytes[i] = bz[i]}
switch magicBytes {
case [4]byte{75,69,41,151}:
v, n, err := DecodeContinuousVestingAccount(bz[4:])
return &v, n+5, err
case [4]byte{59,193,203,230}:
v, n, err := DecodeDelayedVestingAccount(bz[4:])
return &v, n+5, err
default:
panic("Unknown type")
} // end of switch
return v, n, nil
} // end of DecodeVestingAccount
func EncodeVestingAccount(w *[]byte, x interface{}) {
codonEncodeBool(w, x != nil)
if x == nil {return}

switch v := x.(type) {
case ContinuousVestingAccount:
*w = append(*w, getMagicBytes("ContinuousVestingAccount")...)
EncodeContinuousVestingAccount(w, v)
case *ContinuousVestingAccount:
*w = append(*w, getMagicBytes("ContinuousVestingAccount")...)
EncodeContinuousVestingAccount(w, *v)
case DelayedVestingAccount:
*w = append(*w, getMagicBytes("DelayedVestingAccount")...)
EncodeDelayedVestingAccount(w, v)
case *DelayedVestingAccount:
*w = append(*w, getMagicBytes("DelayedVestingAccount")...)
EncodeDelayedVestingAccount(w, *v)
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
func DecodeContent(bz []byte) (Content, int, error) {
var v Content
var magicBytes [4]byte
var n int
var err error
notNil := codonDecodeBool(bz, &n, &err)
if err != nil {return v, n, err}
bz = bz[n:]
if !notNil {
return nil, n, nil
}
for i:=0; i<4; i++ {magicBytes[i] = bz[i]}
switch magicBytes {
case [4]byte{31,93,37,208}:
v, n, err := DecodeCommunityPoolSpendProposal(bz[4:])
return v, n+5, err
case [4]byte{49,37,122,86}:
v, n, err := DecodeParameterChangeProposal(bz[4:])
return v, n+5, err
case [4]byte{162,148,222,207}:
v, n, err := DecodeSoftwareUpgradeProposal(bz[4:])
return v, n+5, err
case [4]byte{207,179,211,152}:
v, n, err := DecodeTextProposal(bz[4:])
return v, n+5, err
default:
panic("Unknown type")
} // end of switch
return v, n, nil
} // end of DecodeContent
func EncodeContent(w *[]byte, x interface{}) {
codonEncodeBool(w, x != nil)
if x == nil {return}

switch v := x.(type) {
case CommunityPoolSpendProposal:
*w = append(*w, getMagicBytes("CommunityPoolSpendProposal")...)
EncodeCommunityPoolSpendProposal(w, v)
case *CommunityPoolSpendProposal:
*w = append(*w, getMagicBytes("CommunityPoolSpendProposal")...)
EncodeCommunityPoolSpendProposal(w, *v)
case ParameterChangeProposal:
*w = append(*w, getMagicBytes("ParameterChangeProposal")...)
EncodeParameterChangeProposal(w, v)
case *ParameterChangeProposal:
*w = append(*w, getMagicBytes("ParameterChangeProposal")...)
EncodeParameterChangeProposal(w, *v)
case SoftwareUpgradeProposal:
*w = append(*w, getMagicBytes("SoftwareUpgradeProposal")...)
EncodeSoftwareUpgradeProposal(w, v)
case *SoftwareUpgradeProposal:
*w = append(*w, getMagicBytes("SoftwareUpgradeProposal")...)
EncodeSoftwareUpgradeProposal(w, *v)
case TextProposal:
*w = append(*w, getMagicBytes("TextProposal")...)
EncodeTextProposal(w, v)
case *TextProposal:
*w = append(*w, getMagicBytes("TextProposal")...)
EncodeTextProposal(w, *v)
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
func DecodeTx(bz []byte) (Tx, int, error) {
var v Tx
var magicBytes [4]byte
var n int
var err error
notNil := codonDecodeBool(bz, &n, &err)
if err != nil {return v, n, err}
bz = bz[n:]
if !notNil {
return nil, n, nil
}
for i:=0; i<4; i++ {magicBytes[i] = bz[i]}
switch magicBytes {
case [4]byte{247,170,118,185}:
v, n, err := DecodeStdTx(bz[4:])
return v, n+5, err
default:
panic("Unknown type")
} // end of switch
return v, n, nil
} // end of DecodeTx
func EncodeTx(w *[]byte, x interface{}) {
codonEncodeBool(w, x != nil)
if x == nil {return}

switch v := x.(type) {
case StdTx:
*w = append(*w, getMagicBytes("StdTx")...)
EncodeStdTx(w, v)
case *StdTx:
*w = append(*w, getMagicBytes("StdTx")...)
EncodeStdTx(w, *v)
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
func DecodeModuleAccountI(bz []byte) (ModuleAccountI, int, error) {
var v ModuleAccountI
var magicBytes [4]byte
var n int
var err error
notNil := codonDecodeBool(bz, &n, &err)
if err != nil {return v, n, err}
bz = bz[n:]
if !notNil {
return nil, n, nil
}
for i:=0; i<4; i++ {magicBytes[i] = bz[i]}
switch magicBytes {
case [4]byte{37,29,227,212}:
v, n, err := DecodeModuleAccount(bz[4:])
return v, n+5, err
default:
panic("Unknown type")
} // end of switch
return v, n, nil
} // end of DecodeModuleAccountI
func EncodeModuleAccountI(w *[]byte, x interface{}) {
codonEncodeBool(w, x != nil)
if x == nil {return}

switch v := x.(type) {
case ModuleAccount:
*w = append(*w, getMagicBytes("ModuleAccount")...)
EncodeModuleAccount(w, v)
case *ModuleAccount:
*w = append(*w, getMagicBytes("ModuleAccount")...)
EncodeModuleAccount(w, *v)
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
func DecodeSupplyI(bz []byte) (SupplyI, int, error) {
var v SupplyI
var magicBytes [4]byte
var n int
var err error
notNil := codonDecodeBool(bz, &n, &err)
if err != nil {return v, n, err}
bz = bz[n:]
if !notNil {
return nil, n, nil
}
for i:=0; i<4; i++ {magicBytes[i] = bz[i]}
switch magicBytes {
case [4]byte{191,66,141,63}:
v, n, err := DecodeSupply(bz[4:])
return v, n+5, err
default:
panic("Unknown type")
} // end of switch
return v, n, nil
} // end of DecodeSupplyI
func EncodeSupplyI(w *[]byte, x interface{}) {
codonEncodeBool(w, x != nil)
if x == nil {return}

switch v := x.(type) {
case Supply:
*w = append(*w, getMagicBytes("Supply")...)
EncodeSupply(w, v)
case *Supply:
*w = append(*w, getMagicBytes("Supply")...)
EncodeSupply(w, *v)
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
func getMagicBytes(name string) []byte {
switch name {
case "AccAddress":
return []byte{0,157,18,162}
case "AccAddressList":
return []byte{37,72,3,140}
case "BaseAccount":
return []byte{153,157,134,34}
case "BaseVestingAccount":
return []byte{78,248,144,54}
case "BondStatus":
return []byte{22,131,204,195}
case "Coin":
return []byte{2,65,204,255}
case "CommitInfo":
return []byte{2,26,137,96}
case "CommunityPoolSpendProposal":
return []byte{31,93,37,208}
case "ConsAddress":
return []byte{28,53,138,173}
case "ContinuousVestingAccount":
return []byte{75,69,41,151}
case "DecCoin":
return []byte{56,163,255,105}
case "DecCoins":
return []byte{57,34,102,152}
case "DelayedVestingAccount":
return []byte{59,193,203,230}
case "Delegation":
return []byte{166,217,16,186}
case "DelegatorStartingInfo":
return []byte{191,63,17,72}
case "FeePool":
return []byte{18,193,65,104}
case "Input":
return []byte{54,236,180,248}
case "ModuleAccount":
return []byte{37,29,227,212}
case "MsgBeginRedelegate":
return []byte{141,7,107,68}
case "MsgCreateValidator":
return []byte{24,79,66,107}
case "MsgDelegate":
return []byte{184,121,196,185}
case "MsgDeposit":
return []byte{234,76,240,151}
case "MsgEditValidator":
return []byte{9,254,168,109}
case "MsgMultiSend":
return []byte{64,119,59,163}
case "MsgSend":
return []byte{212,255,125,220}
case "MsgSetWithdrawAddress":
return []byte{208,136,199,77}
case "MsgUndelegate":
return []byte{21,241,6,56}
case "MsgUnjail":
return []byte{139,110,39,159}
case "MsgVerifyInvariant":
return []byte{109,173,240,7}
case "MsgVote":
return []byte{233,121,28,250}
case "MsgWithdrawDelegatorReward":
return []byte{43,19,183,111}
case "MsgWithdrawValidatorCommission":
return []byte{84,85,236,88}
case "Output":
return []byte{178,67,155,203}
case "ParamChange":
return []byte{66,250,248,208}
case "ParameterChangeProposal":
return []byte{49,37,122,86}
case "PrivKeyEd25519":
return []byte{158,94,112,161}
case "PrivKeySecp256k1":
return []byte{83,16,177,42}
case "PubKeyEd25519":
return []byte{114,76,37,23}
case "PubKeyMultisigThreshold":
return []byte{14,33,23,141}
case "PubKeySecp256k1":
return []byte{51,161,20,197}
case "SdkDec":
return []byte{131,101,23,4}
case "SdkInt":
return []byte{189,210,54,221}
case "SignedMsgType":
return []byte{67,52,162,78}
case "SoftwareUpgradeProposal":
return []byte{162,148,222,207}
case "StdSignature":
return []byte{247,42,43,179}
case "StdTx":
return []byte{247,170,118,185}
case "StoreInfo":
return []byte{224,49,135,8}
case "Supply":
return []byte{191,66,141,63}
case "TextProposal":
return []byte{207,179,211,152}
case "ValAddress":
return []byte{98,39,43,120}
case "ValAddressList":
return []byte{11,142,217,17}
case "Validator":
return []byte{11,10,9,103}
case "ValidatorCurrentRewards":
return []byte{178,20,101,34}
case "ValidatorHistoricalRewards":
return []byte{112,41,104,134}
case "ValidatorSigningInfo":
return []byte{176,207,98,135}
case "ValidatorSlashEvent":
return []byte{76,125,139,93}
case "Vote":
return []byte{205,85,136,219}
case "VoteOption":
return []byte{170,208,50,2}
case "int64":
return []byte{188,34,41,102}
case "uint64":
return []byte{36,210,58,112}
} // end of switch
panic("Should not reach here")
return []byte{}
} // end of getMagicBytes
func getMagicBytesOfVar(x interface{}) ([4]byte, error) {
switch x.(type) {
case *AccAddress, AccAddress:
return [4]byte{0,157,18,162}, nil
case *AccAddressList, AccAddressList:
return [4]byte{37,72,3,140}, nil
case *BaseAccount, BaseAccount:
return [4]byte{153,157,134,34}, nil
case *BaseVestingAccount, BaseVestingAccount:
return [4]byte{78,248,144,54}, nil
case *BondStatus, BondStatus:
return [4]byte{22,131,204,195}, nil
case *Coin, Coin:
return [4]byte{2,65,204,255}, nil
case *CommitInfo, CommitInfo:
return [4]byte{2,26,137,96}, nil
case *CommunityPoolSpendProposal, CommunityPoolSpendProposal:
return [4]byte{31,93,37,208}, nil
case *ConsAddress, ConsAddress:
return [4]byte{28,53,138,173}, nil
case *ContinuousVestingAccount, ContinuousVestingAccount:
return [4]byte{75,69,41,151}, nil
case *DecCoin, DecCoin:
return [4]byte{56,163,255,105}, nil
case *DecCoins, DecCoins:
return [4]byte{57,34,102,152}, nil
case *DelayedVestingAccount, DelayedVestingAccount:
return [4]byte{59,193,203,230}, nil
case *Delegation, Delegation:
return [4]byte{166,217,16,186}, nil
case *DelegatorStartingInfo, DelegatorStartingInfo:
return [4]byte{191,63,17,72}, nil
case *FeePool, FeePool:
return [4]byte{18,193,65,104}, nil
case *Input, Input:
return [4]byte{54,236,180,248}, nil
case *ModuleAccount, ModuleAccount:
return [4]byte{37,29,227,212}, nil
case *MsgBeginRedelegate, MsgBeginRedelegate:
return [4]byte{141,7,107,68}, nil
case *MsgCreateValidator, MsgCreateValidator:
return [4]byte{24,79,66,107}, nil
case *MsgDelegate, MsgDelegate:
return [4]byte{184,121,196,185}, nil
case *MsgDeposit, MsgDeposit:
return [4]byte{234,76,240,151}, nil
case *MsgEditValidator, MsgEditValidator:
return [4]byte{9,254,168,109}, nil
case *MsgMultiSend, MsgMultiSend:
return [4]byte{64,119,59,163}, nil
case *MsgSend, MsgSend:
return [4]byte{212,255,125,220}, nil
case *MsgSetWithdrawAddress, MsgSetWithdrawAddress:
return [4]byte{208,136,199,77}, nil
case *MsgUndelegate, MsgUndelegate:
return [4]byte{21,241,6,56}, nil
case *MsgUnjail, MsgUnjail:
return [4]byte{139,110,39,159}, nil
case *MsgVerifyInvariant, MsgVerifyInvariant:
return [4]byte{109,173,240,7}, nil
case *MsgVote, MsgVote:
return [4]byte{233,121,28,250}, nil
case *MsgWithdrawDelegatorReward, MsgWithdrawDelegatorReward:
return [4]byte{43,19,183,111}, nil
case *MsgWithdrawValidatorCommission, MsgWithdrawValidatorCommission:
return [4]byte{84,85,236,88}, nil
case *Output, Output:
return [4]byte{178,67,155,203}, nil
case *ParamChange, ParamChange:
return [4]byte{66,250,248,208}, nil
case *ParameterChangeProposal, ParameterChangeProposal:
return [4]byte{49,37,122,86}, nil
case *PrivKeyEd25519, PrivKeyEd25519:
return [4]byte{158,94,112,161}, nil
case *PrivKeySecp256k1, PrivKeySecp256k1:
return [4]byte{83,16,177,42}, nil
case *PubKeyEd25519, PubKeyEd25519:
return [4]byte{114,76,37,23}, nil
case *PubKeyMultisigThreshold, PubKeyMultisigThreshold:
return [4]byte{14,33,23,141}, nil
case *PubKeySecp256k1, PubKeySecp256k1:
return [4]byte{51,161,20,197}, nil
case *SdkDec, SdkDec:
return [4]byte{131,101,23,4}, nil
case *SdkInt, SdkInt:
return [4]byte{189,210,54,221}, nil
case *SignedMsgType, SignedMsgType:
return [4]byte{67,52,162,78}, nil
case *SoftwareUpgradeProposal, SoftwareUpgradeProposal:
return [4]byte{162,148,222,207}, nil
case *StdSignature, StdSignature:
return [4]byte{247,42,43,179}, nil
case *StdTx, StdTx:
return [4]byte{247,170,118,185}, nil
case *StoreInfo, StoreInfo:
return [4]byte{224,49,135,8}, nil
case *Supply, Supply:
return [4]byte{191,66,141,63}, nil
case *TextProposal, TextProposal:
return [4]byte{207,179,211,152}, nil
case *ValAddress, ValAddress:
return [4]byte{98,39,43,120}, nil
case *ValAddressList, ValAddressList:
return [4]byte{11,142,217,17}, nil
case *Validator, Validator:
return [4]byte{11,10,9,103}, nil
case *ValidatorCurrentRewards, ValidatorCurrentRewards:
return [4]byte{178,20,101,34}, nil
case *ValidatorHistoricalRewards, ValidatorHistoricalRewards:
return [4]byte{112,41,104,134}, nil
case *ValidatorSigningInfo, ValidatorSigningInfo:
return [4]byte{176,207,98,135}, nil
case *ValidatorSlashEvent, ValidatorSlashEvent:
return [4]byte{76,125,139,93}, nil
case *Vote, Vote:
return [4]byte{205,85,136,219}, nil
case *VoteOption, VoteOption:
return [4]byte{170,208,50,2}, nil
case *int64, int64:
return [4]byte{188,34,41,102}, nil
case *uint64, uint64:
return [4]byte{36,210,58,112}, nil
default:
panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
} // end of switch
} // end of func
func EncodeAny(w *[]byte, x interface{}) {
switch v := x.(type) {
case AccAddress:
*w = append(*w, getMagicBytes("AccAddress")...)
EncodeAccAddress(w, v)
case *AccAddress:
*w = append(*w, getMagicBytes("AccAddress")...)
EncodeAccAddress(w, *v)
case AccAddressList:
*w = append(*w, getMagicBytes("AccAddressList")...)
EncodeAccAddressList(w, v)
case *AccAddressList:
*w = append(*w, getMagicBytes("AccAddressList")...)
EncodeAccAddressList(w, *v)
case BaseAccount:
*w = append(*w, getMagicBytes("BaseAccount")...)
EncodeBaseAccount(w, v)
case *BaseAccount:
*w = append(*w, getMagicBytes("BaseAccount")...)
EncodeBaseAccount(w, *v)
case BaseVestingAccount:
*w = append(*w, getMagicBytes("BaseVestingAccount")...)
EncodeBaseVestingAccount(w, v)
case *BaseVestingAccount:
*w = append(*w, getMagicBytes("BaseVestingAccount")...)
EncodeBaseVestingAccount(w, *v)
case BondStatus:
*w = append(*w, getMagicBytes("BondStatus")...)
EncodeBondStatus(w, v)
case *BondStatus:
*w = append(*w, getMagicBytes("BondStatus")...)
EncodeBondStatus(w, *v)
case Coin:
*w = append(*w, getMagicBytes("Coin")...)
EncodeCoin(w, v)
case *Coin:
*w = append(*w, getMagicBytes("Coin")...)
EncodeCoin(w, *v)
case CommitInfo:
*w = append(*w, getMagicBytes("CommitInfo")...)
EncodeCommitInfo(w, v)
case *CommitInfo:
*w = append(*w, getMagicBytes("CommitInfo")...)
EncodeCommitInfo(w, *v)
case CommunityPoolSpendProposal:
*w = append(*w, getMagicBytes("CommunityPoolSpendProposal")...)
EncodeCommunityPoolSpendProposal(w, v)
case *CommunityPoolSpendProposal:
*w = append(*w, getMagicBytes("CommunityPoolSpendProposal")...)
EncodeCommunityPoolSpendProposal(w, *v)
case ConsAddress:
*w = append(*w, getMagicBytes("ConsAddress")...)
EncodeConsAddress(w, v)
case *ConsAddress:
*w = append(*w, getMagicBytes("ConsAddress")...)
EncodeConsAddress(w, *v)
case ContinuousVestingAccount:
*w = append(*w, getMagicBytes("ContinuousVestingAccount")...)
EncodeContinuousVestingAccount(w, v)
case *ContinuousVestingAccount:
*w = append(*w, getMagicBytes("ContinuousVestingAccount")...)
EncodeContinuousVestingAccount(w, *v)
case DecCoin:
*w = append(*w, getMagicBytes("DecCoin")...)
EncodeDecCoin(w, v)
case *DecCoin:
*w = append(*w, getMagicBytes("DecCoin")...)
EncodeDecCoin(w, *v)
case DecCoins:
*w = append(*w, getMagicBytes("DecCoins")...)
EncodeDecCoins(w, v)
case *DecCoins:
*w = append(*w, getMagicBytes("DecCoins")...)
EncodeDecCoins(w, *v)
case DelayedVestingAccount:
*w = append(*w, getMagicBytes("DelayedVestingAccount")...)
EncodeDelayedVestingAccount(w, v)
case *DelayedVestingAccount:
*w = append(*w, getMagicBytes("DelayedVestingAccount")...)
EncodeDelayedVestingAccount(w, *v)
case Delegation:
*w = append(*w, getMagicBytes("Delegation")...)
EncodeDelegation(w, v)
case *Delegation:
*w = append(*w, getMagicBytes("Delegation")...)
EncodeDelegation(w, *v)
case DelegatorStartingInfo:
*w = append(*w, getMagicBytes("DelegatorStartingInfo")...)
EncodeDelegatorStartingInfo(w, v)
case *DelegatorStartingInfo:
*w = append(*w, getMagicBytes("DelegatorStartingInfo")...)
EncodeDelegatorStartingInfo(w, *v)
case FeePool:
*w = append(*w, getMagicBytes("FeePool")...)
EncodeFeePool(w, v)
case *FeePool:
*w = append(*w, getMagicBytes("FeePool")...)
EncodeFeePool(w, *v)
case Input:
*w = append(*w, getMagicBytes("Input")...)
EncodeInput(w, v)
case *Input:
*w = append(*w, getMagicBytes("Input")...)
EncodeInput(w, *v)
case ModuleAccount:
*w = append(*w, getMagicBytes("ModuleAccount")...)
EncodeModuleAccount(w, v)
case *ModuleAccount:
*w = append(*w, getMagicBytes("ModuleAccount")...)
EncodeModuleAccount(w, *v)
case MsgBeginRedelegate:
*w = append(*w, getMagicBytes("MsgBeginRedelegate")...)
EncodeMsgBeginRedelegate(w, v)
case *MsgBeginRedelegate:
*w = append(*w, getMagicBytes("MsgBeginRedelegate")...)
EncodeMsgBeginRedelegate(w, *v)
case MsgCreateValidator:
*w = append(*w, getMagicBytes("MsgCreateValidator")...)
EncodeMsgCreateValidator(w, v)
case *MsgCreateValidator:
*w = append(*w, getMagicBytes("MsgCreateValidator")...)
EncodeMsgCreateValidator(w, *v)
case MsgDelegate:
*w = append(*w, getMagicBytes("MsgDelegate")...)
EncodeMsgDelegate(w, v)
case *MsgDelegate:
*w = append(*w, getMagicBytes("MsgDelegate")...)
EncodeMsgDelegate(w, *v)
case MsgDeposit:
*w = append(*w, getMagicBytes("MsgDeposit")...)
EncodeMsgDeposit(w, v)
case *MsgDeposit:
*w = append(*w, getMagicBytes("MsgDeposit")...)
EncodeMsgDeposit(w, *v)
case MsgEditValidator:
*w = append(*w, getMagicBytes("MsgEditValidator")...)
EncodeMsgEditValidator(w, v)
case *MsgEditValidator:
*w = append(*w, getMagicBytes("MsgEditValidator")...)
EncodeMsgEditValidator(w, *v)
case MsgMultiSend:
*w = append(*w, getMagicBytes("MsgMultiSend")...)
EncodeMsgMultiSend(w, v)
case *MsgMultiSend:
*w = append(*w, getMagicBytes("MsgMultiSend")...)
EncodeMsgMultiSend(w, *v)
case MsgSend:
*w = append(*w, getMagicBytes("MsgSend")...)
EncodeMsgSend(w, v)
case *MsgSend:
*w = append(*w, getMagicBytes("MsgSend")...)
EncodeMsgSend(w, *v)
case MsgSetWithdrawAddress:
*w = append(*w, getMagicBytes("MsgSetWithdrawAddress")...)
EncodeMsgSetWithdrawAddress(w, v)
case *MsgSetWithdrawAddress:
*w = append(*w, getMagicBytes("MsgSetWithdrawAddress")...)
EncodeMsgSetWithdrawAddress(w, *v)
case MsgUndelegate:
*w = append(*w, getMagicBytes("MsgUndelegate")...)
EncodeMsgUndelegate(w, v)
case *MsgUndelegate:
*w = append(*w, getMagicBytes("MsgUndelegate")...)
EncodeMsgUndelegate(w, *v)
case MsgUnjail:
*w = append(*w, getMagicBytes("MsgUnjail")...)
EncodeMsgUnjail(w, v)
case *MsgUnjail:
*w = append(*w, getMagicBytes("MsgUnjail")...)
EncodeMsgUnjail(w, *v)
case MsgVerifyInvariant:
*w = append(*w, getMagicBytes("MsgVerifyInvariant")...)
EncodeMsgVerifyInvariant(w, v)
case *MsgVerifyInvariant:
*w = append(*w, getMagicBytes("MsgVerifyInvariant")...)
EncodeMsgVerifyInvariant(w, *v)
case MsgVote:
*w = append(*w, getMagicBytes("MsgVote")...)
EncodeMsgVote(w, v)
case *MsgVote:
*w = append(*w, getMagicBytes("MsgVote")...)
EncodeMsgVote(w, *v)
case MsgWithdrawDelegatorReward:
*w = append(*w, getMagicBytes("MsgWithdrawDelegatorReward")...)
EncodeMsgWithdrawDelegatorReward(w, v)
case *MsgWithdrawDelegatorReward:
*w = append(*w, getMagicBytes("MsgWithdrawDelegatorReward")...)
EncodeMsgWithdrawDelegatorReward(w, *v)
case MsgWithdrawValidatorCommission:
*w = append(*w, getMagicBytes("MsgWithdrawValidatorCommission")...)
EncodeMsgWithdrawValidatorCommission(w, v)
case *MsgWithdrawValidatorCommission:
*w = append(*w, getMagicBytes("MsgWithdrawValidatorCommission")...)
EncodeMsgWithdrawValidatorCommission(w, *v)
case Output:
*w = append(*w, getMagicBytes("Output")...)
EncodeOutput(w, v)
case *Output:
*w = append(*w, getMagicBytes("Output")...)
EncodeOutput(w, *v)
case ParamChange:
*w = append(*w, getMagicBytes("ParamChange")...)
EncodeParamChange(w, v)
case *ParamChange:
*w = append(*w, getMagicBytes("ParamChange")...)
EncodeParamChange(w, *v)
case ParameterChangeProposal:
*w = append(*w, getMagicBytes("ParameterChangeProposal")...)
EncodeParameterChangeProposal(w, v)
case *ParameterChangeProposal:
*w = append(*w, getMagicBytes("ParameterChangeProposal")...)
EncodeParameterChangeProposal(w, *v)
case PrivKeyEd25519:
*w = append(*w, getMagicBytes("PrivKeyEd25519")...)
EncodePrivKeyEd25519(w, v)
case *PrivKeyEd25519:
*w = append(*w, getMagicBytes("PrivKeyEd25519")...)
EncodePrivKeyEd25519(w, *v)
case PrivKeySecp256k1:
*w = append(*w, getMagicBytes("PrivKeySecp256k1")...)
EncodePrivKeySecp256k1(w, v)
case *PrivKeySecp256k1:
*w = append(*w, getMagicBytes("PrivKeySecp256k1")...)
EncodePrivKeySecp256k1(w, *v)
case PubKeyEd25519:
*w = append(*w, getMagicBytes("PubKeyEd25519")...)
EncodePubKeyEd25519(w, v)
case *PubKeyEd25519:
*w = append(*w, getMagicBytes("PubKeyEd25519")...)
EncodePubKeyEd25519(w, *v)
case PubKeyMultisigThreshold:
*w = append(*w, getMagicBytes("PubKeyMultisigThreshold")...)
EncodePubKeyMultisigThreshold(w, v)
case *PubKeyMultisigThreshold:
*w = append(*w, getMagicBytes("PubKeyMultisigThreshold")...)
EncodePubKeyMultisigThreshold(w, *v)
case PubKeySecp256k1:
*w = append(*w, getMagicBytes("PubKeySecp256k1")...)
EncodePubKeySecp256k1(w, v)
case *PubKeySecp256k1:
*w = append(*w, getMagicBytes("PubKeySecp256k1")...)
EncodePubKeySecp256k1(w, *v)
case SdkDec:
*w = append(*w, getMagicBytes("SdkDec")...)
EncodeSdkDec(w, v)
case *SdkDec:
*w = append(*w, getMagicBytes("SdkDec")...)
EncodeSdkDec(w, *v)
case SdkInt:
*w = append(*w, getMagicBytes("SdkInt")...)
EncodeSdkInt(w, v)
case *SdkInt:
*w = append(*w, getMagicBytes("SdkInt")...)
EncodeSdkInt(w, *v)
case SignedMsgType:
*w = append(*w, getMagicBytes("SignedMsgType")...)
EncodeSignedMsgType(w, v)
case *SignedMsgType:
*w = append(*w, getMagicBytes("SignedMsgType")...)
EncodeSignedMsgType(w, *v)
case SoftwareUpgradeProposal:
*w = append(*w, getMagicBytes("SoftwareUpgradeProposal")...)
EncodeSoftwareUpgradeProposal(w, v)
case *SoftwareUpgradeProposal:
*w = append(*w, getMagicBytes("SoftwareUpgradeProposal")...)
EncodeSoftwareUpgradeProposal(w, *v)
case StdSignature:
*w = append(*w, getMagicBytes("StdSignature")...)
EncodeStdSignature(w, v)
case *StdSignature:
*w = append(*w, getMagicBytes("StdSignature")...)
EncodeStdSignature(w, *v)
case StdTx:
*w = append(*w, getMagicBytes("StdTx")...)
EncodeStdTx(w, v)
case *StdTx:
*w = append(*w, getMagicBytes("StdTx")...)
EncodeStdTx(w, *v)
case StoreInfo:
*w = append(*w, getMagicBytes("StoreInfo")...)
EncodeStoreInfo(w, v)
case *StoreInfo:
*w = append(*w, getMagicBytes("StoreInfo")...)
EncodeStoreInfo(w, *v)
case Supply:
*w = append(*w, getMagicBytes("Supply")...)
EncodeSupply(w, v)
case *Supply:
*w = append(*w, getMagicBytes("Supply")...)
EncodeSupply(w, *v)
case TextProposal:
*w = append(*w, getMagicBytes("TextProposal")...)
EncodeTextProposal(w, v)
case *TextProposal:
*w = append(*w, getMagicBytes("TextProposal")...)
EncodeTextProposal(w, *v)
case ValAddress:
*w = append(*w, getMagicBytes("ValAddress")...)
EncodeValAddress(w, v)
case *ValAddress:
*w = append(*w, getMagicBytes("ValAddress")...)
EncodeValAddress(w, *v)
case ValAddressList:
*w = append(*w, getMagicBytes("ValAddressList")...)
EncodeValAddressList(w, v)
case *ValAddressList:
*w = append(*w, getMagicBytes("ValAddressList")...)
EncodeValAddressList(w, *v)
case Validator:
*w = append(*w, getMagicBytes("Validator")...)
EncodeValidator(w, v)
case *Validator:
*w = append(*w, getMagicBytes("Validator")...)
EncodeValidator(w, *v)
case ValidatorCurrentRewards:
*w = append(*w, getMagicBytes("ValidatorCurrentRewards")...)
EncodeValidatorCurrentRewards(w, v)
case *ValidatorCurrentRewards:
*w = append(*w, getMagicBytes("ValidatorCurrentRewards")...)
EncodeValidatorCurrentRewards(w, *v)
case ValidatorHistoricalRewards:
*w = append(*w, getMagicBytes("ValidatorHistoricalRewards")...)
EncodeValidatorHistoricalRewards(w, v)
case *ValidatorHistoricalRewards:
*w = append(*w, getMagicBytes("ValidatorHistoricalRewards")...)
EncodeValidatorHistoricalRewards(w, *v)
case ValidatorSigningInfo:
*w = append(*w, getMagicBytes("ValidatorSigningInfo")...)
EncodeValidatorSigningInfo(w, v)
case *ValidatorSigningInfo:
*w = append(*w, getMagicBytes("ValidatorSigningInfo")...)
EncodeValidatorSigningInfo(w, *v)
case ValidatorSlashEvent:
*w = append(*w, getMagicBytes("ValidatorSlashEvent")...)
EncodeValidatorSlashEvent(w, v)
case *ValidatorSlashEvent:
*w = append(*w, getMagicBytes("ValidatorSlashEvent")...)
EncodeValidatorSlashEvent(w, *v)
case Vote:
*w = append(*w, getMagicBytes("Vote")...)
EncodeVote(w, v)
case *Vote:
*w = append(*w, getMagicBytes("Vote")...)
EncodeVote(w, *v)
case VoteOption:
*w = append(*w, getMagicBytes("VoteOption")...)
EncodeVoteOption(w, v)
case *VoteOption:
*w = append(*w, getMagicBytes("VoteOption")...)
EncodeVoteOption(w, *v)
case int64:
*w = append(*w, getMagicBytes("int64")...)
Encodeint64(w, v)
case *int64:
*w = append(*w, getMagicBytes("int64")...)
Encodeint64(w, *v)
case uint64:
*w = append(*w, getMagicBytes("uint64")...)
Encodeuint64(w, v)
case *uint64:
*w = append(*w, getMagicBytes("uint64")...)
Encodeuint64(w, *v)
default:
panic(fmt.Sprintf("Unknown Type %v %v\n", x, reflect.TypeOf(x)))
} // end of switch
} // end of func
func DecodeAny(bz []byte) (interface{}, int, error) {
var v interface{}
var magicBytes [4]byte
var n int
for i:=0; i<4; i++ {magicBytes[i] = bz[i]}
switch magicBytes {
case [4]byte{0,157,18,162}:
v, n, err := DecodeAccAddress(bz[4:])
return v, n+4, err
case [4]byte{37,72,3,140}:
v, n, err := DecodeAccAddressList(bz[4:])
return v, n+4, err
case [4]byte{153,157,134,34}:
v, n, err := DecodeBaseAccount(bz[4:])
return v, n+4, err
case [4]byte{78,248,144,54}:
v, n, err := DecodeBaseVestingAccount(bz[4:])
return v, n+4, err
case [4]byte{22,131,204,195}:
v, n, err := DecodeBondStatus(bz[4:])
return v, n+4, err
case [4]byte{2,65,204,255}:
v, n, err := DecodeCoin(bz[4:])
return v, n+4, err
case [4]byte{2,26,137,96}:
v, n, err := DecodeCommitInfo(bz[4:])
return v, n+4, err
case [4]byte{31,93,37,208}:
v, n, err := DecodeCommunityPoolSpendProposal(bz[4:])
return v, n+4, err
case [4]byte{28,53,138,173}:
v, n, err := DecodeConsAddress(bz[4:])
return v, n+4, err
case [4]byte{75,69,41,151}:
v, n, err := DecodeContinuousVestingAccount(bz[4:])
return v, n+4, err
case [4]byte{56,163,255,105}:
v, n, err := DecodeDecCoin(bz[4:])
return v, n+4, err
case [4]byte{57,34,102,152}:
v, n, err := DecodeDecCoins(bz[4:])
return v, n+4, err
case [4]byte{59,193,203,230}:
v, n, err := DecodeDelayedVestingAccount(bz[4:])
return v, n+4, err
case [4]byte{166,217,16,186}:
v, n, err := DecodeDelegation(bz[4:])
return v, n+4, err
case [4]byte{191,63,17,72}:
v, n, err := DecodeDelegatorStartingInfo(bz[4:])
return v, n+4, err
case [4]byte{18,193,65,104}:
v, n, err := DecodeFeePool(bz[4:])
return v, n+4, err
case [4]byte{54,236,180,248}:
v, n, err := DecodeInput(bz[4:])
return v, n+4, err
case [4]byte{37,29,227,212}:
v, n, err := DecodeModuleAccount(bz[4:])
return v, n+4, err
case [4]byte{141,7,107,68}:
v, n, err := DecodeMsgBeginRedelegate(bz[4:])
return v, n+4, err
case [4]byte{24,79,66,107}:
v, n, err := DecodeMsgCreateValidator(bz[4:])
return v, n+4, err
case [4]byte{184,121,196,185}:
v, n, err := DecodeMsgDelegate(bz[4:])
return v, n+4, err
case [4]byte{234,76,240,151}:
v, n, err := DecodeMsgDeposit(bz[4:])
return v, n+4, err
case [4]byte{9,254,168,109}:
v, n, err := DecodeMsgEditValidator(bz[4:])
return v, n+4, err
case [4]byte{64,119,59,163}:
v, n, err := DecodeMsgMultiSend(bz[4:])
return v, n+4, err
case [4]byte{212,255,125,220}:
v, n, err := DecodeMsgSend(bz[4:])
return v, n+4, err
case [4]byte{208,136,199,77}:
v, n, err := DecodeMsgSetWithdrawAddress(bz[4:])
return v, n+4, err
case [4]byte{21,241,6,56}:
v, n, err := DecodeMsgUndelegate(bz[4:])
return v, n+4, err
case [4]byte{139,110,39,159}:
v, n, err := DecodeMsgUnjail(bz[4:])
return v, n+4, err
case [4]byte{109,173,240,7}:
v, n, err := DecodeMsgVerifyInvariant(bz[4:])
return v, n+4, err
case [4]byte{233,121,28,250}:
v, n, err := DecodeMsgVote(bz[4:])
return v, n+4, err
case [4]byte{43,19,183,111}:
v, n, err := DecodeMsgWithdrawDelegatorReward(bz[4:])
return v, n+4, err
case [4]byte{84,85,236,88}:
v, n, err := DecodeMsgWithdrawValidatorCommission(bz[4:])
return v, n+4, err
case [4]byte{178,67,155,203}:
v, n, err := DecodeOutput(bz[4:])
return v, n+4, err
case [4]byte{66,250,248,208}:
v, n, err := DecodeParamChange(bz[4:])
return v, n+4, err
case [4]byte{49,37,122,86}:
v, n, err := DecodeParameterChangeProposal(bz[4:])
return v, n+4, err
case [4]byte{158,94,112,161}:
v, n, err := DecodePrivKeyEd25519(bz[4:])
return v, n+4, err
case [4]byte{83,16,177,42}:
v, n, err := DecodePrivKeySecp256k1(bz[4:])
return v, n+4, err
case [4]byte{114,76,37,23}:
v, n, err := DecodePubKeyEd25519(bz[4:])
return v, n+4, err
case [4]byte{14,33,23,141}:
v, n, err := DecodePubKeyMultisigThreshold(bz[4:])
return v, n+4, err
case [4]byte{51,161,20,197}:
v, n, err := DecodePubKeySecp256k1(bz[4:])
return v, n+4, err
case [4]byte{131,101,23,4}:
v, n, err := DecodeSdkDec(bz[4:])
return v, n+4, err
case [4]byte{189,210,54,221}:
v, n, err := DecodeSdkInt(bz[4:])
return v, n+4, err
case [4]byte{67,52,162,78}:
v, n, err := DecodeSignedMsgType(bz[4:])
return v, n+4, err
case [4]byte{162,148,222,207}:
v, n, err := DecodeSoftwareUpgradeProposal(bz[4:])
return v, n+4, err
case [4]byte{247,42,43,179}:
v, n, err := DecodeStdSignature(bz[4:])
return v, n+4, err
case [4]byte{247,170,118,185}:
v, n, err := DecodeStdTx(bz[4:])
return v, n+4, err
case [4]byte{224,49,135,8}:
v, n, err := DecodeStoreInfo(bz[4:])
return v, n+4, err
case [4]byte{191,66,141,63}:
v, n, err := DecodeSupply(bz[4:])
return v, n+4, err
case [4]byte{207,179,211,152}:
v, n, err := DecodeTextProposal(bz[4:])
return v, n+4, err
case [4]byte{98,39,43,120}:
v, n, err := DecodeValAddress(bz[4:])
return v, n+4, err
case [4]byte{11,142,217,17}:
v, n, err := DecodeValAddressList(bz[4:])
return v, n+4, err
case [4]byte{11,10,9,103}:
v, n, err := DecodeValidator(bz[4:])
return v, n+4, err
case [4]byte{178,20,101,34}:
v, n, err := DecodeValidatorCurrentRewards(bz[4:])
return v, n+4, err
case [4]byte{112,41,104,134}:
v, n, err := DecodeValidatorHistoricalRewards(bz[4:])
return v, n+4, err
case [4]byte{176,207,98,135}:
v, n, err := DecodeValidatorSigningInfo(bz[4:])
return v, n+4, err
case [4]byte{76,125,139,93}:
v, n, err := DecodeValidatorSlashEvent(bz[4:])
return v, n+4, err
case [4]byte{205,85,136,219}:
v, n, err := DecodeVote(bz[4:])
return v, n+4, err
case [4]byte{170,208,50,2}:
v, n, err := DecodeVoteOption(bz[4:])
return v, n+4, err
case [4]byte{188,34,41,102}:
v, n, err := Decodeint64(bz[4:])
return v, n+4, err
case [4]byte{36,210,58,112}:
v, n, err := Decodeuint64(bz[4:])
return v, n+4, err
default:
panic("Unknown type")
} // end of switch
return v, n, nil
} // end of DecodeAny
func AssignIfcPtrFromStruct(ifcPtrIn interface{}, structObjIn interface{}) {
switch ifcPtr := ifcPtrIn.(type) {
case *PubKey:
switch structObj := structObjIn.(type) {
	case PubKeySecp256k1:
	*ifcPtr = &structObj
	case PubKeyEd25519:
	*ifcPtr = &structObj
	case PubKeyMultisigThreshold:
	*ifcPtr = &structObj
	case StdSignature:
	*ifcPtr = &structObj
	default:
	panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
	} // end switch of structs
case *PrivKey:
switch structObj := structObjIn.(type) {
	case PrivKeyEd25519:
	*ifcPtr = &structObj
	case PrivKeySecp256k1:
	*ifcPtr = &structObj
	default:
	panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
	} // end switch of structs
case *Msg:
switch structObj := structObjIn.(type) {
	case MsgWithdrawValidatorCommission:
	*ifcPtr = &structObj
	case MsgSend:
	*ifcPtr = &structObj
	case MsgWithdrawDelegatorReward:
	*ifcPtr = &structObj
	case MsgDeposit:
	*ifcPtr = &structObj
	case MsgCreateValidator:
	*ifcPtr = &structObj
	case MsgEditValidator:
	*ifcPtr = &structObj
	case MsgBeginRedelegate:
	*ifcPtr = &structObj
	case MsgUndelegate:
	*ifcPtr = &structObj
	case MsgUnjail:
	*ifcPtr = &structObj
	case MsgDelegate:
	*ifcPtr = &structObj
	case MsgSetWithdrawAddress:
	*ifcPtr = &structObj
	case MsgVote:
	*ifcPtr = &structObj
	case MsgMultiSend:
	*ifcPtr = &structObj
	case MsgVerifyInvariant:
	*ifcPtr = &structObj
	default:
	panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
	} // end switch of structs
case *Account:
switch structObj := structObjIn.(type) {
	case ContinuousVestingAccount:
	*ifcPtr = &structObj
	case BaseVestingAccount:
	*ifcPtr = &structObj
	case DelayedVestingAccount:
	*ifcPtr = &structObj
	case ModuleAccount:
	*ifcPtr = &structObj
	case BaseAccount:
	*ifcPtr = &structObj
	default:
	panic(fmt.Sprintf("Type mismatch %v %v\n", reflect.TypeOf(ifcPtr), reflect.TypeOf(structObjIn)))
	} // end switch of structs
case *Content:
switch structObj := structObjIn.(type) {
	case CommunityPoolSpendProposal:
	*ifcPtr = &structObj
	case ParameterChangeProposal:
	*ifcPtr = &structObj
	case SoftwareUpgradeProposal:
	*ifcPtr = &structObj
	case TextProposal:
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
case *VestingAccount:
switch structObj := structObjIn.(type) {
	case DelayedVestingAccount:
	*ifcPtr = &structObj
	case ContinuousVestingAccount:
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
return []string {
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
