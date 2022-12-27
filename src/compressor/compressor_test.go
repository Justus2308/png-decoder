package compressor

import (
	"testing"
)

var str = "/Users/justusklausecker/Downloads/colors_24bpp.bmp"

func TestGetBits(t *testing.T) {
	src = &str
	transformed, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(transformed)
}

func TestMinAbsDiff(t *testing.T) {
	src = &str
	bitsT, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(bitsT)
	for _, v := range bitsT {
		t.Log(MinAbsDiff(v))
	}
}