package encode

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"hash/crc32"
)


func u32toB(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:4], i)
	return b
}

func deflate(filt [][]byte) ([]byte, error) {
	var (
		buf bytes.Buffer
		defl []byte
	)
	w, _ := zlib.NewWriterLevel(&buf, 8)
	for _, f := range filt {
		_, err := bytes.NewReader(f).WriteTo(w)
		if err != nil {
			return nil, err
		}
		w.Flush()
		defl = append(defl, buf.Bytes()...)
		buf.Reset()
	}
	return defl, nil
}

func Chunk(filt [][]byte, w, h, bpp int, alpha, interlaced bool, palette [][4]byte) ([]byte, error) {
	chunked := makeIHDR(w, h, bpp, alpha, interlaced)
	if bpp == 8 {
		chunked = append(chunked, makePLTE(palette)...)
	}
	idat, err := makeIDAT(filt)
	if err != nil {
		return nil, err
	}
	chunked = append(chunked, idat...)
	chunked = append(chunked, makeIEND()...)
	return chunked, nil
}

func makeIHDR(w, h, bpp int, alpha, interlaced bool) []byte {
	ihdr := u32toB(uint32(13)) // data field length
	ihdr = append(ihdr, []byte{73, 72, 68, 82}...) // chunk type field
	ihdr = append(ihdr, u32toB(uint32(w))...) // width in 4 bits
	ihdr = append(ihdr, u32toB(uint32(h))...) // height in 4 bits
	switch bpp { // bit depth + colour type, only supports indexed-colour, truecolour and truecolour+alpha
	case 8:
		ihdr = append(ihdr, []byte{8, 3}...)
	case 24:
		ihdr = append(ihdr, []byte{16, 2}...)
	case 32:
		if alpha {
			ihdr = append(ihdr, []byte{16, 6}...)
		} else {
			ihdr = append(ihdr, []byte{16, 2}...)
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

func makePLTE(palette [][4]byte) []byte {
	plte := u32toB(uint32(len(palette)*3)) // data field length
	plte = append(plte, []byte{80, 76, 84, 69}...) // chunk type field
	for _, p := range palette { // bmp palettes are stored in B-G-R-X format
		plte = append(plte, p[2]) // R
		plte = append(plte, p[1]) // G
		plte = append(plte, p[0]) // B
	}
	plte = append(plte, u32toB(crc32.ChecksumIEEE(plte[4:]))...) // crc32 checksum
	return plte
}

func makeIDAT(filt [][]byte) ([]byte, error) {
	data, err := deflate(filt) // deflated image data stream
	if err != nil {
		return nil, err
	}
	idat := u32toB(uint32(len(data))) // data field length
	idat = append(idat, []byte{73, 68, 65, 84}...) // chunk type field
	idat = append(idat, data...) // data field
	idat = append(idat, u32toB(crc32.ChecksumIEEE(idat[4:]))...) // crc32 checksum
	return idat, nil
}

func makeIEND() []byte {
	iend := []byte{0x00, 0x00, 0x00, 0x00, 73, 69, 78, 68} // data field length + chunk type field
	iend = append(iend, u32toB(crc32.ChecksumIEEE(iend[4:]))...) // crc32 checksum
	return iend
}