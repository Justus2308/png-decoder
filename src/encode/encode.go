package encode

import (
	"bytes"
	"compress/zlib"
	"os"
	"strings"

	"png-decoder/src/global"
)


var (
	suffix = "_enc.png"
)


// TODO: implement adam7 interlacing
func Encode() error {
	bmp, err := os.Open(global.Path())
	if err != nil {
		return err
	}
	defer bmp.Close()
	w, h, bpp, alpha, topDown, err := decodeHeader(bmp)
	if err != nil {
		return err
	}
	if w == 0 || h == 0 {
		return global.ErrNoPixels
	}
	if alpha {
		alpha = global.Alpha()
	}
	trgt := strings.TrimSuffix(global.Path(), ".bmp")
	png, err := os.Create(trgt+suffix)
	if err != nil {
		return err
	}
	defer png.Close()
	png.Write(global.PNG) // magic numbers
	png.Write(makeIHDR(w, h, bpp, alpha, false))
	switch bpp {
	case 8:
		plte, err := getPalette(bmp)
		if err != nil {
			return err
		}
		png.Write(makePLTE(plte))
		if err = decode8BitData(bmp, png, w, h, topDown); err != nil {
			return err
		}
		png.Write(makeIEND())
	case 24:
		if err = decodeImgData(bmp, png, w, h, 3, false, topDown); err != nil {
			return err
		}
		png.Write(makeIEND())
	case 32:
		if err = decodeImgData(bmp, png, w, h, 4, alpha, topDown); err != nil {
			return err
		}
		png.Write(makeIEND())
	default:
		return global.ErrUnsupported
	}
	return nil
}

func decode8BitData(bmp, png *os.File, w, h int, topDown bool) error {
	var buf bytes.Buffer
	z, _ := zlib.NewWriterLevel(&buf, 8)
	y0, y1, yDelta := h-1, -1, -1
	if topDown {
		y0, y1, yDelta = 0, h, +1
	}
	for y := y0; y != y1; y += yDelta {
		line, err := scanLine(bmp, w, 1)
		if err != nil {
			return err
		}
		filt := typeByte(line, none)
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