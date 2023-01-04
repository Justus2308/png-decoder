package encode

import (
	"bytes"
	"errors"
	"io"
	"os"

	"png-decoder/src/global"
	"png-decoder/src/util"
)


var (
	prevCache []byte
)


func decodeHeader(bmp *os.File) (w, h, depth int, alpha, topDown bool, err error) {
	const (
		fileHeaderLen = 14
		infoHeaderLen = 40
		v4HeaderLen = 108
		v5HeaderLen = 124
	)
	head := make([]byte, 138)
	if _, err := bmp.Read(head[:fileHeaderLen+4]); err != nil {
		if err == io.EOF {
			err = global.ErrTransmission
		}
		return 0, 0, 0, false, false, err
	}
	if !bytes.Equal(head[:2], global.BMP) {
		return 0, 0, 0, false, false, global.ErrUnsupported
	}
	offset, dibLen := util.BToU32(head[10:fileHeaderLen]), util.BToU32(head[fileHeaderLen:18])
	if dibLen != infoHeaderLen && dibLen != v4HeaderLen && dibLen != v5HeaderLen {
		return 0, 0, 0, false, false, errors.New("unsupported bmp type")
	}
	if _, err := bmp.Read(head[fileHeaderLen+4:fileHeaderLen+dibLen]); err != nil {
		if err == io.EOF {
			err = global.ErrTransmission
		}
		return 0, 0, 0, false, false, err
	}
	width, height := int(int32(util.BToU32(head[18:22]))), int(int32(util.BToU32(head[22:26])))
	if height < 0 {
		height, topDown = -height, true
	}
	planes, bpp, compression := util.BToU16(head[26:28]), util.BToU16(head[28:30]), util.BToU32(head[30:34])
	if compression == 3 && dibLen > infoHeaderLen &&
		util.BToU32(head[54:58]) == 0xff0000 && util.BToU32(head[58:62]) == 0xff00 &&
		util.BToU32(head[62:66]) == 0xff && util.BToU32(head[66:70]) == 0xff000000 {
		compression = 0
	}
	if planes != 1 || compression != 0 {
		return 0, 0, 0, false, false, global.ErrUnsupported
	}
	switch bpp {
	case 8:
		if offset != fileHeaderLen+dibLen+256*4 {
			return 0, 0, 0, false, false, global.ErrSyntax
		}
		return width, height, 8, false, topDown, nil
	case 24:
		if offset != fileHeaderLen+dibLen {
			return 0, 0, 0, false, false, global.ErrSyntax
		}
		return width, height, 24, false, topDown, nil
	case 32:
		if offset != fileHeaderLen+dibLen {
			return 0, 0, 0, false, false, global.ErrSyntax
		}
		alpha := dibLen > infoHeaderLen
		return width, height, 32, alpha, topDown, nil
	}
	return 0, 0, 0, false, false, global.ErrUnsupported
}

func scanLine(bmp *os.File, w, s int) (line []byte, err error) {
	buf := make([]byte, w*s)
	_, err = bmp.Read(buf)
	if err == io.EOF {
		return nil, global.ErrTransmission
	}
	return buf, nil
}

func getPalette(bmp *os.File) (plte []byte, err error) {
	raw, err := scanLine(bmp, 256, 4)
	if err != nil {
		return nil, err
	}
	plte = make([]byte, 256*3)
	for i, j := 0, 0; i < 256*4; i, j = i+4, j+3 {
		plte[j+0] = raw[i+2]
		plte[j+1] = raw[i+1]
		plte[j+2] = raw[i+0]
		// every 4th bite is padding and gets thrown away
	}
	return plte, nil
} 