package decode

import (
	"bytes"
	"compress/zlib"
	"container/list"
	"fmt"
	"io"
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
	w, h, bpp, /*inter*/_, err := decodeIHDR(png)
	if err != nil {
		return err
	}
	if w == 0 || h == 0 {
		return global.ErrNoPixels
	}
	if bpp == 8 {
		/*plte*/_, err := decodePLTE(png)
		if err != nil {
			return err
		}
	}
	linkedIdat := list.New()
	idat, err := decodeIDAT(png)
	if err != nil {
		if err == errIEND {
			return global.ErrSyntax
		}
		if err == WarnUnknownAncChunk {
			fmt.Println(err)
		} else {
			return err
		}
	}
	linkedIdat.PushFront(idat)
	for {
		nextIdat, err := decodeIDAT(png)
		if err != nil {
			if err == errIEND {
				break
			}
			if err == WarnUnknownAncChunk {
				fmt.Println(err)
				continue
			}
			return err
		}
		linkedIdat.InsertAfter(nextIdat, linkedIdat.Back())
	}
	var readers []io.Reader
	for e := linkedIdat.Front(); e != nil; e = e.Next() {
		readers = append(readers, bytes.NewReader(e.Value.([]byte)))
	}
	s := bpp / 8
	r := io.MultiReader(readers...)
	z, err := zlib.NewReader(r)
	if err != nil {
		return err
	}
	defer z.Close()

	trgt := strings.TrimSuffix(global.Path, ".png")
	bmp, err := os.Create(trgt+suffix)
	if err != nil {
		return err
	}
	defer bmp.Close()
	bmp.Write(makeV5Header(w, h, s, bpp))

	prev := make([]byte, w*s, w*s)
	for i := 0; i < h; i++ {
		line := make([]byte, w*s+1)
		_, err = z.Read(line) // inflate
		if err != nil && err != io.EOF {
			if err == io.ErrUnexpectedEOF {
				return global.ErrTransmission
			}
			return err
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
					toWrite[i+3] = 0xFF
				}
			}
			bmp.Write(toWrite)
		} else {
			bmp.Write(recon)
		}
	}
	return nil
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
	assign(infoHeader, utils.U32toBLit(54), 10) // offset
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

func makeV5Header(w, h, s, bpp int) []byte { // alpha
	v5Header := make([]byte, 138)
	assign(v5Header, makeInfoHeader(w, h, s, bpp), 0)
	// file header and beginning ob DIB header are identical in BITMAPINFOHEADER and V5INFOHEADER
	assign(v5Header, utils.U32toBLit(uint32(14+124+w*s*h)), 2) // change bmp size
	assign(v5Header, utils.U32toBLit(138), 10) // change offset
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