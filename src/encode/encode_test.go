package encode

import (
	"bytes"
	"compress/zlib"
	"os"
	"strings"
	"testing"

	"png-decoder/src/global"
)

var path = "img_test/test_32bpp_transp.bmp"


func TestEncode(t *testing.T) {
	global.Path = path
	err := Encode()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestEncodeUnfiltered(t *testing.T) { // works for 24 and 32 bpp
	global.Path = path
	bmp, err := os.Open(global.Path)
	if err != nil {
		t.Error(err)
		return
	}
	defer bmp.Close()
	w, h, bpp, alpha, topDown, err := decodeHeader(bmp)
	if err != nil {
		t.Error(err)
		return
	}
	if bpp == 8 {
		t.Log("[ERROR] test does not work for 8bpp bmp images")
		return
	}
	if w == 0 || h == 0 {
		t.Error(global.ErrNoPixels)
		return
	}
	if alpha {
		alpha = global.Alpha
	}
	trgt := strings.TrimSuffix(global.Path, ".bmp")
	png, err := os.Create(trgt+"_unfilt.png")
	if err != nil {
		t.Error(err)
		return
	}
	defer png.Close()
	png.Write(global.PNG)
	png.Write(makeIHDR(w, h, bpp, alpha, false))
	// decodeImgData
	s := bpp / 8
	var buf bytes.Buffer
	z, _ := zlib.NewWriterLevel(&buf, 8)
	y0, y1, yDelta := h-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, h, +1
	}
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

func TestEncodeInterlaced(t *testing.T) {
	global.Path = path
	global.Inter = true
	suffix = "_inter.png"
	err := Encode()
	if err != nil {
		t.Error(err)
		return
	}
}