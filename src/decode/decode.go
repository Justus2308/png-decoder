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
	w, h, bpp, inter, err := decodeIHDR(png)
	if err != nil {
		return err
	}
	if w == 0 || h == 0 {
		return global.ErrNoPixels
	}
	fmt.Println(w, h, bpp, inter)
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
	bmp.Write(makeBMPHeader(w, h, s, bpp))

	prev := make([]byte, w*s)
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
		if bpp == 32 && !global.Alpha {
			toWrite := make([]byte, w*s)
			copy(toWrite, recon)
			for i := 3; i < w*s; i+=4 {
				toWrite[i] = 0xFF
			}
			bmp.Write(toWrite)
		} else {
			bmp.Write(recon)
		}
	}
	return nil
}

func makeBMPHeader(w, h, s, bpp int) []byte {
	infoHeader := global.BMP // magic numbers
	infoHeader = append(infoHeader, utils.U32toBLit(uint32(14+40+w*s*h))...) // bmp size
	infoHeader = append(infoHeader, []byte{0x00, 0x00, 0x00, 0x00}...) // reserved
	infoHeader = append(infoHeader, utils.U32toBBig(54)...) // offset
	infoHeader = append(infoHeader, utils.U16toBLit(40)...) // header size
	infoHeader = append(infoHeader, utils.U32toBLit(uint32(w))...) // width
	infoHeader = append(infoHeader, utils.U32toBLit(uint32(h))...) // height
	infoHeader = append(infoHeader, utils.U16toBLit(1)...) // planes
	infoHeader = append(infoHeader, utils.U16toBLit(uint16(bpp))...) // bit count
	infoHeader = append(infoHeader, utils.U32toBLit(0)...) // compression
	infoHeader = append(infoHeader, utils.U32toBLit(0)...) // image size
	infoHeader = append(infoHeader, utils.U32toBLit(0)...) // horizontal resolution
	infoHeader = append(infoHeader, utils.U32toBLit(0)...) // vertical resolution
	infoHeader = append(infoHeader, utils.U32toBLit(0)...) // number of colors in palette
	infoHeader = append(infoHeader, utils.U32toBLit(0)...) // number of important colors
	return infoHeader
}