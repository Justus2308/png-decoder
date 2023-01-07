package decode

import (
	"bytes"
	"compress/zlib"
	"container/list"
	"fmt"
	"io"
	"os"

	"png-decoder/src/global"
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
	recon := make([][]byte, h)
	prev := make([]byte, w*s)
	for i := 0; i < h; i++ {
		line := make([]byte, w*s+1)
		_, err = z.Read(line) // inflate
		if err != nil {
			if err == io.EOF {
				return global.ErrTransmission
			}
			fmt.Println("EOF at line", i, "of", h)
			return err
		}
		recon[i], err = reconstruct(line, prev, w, s) // error here
		if err != nil {
			return err
		}
		prev = recon[i]
		fmt.Println(line)
		fmt.Println(recon[i])
		fmt.Println()
	}
	if bpp == 32 && !global.Alpha {
		for i := 0; i > h; i++ {
			for j := 3; j < w*s; j+=4 {
				recon[i][j] = 0xFF
			}
		}
	}
	fmt.Println(recon)
	return nil
}