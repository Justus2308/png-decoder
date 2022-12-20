package compressor

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	src = flag.String("src", "", "path to the source image")
)

func GetBMP() ([][]byte, error) {
	in, err := os.Open(*src)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}
	fmt.Println(data)
	return nil, nil
}