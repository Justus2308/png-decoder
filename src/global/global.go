package global

import (
	"errors"
	"flag"
)

var ( // flags
	path = flag.String("src", "", "src: path to the source image")
	mode = flag.String("mode", "encode", "mode: either encode or decode\ndefault mode: encode")
	alpha = flag.Bool("alpha", true, "alpha: enable alpha channel, if available\ndefault value: true")
	inter = flag.Bool("inter", false, "inter: enable adam7-interlacing\ndefault value: false")
)

var ( // errors
	ErrUnsupported = errors.New("unsupported format")
	ErrTransmission = errors.New("faulty transmission")
	ErrSyntax = errors.New("syntax error")
)

var ( // interlacing pattern
	Adam7 = [][]int{
		{1, 6, 4, 6, 2, 6, 4, 6},
		{7, 7, 7, 7, 7, 7, 7, 7},
		{5, 6, 5, 6, 5, 6, 5, 6},
		{7, 7, 7, 7, 7, 7, 7, 7},
		{3, 6, 4, 6, 3, 6, 4, 6},
		{7, 7, 7, 7, 7, 7, 7, 7},
		{5, 6, 5, 6, 5, 6, 5, 6},
		{7, 7, 7, 7, 7, 7, 7, 7},
	}
)

var ( // magic numbers
	BMP = []byte{0x42, 0x4D}
	PNG = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	IHDR = []byte{73, 72, 68, 82}
	PLTE = []byte{80, 76, 84, 69}
	IDAT = []byte{73, 68, 65, 84}
	IEND = []byte{73, 69, 78, 68}
)


func Path() string {
	return *path
}

func Mode() (bool, error) { // true: encode ; false: decode
	switch *mode {
	case "encode":
		return true, nil
	case "decode":
		return false, nil
	default:
		return true, errors.New("invalid operation mode")
	}
}

func Alpha() bool {
	return *alpha
}

func Interlaced() bool {
	return *inter
}


// for testing
func SetPath(p string) {
	*path = p
}

func SetMode(m string) {
	*mode = m
}

func SetAlpha(a bool) {
	*alpha = a
}

func SetInterlaced(i bool) {
	*inter = i
}