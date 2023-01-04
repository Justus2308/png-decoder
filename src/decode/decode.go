package decode

import (
	"bytes"
	"compress/zlib"
	"container/list"
	"io"
	"log"
	"os"

	"png-decoder/src/global"
)

var (
	suffix = "_dec.bmp"
)


func Decode() {
	png, err := os.Open(global.Path())
	if err != nil {
		panic(err)
	}
	defer png.Close()
	w, h, bpp, alpha, inter, err := decodeIHDR(png)
	if err != nil {
		panic(err)
	}
	if w == 0 || h == 0 {
		panic("file contains no pixels")
	}
	log.Println(w, h, bpp, alpha, inter)
	if bpp == 8 {
		/*plte*/_, err := decodePLTE(png)
		if err != nil {
			panic(err)
		}
	}
	linkedIdat := list.New()
	idat, err := decodeIDAT(png)
	if err != nil {
		if err == errIEND {
			panic(global.ErrSyntax)
		}
		if err == errUnknownAncChunk {
			log.Println("warning: file contains unsupported ancilliary chunk")
		} else {
			panic(err)
		}
	}
	linkedIdat.PushFront(idat)
	for {
		nextIdat, err := decodeIDAT(png)
		if err != nil {
			if err == errIEND {
				break
			}
			if err == errUnknownAncChunk {
				log.Println("warning: file contains unsupported ancilliary chunk")
				continue
			}
			panic(err)
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
		panic(err)
	}
	defer z.Close()
	inflated := make([][]byte, h)
	for i := 0; i < h; i++ {
		line := make([]byte, w*s+1)
		_, err = z.Read(line)
		if err != nil {
			if err == io.EOF {
				panic(global.ErrTransmission)
			}
			panic(err)
		}
		inflated[i] = line
	}
	log.Println(inflated)
}