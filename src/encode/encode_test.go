package encode

import (
	"testing"

	"png-decoder/src/global"
)

var path = "test_images/test_32bpp.bmp"


func TestGetBits(t *testing.T) {
	global.SetPath(path)
	transformed, _, _, bpp, _, err := GetBits()
	t.Log("depth:", bpp)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(transformed)
}

func TestMinAbsDiff(t *testing.T) {
	global.SetPath(path)
	bitsT, _, _, _, _, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	for _, v := range bitsT {
		t.Log(minAbsDiff(v))
	}
}

func TestFilter(t *testing.T) {
	global.SetPath(path)
	bits, w, h, bpp, _, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	filtered := Filter(bits, w, h, bpp)
	t.Log(filtered)
}

func TestIHDR(t *testing.T) {
	ihdr := makeIHDR(600, 750, 32, true, false)
	t.Log(ihdr)
}

func TestDeflate(t *testing.T) {
	global.SetPath(path)
	bits, _, _, _, _, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	defl := deflate(bits)
	t.Log(defl)
}

func TestChunker(t *testing.T) {
	global.SetPath(path)
	bits, w, h, bpp, alpha, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	filt := Filter(bits, w, h, bpp)
	chunked := Chunk(filt, w, h, bpp, alpha, false, nil)
	t.Log(chunked)
}

func TestCreatePng(t *testing.T) {
	global.SetPath(path)
	bits, w, h, bpp, alpha, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	filt := Filter(bits, w, h, bpp)
	chunked := Chunk(filt, w, h, bpp, alpha, false, nil)
	err = createPng(chunked)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCreateUnfilteredPng(t *testing.T) {
	global.SetPath(path)
	bits, w, h, bpp, alpha, err := GetBits()
	if err != nil {
		t.Error(err)
		return
	}
	// inter := Interlace(bits, w, h)
	// t.Log(inter)
	filt := make([][]byte, h, h)
	for i := range bits {
		filt[i] = subFltr(bits, i, w, 4)
		t.Log(cap(filt[i]))
		// filt[i] = typeByte(bits[i], 0)
	}
	t.Log(filt)
	chunked := Chunk(filt, w, h, bpp, alpha, false, nil)
	suffix = "_unfilt.png"
	err = createPng(chunked)
	if err != nil {
		t.Error(err)
		return
	}
}