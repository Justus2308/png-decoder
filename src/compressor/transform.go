package compressor

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
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

func decode8Bit(w, h, offset int, topDown bool) ([][]byte, error) {
	return nil, nil
}

func decode24Bit(w, h, offset int, topDown bool) ([][]byte, error) {
	return nil, nil
}

func decode32Bit(w, h, offset int, alpha, topDown bool) ([][]byte, error) {
	return nil, nil
}

func getBMP() ([][]byte, error) {
	data, err := readData()
	if err != nil {
		return nil, err
	}
	w, h, offset, bpp, alpha, topDown, err := decodeHeader((*[138]byte)(data[:138]))
	if err != nil {
		return nil, err
	}
	switch bpp {
	case 8:
		return decode8Bit(w, h, offset, topDown)
	case 24:
		return decode24Bit(w, h, offset, topDown)
	case 32:
		return decode32Bit(w, h, offset, alpha, topDown)
	}
	return nil, errUnsupported
}

func decodeHeader(head *[138]byte) (int, int, int, int, bool, bool, error) { // returns width, height, offset, bpp, alpha, topDown, err
	const (
		fileHeaderLen = 14
		infoHeaderLen = 40
		v4HeaderLen = 108
		v5HeaderLen = 124
	)
	if !bytes.Equal(head[:2], []byte("BM")) {
		return 0, 0, 0, 0, false, false, errUnsupported
	}
	offset, dibLen := bToU32(head[10:fileHeaderLen]), bToU32(head[fileHeaderLen:18])
	if dibLen != infoHeaderLen || dibLen != v4HeaderLen || dibLen != v5HeaderLen {
		return 0, 0, 0, 0, false, false, errors.New("unsupported bmp type")
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
		return 0, 0, 0, 0, false, false, errUnsupported
	}
	switch bpp {
	case 8:
		if offset != fileHeaderLen+dibLen+256*4 {
			return 0, 0, 0, 0, false, false, errUnsupported
		}
		return width, height, int(int32(offset)), 8, false, topDown, nil
	case 24:
		if offset != fileHeaderLen+dibLen {
			return 0, 0, 0, 0, false, false, errUnsupported
		}
		return width, height, int(int32(offset)), 24, false, topDown, nil
	case 32:
		if offset != fileHeaderLen+dibLen {
			return 0, 0, 0, 0, false, false, errUnsupported
		}
		alpha := dibLen > infoHeaderLen
		return width, height, int(int32(offset)), 32, alpha, topDown, nil
	}
	return 0, 0, 0, 0, false, false, errUnsupported
}

func getRGBAX() [5]int {
	return [5]int{}
}

func main() {
	str := "C:\\Users\\justu\\Downloads\\bmp_24.bmp"
	src = &str
	fmt.Println(getBMP())
}