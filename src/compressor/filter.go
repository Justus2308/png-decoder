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

	bpp int
)

func sortSlc(row []byte) []byte {
	sorted := row
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	return sorted
}

func min(i1, i2 uint8) uint8 {
	if i1 > i2 {
		return i1
	}
	return i2
}

func sign128(b byte) int8 {
	if b < 128 {
		return int8(b)
	}
	return -int8(255-b+1)
}

func absSum(slc []byte) int8 {
	var sum int8 = 0
	n := len(slc)

	sum += sign128(slc[0] - slc[1])
	sum += sign128(slc[n-1] - slc[n-2])

	for i := 1; i < n-1; i++ {
		sum += sign128(min(slc[i] - slc[i-1], slc[i] - slc[i+1]))
	}
	return sum
}

func MinAbsDiff(slc []byte) int8 {
	sorted := sortSlc(slc)
	return absSum(sorted)
}

func Filter(row []byte) ([]byte, int) { // returns filtered row and filter index
	if bpp == 8 {
		return row, none
	}
	mad := MinAbsDiff(row)
	return nil, 0
}

func subFltr(orig *([][]byte), r, w int) []byte {
	filt := make([]byte, w)
	filt[0], filt[1], filt[2], filt[3] = (*orig)[r][0], (*orig)[r][1], (*orig)[r][2], (*orig)[r][3]
	for i := 4; i < w; i++ {
		filt[i] = (*orig)[r][i] - (*orig)[r][i-4]
	}
	return filt
}

func upFltr(orig *([][]byte), r, w int) []byte {
	if r == 0 {
		return (*orig)[r]
	}
	filt := make([]byte, w)
	for i := range (*orig)[r] {
		filt[i] = (*orig)[r][i] - (*orig)[r-1][i]
	}
	return filt
}

func averageFltr(orig *([][]byte), r, w int) []byte {
	var prev *[]byte
	if r == 0 {
		zero := make([]byte, w)
		prev = &zero
	} else {
		prev = &(*orig)[r-1]
	}
	filt := make([]byte, w)
	for i := 0; i < 4; i++ {
		filt[i] = (*orig)[r][i] - ((*prev)[i] / 2)
	}
	for i := 4; i < w; i++ {
		filt[i] = (*orig)[r][i] - uint8((uint16((*orig)[r][i-4]) + uint16((*prev)[i])) / 2)
	}
	return filt
}

func paethFltr(orig *([][]byte), r, w int) []byte {
	var prev *[]byte
	if r == 0 {
		zero := make([]byte, w)
		prev = &zero
	} else {
		prev = &(*orig)[r-1]
	}
	filt := make([]byte, w)
	for i := 0; i < 4; i++ {
		filt[i] = (*orig)[r][i] - paethPred(0, (*prev)[i], (*prev)[i-1])
	}
	for i := 4; i < w; i++ {
		filt[i] = (*orig)[r][i] - paethPred((*orig)[r][i-1], (*prev)[i], (*prev)[i-1])
	}
	return filt
}

func paethPred(a, b, c byte) byte {
	p := int16(a) + int16(b) - int16(c)
	pa := absU8(p - int16(a))
	pb := absU8(p - int16(b))
	pc := absU8(p - int16(c))
	var pr byte
	if pa <= pb && pa <= pc {
		pr = a
	} else if pb <= pc {
		pr = b
	} else {
		pr = c
	}
	return pr
}

func absU8(i int16) uint8 {
	if i < 0 {
		return uint8(-i)
	}
	return uint8(i)
}