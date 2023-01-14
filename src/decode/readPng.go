package decode

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"unicode"

	"png-decoder/src/global"
	"png-decoder/src/utils"
)

var ( // errors
	warnUnknownAncChunk = errors.New("file contains unsupported ancilliary chunk")
	isPLTE = errors.New("INTERNAL: decoded PLTE chunk")
	isIDAT = errors.New("INTERNAL: decoded IDAT chunk")
	isIEND = errors.New("INTERNAL: reached IEND chunk")
)

var ( // image state
	hasPLTE = false
	readingIDATs = false
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

func decodeNext(png *os.File, pal bool) (data []byte, err error) {
	len := make([]byte, 4)
	_, err = png.Read(len)
	if err == io.EOF {
		return nil, global.ErrTransmission
	}
	dataLen := utils.BToU32Big(len)
	data = make([]byte, 4+dataLen+4)
	_, err = png.Read(data[:4])
	if err == io.EOF {
		return nil, global.ErrTransmission
	}
	switch {
	case bytes.Equal(data[:4], global.PLTE):
		if hasPLTE || readingIDATs {
			return nil, global.ErrSyntax
		}
		hasPLTE = true
		if dataLen%3 != 0 {
			return nil, global.ErrTransmission
		}
		_, err = png.Read(data[4:])
		if err == io.EOF {
			return nil, global.ErrTransmission
		}
		checksum := utils.U32toBBig(crc32.ChecksumIEEE(data[:4+dataLen]))
		if !bytes.Equal(data[4+dataLen:], checksum) {
			return nil, global.ErrTransmission
		}
		hasPLTE = true
		return data[4:4+dataLen], isPLTE
	case bytes.Equal(data[:4], global.IDAT):
		if pal && !hasPLTE {
			return nil, global.ErrSyntax
		}
		readingIDATs = true
		_, err = png.Read(data[4:])
		if err == io.EOF {
			return nil, global.ErrTransmission
		}
		checksum := utils.U32toBBig(crc32.ChecksumIEEE(data[:4+dataLen]))
		if !bytes.Equal(data[4+dataLen:], checksum) {
			return nil, global.ErrTransmission
		}
		return data[4:4+dataLen], isIDAT
	case bytes.Equal(data[:4], global.IEND):
		_, err = png.Read(data[4:])
		if err == io.EOF {
			return nil, global.ErrTransmission
		}
		checksum := utils.U32toBBig(crc32.ChecksumIEEE(data[:4]))
		if dataLen == 0 && bytes.Equal(data[4:], checksum) {
			return nil, isIEND
		}
		return nil, global.ErrTransmission
	}
	if readingIDATs {
		return nil, global.ErrSyntax
	}
	if unicode.IsUpper(rune(data[0])) { // checks whether chunk is critical or ancilliary
		return nil, global.ErrSyntax
	}
	_, err = png.Read(data[4:])
	if err == io.EOF {
		return nil, global.ErrTransmission
	}
	checksum := utils.U32toBBig(crc32.ChecksumIEEE(data[:4+dataLen]))
	if !bytes.Equal(data[4+dataLen:], checksum) {
		return nil, global.ErrTransmission
	}
	privB := "private"
	if unicode.IsUpper(rune(data[1])) { // checks whether chunk is public or private
		privB = "public"
	}
	return nil, fmt.Errorf("%w %v (%v)", warnUnknownAncChunk, string(data[:4]), privB)
}