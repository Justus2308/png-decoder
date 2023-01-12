package decode

import (
	"bytes"
	"compress/zlib"
	"container/list"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"png-decoder/src/global"
	"png-decoder/src/utils"
)

var (
	suffix = "_dec.bmp"
)


func Decode() error {
	png, err := os.Open(global.Path)
	if err != nil {
		return err
	}
	defer png.Close()
	w, h, bpp, inter, err := decodeIHDR(png)
	if err != nil {
		return err
	}
	if inter {
		return errors.New("interlacing is not supported yet")
	}
	// log.Println(w, h, bpp, inter)
	if w == 0 || h == 0 {
		return global.ErrNoPixels
	}
	s := bpp / 8
	trgt := strings.TrimSuffix(global.Path, ".png")
	bmp, err := os.Create(trgt+suffix)
	if err != nil {
		return err
	}
	defer bmp.Close()

	pal := false
	switch bpp {
	case 8:
		pal = true
		var plte []byte
		for {
			plte, err = decodeNext(png, pal)
			if err != nil {
				if err == isPLTE {
					break
				}
				if err == isIEND {
					return global.ErrSyntax
				}
				if errors.Is(err, warnUnknownAncChunk) {
					log.Println("[WARNING]", err)
				}
				return err
			}
		}
		bmp.Write(makeInfoHeaderPaletted(w, h, s, bpp))
		palette, err := makePalette(plte)
		if err != nil {
			return err
		}
		bmp.Write(palette)
	case 24:
		bmp.Write(makeInfoHeader(w, h, s, bpp))
	case 32:
		if global.Alpha {
			bmp.Write(makeV5Header(w, h, s, bpp))
		} else {
			bmp.Write(makeInfoHeader(w, h, s, bpp))
		}
	}
	concIdat, err := concatenateIDATs(png, pal)
	if err != nil {
		return err
	}
	var readers []io.Reader
	for e := concIdat.Front(); e != nil; e = e.Next() {
		readers = append(readers, bytes.NewReader(e.Value.([]byte)))
	}
	r := io.MultiReader(readers...)
	z, err := zlib.NewReader(r)
	if err != nil {
		return err
	}
	defer z.Close()

	prev := make([]byte, w*s, w*s)
	for i := 0; i < h; i++ {
		line := make([]byte, w*s+1)
		_, err := z.Read(line) // inflate
		if err != nil && err != io.EOF {
			if err == io.ErrUnexpectedEOF {
				return global.ErrTransmission
			}
			return err
		}
		if err == io.EOF && i-1 != h {
			return global.ErrTransmission
		}
		recon, err := reconstruct(line, prev, w, s)
		if err != nil {
			return err
		}
		prev = recon
		if bpp != 8 {
			toWrite := make([]byte, w*s, w*s)
			copy(toWrite, recon)
			for i := 0; i < w*s; i+=s {
				toWrite[i+0], toWrite[i+2] = toWrite[i+2], toWrite[i+0] // flip RGB to BGR
				if bpp == 32 && !global.Alpha {
					toWrite[i+3] = 0xFF // make alpha channel fully opaque
				}
			}
			bmp.Write(toWrite)
		} else {
			bmp.Write(recon)
		}
	}
	return nil
}

func makePalette(plte []byte) ([]byte, error) {
	palette := make([]byte, 256*4)
	for i, j := 0, 0; i < len(plte); i, j = i+3, j+4 {
		if j >= len(palette) {
			return nil, global.ErrTransmission
		}
		palette[j+0] = plte[i+2]
		palette[j+1] = plte[i+1]
		palette[j+2] = plte[i+0]
	}
	return palette, nil
}

