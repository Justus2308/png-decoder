package utils

import "encoding/binary"


// ...Big: Big-Endian (Network) byte order
// ...Lit: Little-Endian byte order

func BToU16Big(b []byte) uint16 {
	return uint16(b[0])<<8 | uint16(b[1])
}

func BToU32Big(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func BToU16Lit(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func BToU32Lit(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func U16toBBig(i uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b[:2], i)
	return b
}

func U32toBBig(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:4], i)
	return b
}

func U16toBLit(i uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b[:2], i)
	return b
}

func U32toBLit(i uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b[:4], i)
	return b
}

func CompLit(b []byte) []byte {
	for i := range b {
		b[i] = ^b[i]
	}
	b[0]++
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