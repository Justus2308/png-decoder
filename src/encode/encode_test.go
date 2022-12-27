package encode

import (
	"testing"
)

var str = "/Users/justusklausecker//git/png-decoder/src/encode/test_images/test_24bpp_source.bmp"

func TestGetBits(t *testing.T) {
	source = &str
	transformed, _, _, _, _, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(transformed)
}

func TestMinAbsDiff(t *testing.T) {
	source = &str
	bitsT, _, _, _, _, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(bitsT)
	for _, v := range bitsT {
		t.Log(minAbsDiff(v))
	}
}

func TestFilter(t *testing.T) {
	source = &str
	bits, w, h, bpp, _, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	filtered, ids := Filter(&bits, w, h, bpp)
	t.Log(filtered)
	t.Log(ids)
}

func TestIHDR(t *testing.T) {
	ihdr := makeIHDR(600, 750, 32, true, false)
	t.Log(ihdr)
}

func TestDeflate(t *testing.T) {
	source = &str
	bits, _, _, _, _, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	defl, err := deflate(bits)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(defl)
}

func TestChunker(t *testing.T) {
	source = &str
	bits, w, h, bpp, alpha, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	filt, _ := Filter(&bits, w, h, bpp)
	chunked, err := Chunk(filt, w, h, bpp, alpha, false, nil)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(chunked)
}

func TestCreatePng(t *testing.T) {
	source = &str
	bits, w, h, bpp, alpha, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	filt, _ := Filter(&bits, w, h, bpp)
	chunked, err := Chunk(filt, w, h, bpp, alpha, false, nil)
	if err != nil {
		t.Error(err)
		return
	}
	err = makePng(chunked, "/Users/justusklausecker/git/png-decoder/src/encode/test_images", "test_24bpp")
	if err != nil {
		t.Error(err)
		return
	}
}