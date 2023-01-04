package decode

import (
	"os"

	"png-decoder/src/global"
)

var (
	suffix = "_dec.bmp"
)


func Decode() {
	png, err := os.Open(global.Path())
	if err != nil {
		panic(err)
	}
	defer png.Close()

}