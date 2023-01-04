package util

import "encoding/binary"


func BToU16(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func BToU32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func U16toB(i uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b[:2], i)
	return b
}

func U32toB(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:4], i)
	return b
}

func PaethPred(a, b, c byte) byte {
	p := int16(a) + int16(b) - int16(c)
	pa := absU8(p - int16(a))
	pb := absU8(p - int16(b))
	pc := absU8(p - int16(c))
	if pa <= pb && pa <= pc {
		return a
	} else if pb <= pc {
		return b
	} else {
		return c
	}
}

func absU8(i int16) uint8 {
	if i < 0 {
		return uint8(-i)
	}
	return uint8(i)
}