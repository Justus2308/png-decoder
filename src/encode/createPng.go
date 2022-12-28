package encode

import (
	"os"
	"strings"

	"png-decoder/src/global"
)

var (
	magicNumbers = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
)

func createPng(stream []byte) error {
	trgt := strings.TrimSuffix(global.Path(), ".bmp")
	f, err := os.Create(trgt+"_enc.png")
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(magicNumbers)
	if err != nil {
		return err
	}
	_, err = f.Write(stream)
	if err != nil {
		return err
	}
	return nil
}