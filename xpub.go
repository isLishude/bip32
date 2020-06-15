package bip32

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"strconv"
)

type XPub struct {
	xpub []byte
}

func NewXPub(raw []byte) XPub {
	if len(raw) != XPubSize {
		panic("bip32: xpub size should be 64 bytes")
	}
	return XPub{xpub: append([]byte(nil), raw...)}
}

func (x XPub) String() string {
	return hex.EncodeToString(x.xpub)
}

func (x XPub) Bytes() []byte {
	return append([]byte(nil), x.xpub...)
}

func (x XPub) PublicKey() ed25519.PublicKey {
	return append([]byte(nil), x.xpub[:32]...)
}

func (x XPub) Derive(index uint32) XPub {
	if index > HardIndex {
		panic("bip32: xpub: expected a soft derivation," + strconv.FormatUint(uint64(index), 10))
	}

	var pubkey [32]byte
	copy(pubkey[:], x.xpub[:32])
	chaincode := append([]byte(nil), x.xpub[32:]...)

	zmac := hmac.New(sha512.New, chaincode)
	imac := hmac.New(sha512.New, chaincode)

	seri := make([]byte, 4)
	binary.LittleEndian.PutUint32(seri, index)

	_, _ = zmac.Write([]byte{2})
	_, _ = zmac.Write(pubkey[:])
	_, _ = zmac.Write(seri)

	_, _ = imac.Write([]byte{3})
	_, _ = imac.Write(pubkey[:])
	_, _ = imac.Write(seri)

	left, ok := pointPlus(&pubkey, pointOfTrunc28Mul8(zmac.Sum(nil)[:32]))
	if !ok {
		panic("can't convert bytes to edwards25519.ExtendedGroupElement")
	}

	var out [64]byte
	copy(out[:32], left[:32])
	copy(out[32:], imac.Sum(nil)[32:])
	return XPub{xpub: out[:]}
}

func (x XPub) Verify(msg, sig []byte) bool {
	pk := x.xpub[:32]
	return ed25519.Verify(pk, msg, sig)
}
