package global

import (
	"errors"
	"flag"
)

var (
	mode = flag.String("mode", "encode", "mode: either encode or decode\ndefault mode: encode")
	path = flag.String("src", "", "src: path to the source image")
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

func SetPath(p string) { // for testing
	*path = p
}

func SetMode(m string) { // for testing
	*mode = m
}