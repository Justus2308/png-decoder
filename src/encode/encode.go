package encode

import (
	"bytes"
	"compress/zlib"
	"os"
	"strings"

	"png-decoder/src/global"
)


var (
	magicNumbers = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
)


func Encode() {
	bmp, err := os.Open(global.Path())
	if err != nil {
		panic(err)
	}
	defer bmp.Close()
	w, h, bpp, alpha, topDown, fileInfoLen, err := decodeHeader(bmp)
	if err != nil {
		panic(err)
	}
	if w == 0 || h == 0 {
		panic("file contains no pixels")
	}
	if alpha {
		alpha = global.Alpha()
	}
	trgt := strings.TrimSuffix(global.Path(), ".bmp")
	png, err := os.Create(trgt+"_enc.png")
	if err != nil {
		panic(err)
	}
	defer png.Close()
	png.Write(magicNumbers)
	png.Write(makeIHDR(w, h, bpp, alpha, false))
	switch bpp {
	case 8:
		fileInfoLen = fileInfoLen
	case 24:
		if err = decodeImgData(bmp, png, w, h, 3, false, topDown); err != nil {
			panic(err)
		}
		png.Write(makeIEND())
	case 32:
		if err = decodeImgData(bmp, png, w, h, 4, alpha, topDown); err != nil {
			panic(err)
		}
		png.Write(makeIEND())
	default:
		panic(errUnsupported)
	}
}

// BUG: same zlib writer has to be used for all image data
// or the stream will be ended preemptively at the first adler32 chunk
func decodeImgData(bmp, png *os.File, w, h, s int, alpha, topDown bool) error {
	prev := make([]byte, w*s, w*s)
	var buf bytes.Buffer
	z, _ := zlib.NewWriterLevel(&buf, 8)
	y0, y1, yDelta := h-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, h, +1
	}
	for y := y0; y != y1; y += yDelta {
		line, err := scanLine(bmp, w, s)
		if err != nil {
			return err
		}
		for x := 0; x < w*s; x += s {
			line[x+0], line[x+2] = line[x+2], line[x+0]
			if s == 4 && !alpha {
				line[x+3] = 0xFF
			}
		}
		if err != nil {
			return err
		}
		filt := filter(line, prev, w, s)
		prev = line
		z.Write(filt) // deflate
		z.Flush()
		data := buf.Bytes()
		buf.Reset()
		_, err = png.Write(makeIDAT(data))
		if err != nil {
			return err
		}
	}
	return nil
}