package encode

import (
	"encoding/binary"
	"hash/crc32"

	"png-decoder/src/global"
)


func u32toB(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:4], i)
	return b
}

func makeIHDR(w, h, bpp int, alpha, interlaced bool) []byte {
	ihdr := u32toB(uint32(13)) // data field length
	ihdr = append(ihdr, global.IHDR...) // chunk type field
	ihdr = append(ihdr, u32toB(uint32(w))...) // width in 4 bits
	ihdr = append(ihdr, u32toB(uint32(h))...) // height in 4 bits
	switch bpp { // sample depth + colour type, only supports indexed-colour, truecolour and truecolour+alpha
	case 8:
		ihdr = append(ihdr, []byte{8, 3}...)
	case 24:
		ihdr = append(ihdr, []byte{8, 2}...)
	case 32:
		if alpha {
			ihdr = append(ihdr, []byte{8, 6}...)
		} else {
			ihdr = append(ihdr, []byte{8, 2}...)
		}
	}
	ihdr = append(ihdr, 0) // compression method (only 0 specified)
	ihdr = append(ihdr, 0) // filter method (only 0 specified)
	if interlaced { // interlace method (Adam7 or none)
		ihdr = append(ihdr, 1)
	} else {
		ihdr = append(ihdr, 0)
	}
	ihdr = append(ihdr, u32toB(crc32.ChecksumIEEE(ihdr[4:]))...) // crc32 checksum
	return ihdr
}

func makePLTE(palette []byte) []byte {
	plte := u32toB(uint32(len(palette))) // data field length
	plte = append(plte, global.PLTE...) // chunk type field
	plte = append(plte, palette...) // data field
	plte = append(plte, u32toB(crc32.ChecksumIEEE(plte[4:]))...) // crc32 checksum
	return plte
}

func makeIDAT(data []byte) []byte {
	idat := u32toB(uint32(len(data))) // data field length
	idat = append(idat, global.IDAT...) // chunk type field
	idat = append(idat, data...) // data field
	idat = append(idat, u32toB(crc32.ChecksumIEEE(idat[4:]))...) // crc32 checksum
	return idat
}

func makeIEND() []byte {
	iend := []byte{0x00, 0x00, 0x00, 0x00} // data field length
	iend = append(iend, global.IEND...) // chunk type field
	iend = append(iend, u32toB(crc32.ChecksumIEEE(iend[4:]))...) // crc32 checksum
	return iend
}