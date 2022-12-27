package encode

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
)


func u32toB(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:4], i)
	return b
}

func deflate(filt [][]byte) []byte {
	var buf bytes.Buffer
	w, _ := zlib.NewWriterLevel(&buf, 8) // deflate with compression level 8
	defer w.Close()
	for _, b := range filt {
		w.Write(b)
	}
	return buf.Bytes()
}

func Chunk(filt [][]byte, w, h, bpp int, alpha, interlaced bool, palette [][4]byte) []byte {
	chunked := makeIHDR(w, h, bpp, alpha, interlaced)
	if bpp == 8 {
		chunked = append(chunked, makePLTE(palette)...)
	}
	chunked = append(chunked, makeIDAT(filt)...)
	chunked = append(chunked, makeIEND()...)
	return chunked
}

func makeIHDR(w, h, bpp int, alpha, interlaced bool) []byte {
	ihdr := []byte{73, 72, 68, 82} // chunk type field
	ihdr = append(ihdr, u32toB(uint32(w))...) // width in 4 bits
	ihdr = append(ihdr, u32toB(uint32(h))...) // height in 4 bits
	ihdr = append(ihdr, byte(bpp)) // bit depth
	switch bpp { // colour type, only supports indexed-colour, truecolour and truecolour+alpha
	case 8:
		ihdr = append(ihdr, 3)
	case 24:
		ihdr = append(ihdr, 2)
	case 32:
		if alpha {
			ihdr = append(ihdr, 6)
		} else {
			ihdr = append(ihdr, 2)
		}
	}
	ihdr = append(ihdr, 0) // compression method (only 0 specified)
	ihdr = append(ihdr, 0) // filter method (only 0 specified)
	if interlaced { // interlace method (Adam7 or none)
		ihdr = append(ihdr, 1)
	} else {
		ihdr = append(ihdr, 0)
	}
	return ihdr
}

func makePLTE(palette [][4]byte) []byte {
	plte := []byte{80, 76, 84, 69} // chunk type field
	for _, p := range palette { // bmp palettes are stored in B-G-R-X format
		plte = append(plte, p[2]) // R
		plte = append(plte, p[1]) // G
		plte = append(plte, p[0]) // B
	}
	return plte
}

func makeIDAT(filt [][]byte) []byte {
	idat := []byte{73, 68, 65, 84}
	data := deflate(filt)
	idat = append(idat, data...)
	return idat
}

func makeIEND() []byte {
	return []byte{73, 69, 78, 68}
}