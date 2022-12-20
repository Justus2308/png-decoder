package compressor

import "bytes"

const (
	None = iota
	Sub
	Up
	Average
	Paeth
)

var (
	search = bytes.NewBuffer(make([]byte, 32))
	lookahead = bytes.NewBuffer(make([]byte, 258))
)

func Filter(row []byte) ([]byte, int) { // returns filtered row and filter index
	bpp := 
	switch {

	}
}

func SubFltr(row []byte, bpp int) []byte {
	for k, v := range row {

	}
}