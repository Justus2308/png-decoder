package decode

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"

	"png-decoder/src/global"
	"png-decoder/src/utils"
)

var (
	WarnUnknownAncChunk = errors.New("file contains unsupported ancilliary chunk")
	errIEND = errors.New("internal: reached IEND chunk")
)


func decodeIHDR(png *os.File) (w, h, depth int, inter bool, err error) {
	const (
		ihdrLen = 13
	)
	head := make([]byte, 33)
	if _, err := png.Read(head[:8+4+4+ihdrLen+4]); err != nil {
		if err == io.EOF {
			err = global.ErrTransmission
		}
		return 0, 0, 0, false, err
	}
	if !bytes.Equal(head[:4], global.PNG[:4]) {
		return 0, 0, 0, false, global.ErrUnsupported
	}
	if !bytes.Equal(head[4:8], global.PNG[4:8]) {
		return 0, 0, 0, false, global.ErrTransmission
	}
	if !bytes.Equal(head[8 : 8+4], utils.U32toBBig(ihdrLen)) {
		return 0, 0, 0, false, global.ErrTransmission
	}
	if !bytes.Equal(head[8+4 : 8+4+4], global.IHDR) {
		return 0, 0, 0, false, global.ErrSyntax
	}
	checksum := utils.U32toBBig(crc32.ChecksumIEEE(head[8+4 : 8+4+4+ihdrLen]))
	if !bytes.Equal(head[8+4+4+ihdrLen : 8+4+4+ihdrLen+4], checksum) {
		return 0, 0, 0, false, global.ErrTransmission
	}
	width, height := int(int32(utils.BToU32Big(head[16:20]))), int(int32(utils.BToU32Big(head[20:24])))
	bps, colType := head[24], head[25]
	if bps != 8 {
		return 0, 0, 0, false, errors.New("unsupported bit depth")
	}
	compMet, filtMet := head[26], head[27]
	if compMet != 0 || filtMet != 0 {
		return 0, 0, 0, false, global.ErrUnsupported
	}
	interlaced := head[28]
	switch interlaced {
	case 0:
		inter = false
	case 1:
		inter = true
	default:
		return 0, 0, 0, false, global.ErrUnsupported
	}
	switch colType {
	case 2: // truecolour
		return width, height, 24, inter, nil
	case 3: // paletted
		return width, height, 8, inter, nil
	case 6: // truecolour with alpha
		return width, height, 32, inter, nil
	}
	return 0, 0, 0, false, errors.New("unsupported colour type")
}

func decodePLTE(png *os.File) (plte []byte, err error) {
	len := make([]byte, 4)
	_, err = png.Read(len)
	if err == io.EOF {
		return nil, global.ErrTransmission
	}
	plteLen := utils.BToU32Big(len)
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
		checksum := utils.U32toBBig(crc32.ChecksumIEEE(plte[:plteLen-4]))
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
	dataLen := utils.BToU32Big(len)
	fmt.Println(dataLen)
	data = make([]byte, 4+dataLen+4)
	_, err = png.Read(data[:4])
	if err == io.EOF {
		return nil, global.ErrTransmission
	}
	fmt.Println(data)
	switch {
	case bytes.Equal(data[:4], global.IDAT):
		_, err = png.Read(data[4:])
		if err == io.EOF {
			return nil, global.ErrTransmission
		}
		checksum := utils.U32toBBig(crc32.ChecksumIEEE(data[:dataLen-4]))
		if !bytes.Equal(data[dataLen-4:], checksum) {
			return nil, global.ErrTransmission
		}
		return data[4:dataLen-4], nil
	case bytes.Equal(data[:4], global.IEND):
		_, err = png.Read(data[4:])
		if err == io.EOF {
			return nil, global.ErrTransmission
		}
		checksum := utils.U32toBBig(crc32.ChecksumIEEE(data[:4]))
		if dataLen == 0 && bytes.Equal(data[4:], checksum) {
			return nil, errIEND
		}
		return nil, global.ErrTransmission
	}
	if data[0] & 0b00001000 != 0 {
		return nil, global.ErrSyntax
	}
	return nil, WarnUnknownAncChunk
}