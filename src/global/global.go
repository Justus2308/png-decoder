package global

import (
	"errors"
	"flag"
)

var (
	path = flag.String("src", "", "src: path to the source image")
	mode = flag.String("mode", "encode", "mode: either encode or decode\ndefault mode: encode")
	adam7 = [][]int{
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

func Adam7() [][]int {
	return adam7
}

func SetPath(p string) { // for testing
	*path = p
}

func SetMode(m string) { // for testing
	*mode = m
}