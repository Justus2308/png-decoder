package global

import (
	"errors"
)

var ( // settings
	path = ""
	alpha = true
	inter = false
)

var ( // errors
	ErrUnsupported = errors.New("unsupported format")
	ErrTransmission = errors.New("faulty transmission")
	ErrSyntax = errors.New("data syntax error")
	ErrNoPixels = errors.New("file contains no pixels")
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
	return path
}

func Alpha() bool {
	return alpha
}

func Interlaced() bool {
	return inter
}


func SetPath(p string) {
	path = p
}

func SetAlpha(a bool) {
	alpha = a
}

func SetInterlaced(i bool) {
	inter = i
}

func Reset() {
	SetPath("")
	SetAlpha(true)
	SetInterlaced(false)
}