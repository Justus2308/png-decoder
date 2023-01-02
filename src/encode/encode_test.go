package encode

import (
	"bytes"
	"compress/zlib"
	"os"
	"strings"
	"testing"

	"png-decoder/src/global"
)

var path = "test_images/test_32bpp.bmp"


func TestIHDR(t *testing.T) {
	ihdr := makeIHDR(600, 750, 32, true, false)
	t.Log(ihdr)
}

func TestEncode(t *testing.T) {
	global.SetPath(path)
	Encode()
}

func TestCreateUnfilteredPng(t *testing.T) { // works for 24 and 32 bpp
	global.SetPath(path)
	bmp, err := os.Open(global.Path())
	if err != nil {
		t.Error(err)
		return
	}
	defer bmp.Close()
	w, h, bpp, alpha, topDown, _, err := decodeHeader(bmp)
	if err != nil {
		t.Error(err)
		return
	}
	if w == 0 || h == 0 {
		t.Error("file contains no pixels")
		return
	}
	if alpha {
		alpha = global.Alpha()
	}
	trgt := strings.TrimSuffix(global.Path(), ".bmp")
	png, err := os.Create(trgt+"_unfilt.png")
	if err != nil {
		t.Error(err)
		return
	}
	defer png.Close()
	png.Write(magicNumbers)
	png.Write(makeIHDR(w, h, bpp, alpha, false))
	// decodeImgData
	var buf bytes.Buffer
	z, _ := zlib.NewWriterLevel(&buf, 8)
	y0, y1, yDelta := h-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, h, +1
	}
	s := bpp / 8
	for y := y0; y != y1; y += yDelta {
		line, err := scanLine(bmp, w, s)
		if err != nil {
			t.Error(err)
			return
		}
		for x := 0; x < w*s; x += s {
			line[x+0], line[x+2] = line[x+2], line[x+0]
			if s == 4 && !alpha {
				line[x+3] = 0xFF
			}
		}
		filt := typeByte(line, none)
		t.Log(y, len(filt), filt)
		z.Write(filt)
		z.Flush()
		data := buf.Bytes()
		buf.Reset()
		t.Log(data)
		b, err := png.Write(makeIDAT(data))
		t.Log(b)
		if err != nil {
			t.Error(err)
			return
		}
	}
	png.Write(makeIEND())
}