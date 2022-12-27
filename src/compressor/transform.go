package compressor

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
)

var (
	src = flag.String("src", "", "path to the source image")
	errUnsupported = errors.New("unsupported format")
)


func bToU16(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func bToU32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func readData() ([]byte, error) {
	in, err := os.Open(*src)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	data, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func decode8Bit(data []byte, w, h, offset int, topDown bool, palette [][4]byte) ([][]byte, error) {
	if w == 0 || h == 0 {
		return [][]byte{}, nil
	}
	raw := data[offset:]
	bits := make([][]byte, h)
	y0, y1, yDelta := h-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, h, +1
	}
	for y := y0; y != y1; y += yDelta {
		p := raw[y*w : y*w+w*4-(w%4)]
		bits[y] = p
	}
	return bits, nil
}

func decode24Bit(data []byte, w, h, offset int, topDown bool) ([][]byte, error) {
	if w == 0 || h == 0 {
		return [][]byte{}, nil
	}
	raw := data[offset:]
	bits := make([][]byte, h)
	b := make([]byte, (3*w+3)&^3)
	y0, y1, yDelta := h-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, h, +1
	}
	for y := y0; y != y1; y += yDelta {
		p := raw[y*w : y*w+w*4]
		for i, j := 0, 0; i < len(p); i, j = i+4, j+3 {
			p[i+0] = b[j+2]
			p[i+1] = b[j+1]
			p[i+2] = b[j+0]
			p[i+3] = 0xFF
		}

		bits[y] = p
	}
	return bits, nil
}

func decode32Bit(data []byte, w, h, offset int, alpha, topDown bool) ([][]byte, error) {
	if w == 0 || h == 0 {
		return [][]byte{}, nil
	}
	raw := data[offset:]
	bits := make([][]byte, h)
	y0, y1, yDelta := h-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, h, +1
	}
	for y := y0; y != y1; y += yDelta {
		p := raw[y*w : y*w+w*4]
		for i := 0; i < len(p); i += 4 {
			p[i+0], p[i+2] = p[i+2], p[i+0]
			if !alpha {
			p[i+3] = 0xFF
			}
		}
		bits[y] = p
	}
	return bits, nil
}

func GetBits() ([][]byte, error) {
	data, err := readData()
	if err != nil {
		return nil, err
	}
	w, h, offset, bpp, alpha, topDown, fileinfolen, err := decodeHeader((*[138]byte)(data[:138]))
	if err != nil {
		return nil, err
	}
	switch bpp {
	case 8:
		b := data[fileinfolen:fileinfolen+256*4]
		palette := make([][4]byte, 256)
		for i := range palette {
			palette[i] = [4]byte{b[4*i+2], b[4*i+1], b[4*i+0], 0xFF}
		}
		return decode8Bit(data, w, h, offset, topDown, palette)
	case 24:
		return decode24Bit(data, w, h, offset, topDown)
	case 32:
		return decode32Bit(data, w, h, offset, alpha, topDown)
	}
	panic(errUnsupported)
}

func decodeHeader(head *[138]byte) (int, int, int, int, bool, bool, int, error) { // returns width, height, offset, bpp, alpha, topDown, file+infoheader len, err
	const (
		fileHeaderLen = 14
		infoHeaderLen = 40
		v4HeaderLen = 108
		v5HeaderLen = 124
	)
	if !bytes.Equal(head[:2], []byte("BM")) {
		return 0, 0, 0, 0, false, false, 0, errUnsupported
	}
	offset, dibLen := bToU32(head[10:fileHeaderLen]), bToU32(head[fileHeaderLen:18])
	if dibLen != infoHeaderLen && dibLen != v4HeaderLen && dibLen != v5HeaderLen {
		return 0, 0, 0, 0, false, false, 0, errors.New("unsupported bmp type")
	}
	topDown := false
	width, height := int(int32(bToU32(head[18:22]))), int(int32(bToU32(head[22:26])))
	if height < 0 {
		height, topDown = -height, true
	}
	planes, bpp, compression := bToU16(head[26:28]), bToU16(head[28:30]), bToU32(head[30:34])
	if compression == 3 && dibLen > infoHeaderLen &&
		bToU32(head[54:58]) == 0xff0000 && bToU32(head[58:62]) == 0xff00 &&
		bToU32(head[62:66]) == 0xff && bToU32(head[66:70]) == 0xff000000 {
		compression = 0
	}
	if planes != 1 || compression != 0 {
		return 0, 0, 0, 0, false, false, 0, errUnsupported
	}
	switch bpp {
	case 8:
		if offset != fileHeaderLen+dibLen+256*4 {
			return 0, 0, 0, 0, false, false, 0, errUnsupported
		}
		return width, height, int(int32(offset)), 8, false, topDown, int(int32(fileHeaderLen+dibLen)), nil
	case 24:
		if offset != fileHeaderLen+dibLen {
			return 0, 0, 0, 0, false, false, 0, errUnsupported
		}
		return width, height, int(int32(offset)), 24, false, topDown, 0, nil
	case 32:
		if offset != fileHeaderLen+dibLen {
			return 0, 0, 0, 0, false, false, 0, errUnsupported
		}
		alpha := dibLen > infoHeaderLen
		return width, height, int(int32(offset)), 32, alpha, topDown, 0, nil
	}
	return 0, 0, 0, 0, false, false, 0, errUnsupported
}