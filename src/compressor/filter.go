package compressor

import (
	"bytes"
	"sort"
)

const (
	none = iota
	sub
	up
	average
	paeth
)

var (
	search = bytes.NewBuffer(make([]byte, 32))
	lookahead = bytes.NewBuffer(make([]byte, 258))
)

func sortSlc(row []byte) []byte {
	sorted := row
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	return sorted
}

func min(i1, i2 int8) int8 {
	if i1 > i2 {
		return i1
	}
	return i2
}

func abs128(b byte) int8 {
	if b < 128 {
		return int8(b)
	}
	return -int8(b)+127
}

func absSum(slc []byte, n int) int8 {
	var sum int8 = 0
	sum += abs128(slc[0] - slc[1])
	sum += abs128(slc[n-1] - slc[n-2])

	for i := 1; i < n-1; i++ {
		sum += min(abs128(slc[i] - slc[i-1]), abs128(slc[i] - slc[i+1]))
	}
	return sum
}

func minAbsDiff(slc []byte) []byte {
	sorted := sortSlc(slc)
}

func Filter(row []byte) ([]byte, int) { // returns filtered row and filter index
	
}

func subFltr(row []byte, bpp int) []byte {
	for i, b := range row {

	}
}