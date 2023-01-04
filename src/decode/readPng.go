package decode

import (
	"bytes"
	"errors"
	"hash/crc32"
	"io"
	"os"

	"png-decoder/src/global"
	"png-decoder/src/util"
)

var (
	errUnknownAncChunk = errors.New("unknown ancilliary chunk type")
	errIEND = errors.New("internal: reached IEND chunk")
)


func decodeIHDR(png *os.File) (w, h, depth int, alpha, inter bool, err error) {
	const (
		ihdrLen = 13
	)
	head := make([]byte, 25)
	if _, err := png.Read(head[:4+4+ihdrLen+4]); err != nil {
		if err == io.EOF {
			err = global.ErrTransmission
		}
		return 0, 0, 0, false, false, err
	}
	if !bytes.Equal(head[:4], global.PNG) {
		return 0, 0, 0, false, false, global.ErrUnsupported
	}
	if !bytes.Equal(head[4 : 4+4], util.U32toB(ihdrLen)) {
		return 0, 0, 0, false, false, global.ErrSyntax
	}
	checksum := util.U32toB(crc32.ChecksumIEEE(head[4 : 4+4+ihdrLen]))
	if !bytes.Equal(head[4+4+ihdrLen : 4+4+ihdrLen+4], checksum) {
		return 0, 0, 0, false, false, global.ErrTransmission
	}
	width, height := int(int32(util.BToU32(head[8:12]))), int(int32(util.BToU32(head[12:16])))
	bps, colType := head[17], head[18]
	if bps != 8 {
		return 0, 0, 0, false, false, errors.New("unsupported bit depth")
	}
	compMet, filtMet := head[19], head[20]
	if compMet != 0 || filtMet != 0 {
		return 0, 0, 0, false, false, global.ErrUnsupported
	}
	interlaced := head[21]
	switch interlaced {
	case 0:
		inter = false
	case 1:
		inter = true
	default:
		return 0, 0, 0, false, false, global.ErrUnsupported
	}
	switch colType {
	case 2: // truecolour
		return width, height, 24, false, inter, nil
	case 3: // paletted
		return width, height, 8, false, inter, nil
	case 6: // truecolour with alpha
		return width, height, 32, true, inter, nil
	}
	return 0, 0, 0, false, false, errors.New("unsupported colour type")
}

func decodePLTE(png *os.File) (plte []byte, err error) {
	len := make([]byte, 4)
	_, err = png.Read(len)
	if err == io.EOF {
		return nil, global.ErrTransmission
	}
	plteLen := util.BToU32(len)
	plte = make([]byte, 4+plteLen+4)
	_, err = png.Read(plte[:4])
	if err == io.EOF {
		return nil, global.ErrTransmission
	}
	if bytes.Equal(plte[:4], global.PLTE) {
		if plteLen%3 != 0 {
			return nil, global.ErrTransmission
		}
		_, err = png.Read(plte[4:])
		if err == io.EOF {
			return nil, global.ErrTransmission
		}
		checksum := util.U32toB(crc32.ChecksumIEEE(plte[:plteLen-4]))
		if !bytes.Equal(plte[plteLen-4:], checksum) {
			return nil, global.ErrTransmission
		}
		return plte[4:plteLen-4], nil
	}
	return nil, global.ErrSyntax
}

func decodeIDAT(png *os.File) (data []byte, err error) {
	len := make([]byte, 4)
	_, err = png.Read(len)
	if err == io.EOF {
		return nil, global.ErrTransmission
	}
	dataLen := util.BToU32(len)
	data = make([]byte, 4+dataLen+4)
	_, err = png.Read(data[:4])
	if err == io.EOF {
		return nil, global.ErrTransmission
	}
	switch {
	case bytes.Equal(data[:4], global.IDAT):
		_, err = png.Read(data[4:])
		if err == io.EOF {
			return nil, global.ErrTransmission
		}
		checksum := util.U32toB(crc32.ChecksumIEEE(data[:dataLen-4]))
		if !bytes.Equal(data[dataLen-4:], checksum) {
			return nil, global.ErrTransmission
		}
		return data[4:dataLen-4], nil
	case bytes.Equal(data[:4], global.IEND):
		_, err = png.Read(data[4:])
		if err == io.EOF {
			return nil, global.ErrTransmission
		}
		checksum := util.U32toB(crc32.ChecksumIEEE(data[:4]))
		if dataLen == 0 && bytes.Equal(data[4:], checksum) {
			return nil, errIEND
		}
		return nil, global.ErrTransmission
	}
	if data[0] & 0b00001000 != 0 {
		return nil, global.ErrSyntax
	}
	return nil, errUnknownAncChunk
}