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
)

func bytesToUint16(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func bytesToUint32(b []byte) uint32 {
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

func getBMP() ([][]byte, error) {
	data, err := readData()
	if err != nil {
		return nil, err
	}
	fmt.Println(data)
	decodeHeader((*[138]byte)(data[:138]))
	return nil, nil
}

func decodeHeader(header *[138]byte) (int, int, error) { // returns size, offset, err
	const (
		fileHeaderLen = 14
		infoHeaderLen = 40
		v4HeaderLen = 108
		v5HeaderLen = 124
	)
	if !bytes.Equal(header[:2], []byte("BM")) {
		return 0, 0, errors.New("unsupported format")
	}
	offset := bytesToUint32(header[10:fileHeaderLen])
	dibLen := bytesToUint32(header[fileHeaderLen:18])
	if dibLen != infoHeaderLen || dibLen != v4HeaderLen || dibLen != v5HeaderLen {
		return 0, 0, errors.New("unsupported bmp type")
	}

}

func getRGBAX() [5]int {
	return [5]int{}
}

func main() {
	str := "C:\\Users\\justu\\Downloads\\bmp_24.bmp"
	src = &str
	fmt.Println(getBMP())
}