func concatenateIDATs(png *os.File, pal bool) (*list.List, error) {
	concIdat := list.New()
	next, err := decodeNext(png, pal)
	if err != isIDAT {
		if err == isIEND {
			return nil, global.ErrSyntax
		}
		if errors.Is(err, warnUnknownAncChunk) {
			log.Println("[WARNING]", err)
		} else {
			return nil, err
		}
	}
	concIdat.PushFront(next)
	for {
		next, err := decodeNext(png, pal)
		if err != isIDAT {
			if err == isIEND {
				break
			}
			if errors.Is(err, warnUnknownAncChunk)  {
				log.Println("[WARNING]", err)
				continue
			}
			return nil, err
		}
		concIdat.InsertAfter(next, concIdat.Back())
	}
	return concIdat, nil
}

func assign(slc, a []byte, i int) []byte {
	for j := 0; j < len(a); j++ {
		slc[i+j] = a[j]
	}
	return slc
}

func makeInfoHeader(w, h, s, bpp int) []byte { // no alpha
	infoHeader := make([]byte, 54)
	// file header
	assign(infoHeader, global.BMP, 0) // magic numbers
	assign(infoHeader, utils.U32toBLit(uint32(14+40+w*s*h)), 2) // bmp size
	assign(infoHeader, []byte{0x00, 0x00, 0x00, 0x00}, 6) // reserved
	assign(infoHeader, utils.U32toBLit(14+40), 10) // offset
	// DIB header
	assign(infoHeader, utils.U32toBLit(40), 14) // header size
	assign(infoHeader, utils.U32toBLit(uint32(w)), 18) // width
	assign(infoHeader, utils.CompLit(utils.U32toBLit(uint32(h))), 22) // -height
	assign(infoHeader, utils.U16toBLit(1), 26) // planes
	assign(infoHeader, utils.U16toBLit(uint16(bpp)), 28) // bit count
	assign(infoHeader, utils.U32toBLit(0), 30) // compression
	assign(infoHeader, utils.U32toBLit(0), 34) // image size
	assign(infoHeader, utils.U32toBLit(0), 38) // horizontal resolution
	assign(infoHeader, utils.U32toBLit(0), 42) // vertical resolution
	assign(infoHeader, utils.U32toBLit(0), 46) // number of colors in palette
	assign(infoHeader, utils.U32toBLit(0), 50) // number of important colors
	return infoHeader
}

func makeInfoHeaderPaletted(w, h, s, bpp int) []byte {
	pInfoHeader := make([]byte, 14+40)
	assign(pInfoHeader, makeInfoHeader(w, h, s, bpp), 0)
	assign(pInfoHeader, utils.U32toBLit(uint32(14+40+256*4+w*s*h)), 2) // change bmp size
	assign(pInfoHeader, utils.U32toBLit(14+40+256*4), 10) // change offset
	return pInfoHeader
}

func makeV5Header(w, h, s, bpp int) []byte { // alpha
	v5Header := make([]byte, 14+124)
	assign(v5Header, makeInfoHeader(w, h, s, bpp), 0)
	// file header and beginning ob DIB header are identical in BITMAPINFOHEADER and V5INFOHEADER
	assign(v5Header, utils.U32toBLit(uint32(14+124+w*s*h)), 2) // change bmp size
	assign(v5Header, utils.U32toBLit(14+124), 10) // change offset
	assign(v5Header, utils.U32toBLit(124), 14) // change header size
	// extended v5 DIB header
	assign(v5Header,[]byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}, 54) // colour masks, are ignored in uncompressed BMPs
	assign(v5Header, utils.U32toBLit(1), 70) // colour space type: LCS_sRGB
	assign(v5Header,[]byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}, 74) // colour endpoints, are ignored if CSType is not LCS_CALIBRATED_RGB
	assign(v5Header,[]byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}, 110) // colour curves, are ignored if CSType is not LCS_CALIBRATED_RGB
	assign(v5Header, []byte{0x00, 0x00, 0x00, 0x00}, 122) // intent: LCS_GM_ABS_COLORIMETRIC
	assign(v5Header,[]byte{
		0x00, 0x00, 0x00, 0x00, // profile data, ignored unless CSType is PROFILE_...
		0x00, 0x00, 0x00, 0x00, // profile size, ignored unless CSType is PROFILE_...
		0x00, 0x00, 0x00, 0x00, // reserved
	}, 126)
	return v5Header
}