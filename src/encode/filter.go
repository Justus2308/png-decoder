package encode

import (
	"fmt"
	"sort"

	"png-decoder/src/paethAlg"
)

const (
	none uint8 = iota
	sub
	up
	average
	paeth
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

func absSum(slc []byte) int {
	sum := 0
	n := len(slc)

	sum += int(sign128(slc[0] - slc[1]))
	sum += int(sign128(slc[n-1] - slc[n-2]))

	for i := 1; i < n-1; i++ {
		sum += int(sign128(min(slc[i] - slc[i-1], slc[i] - slc[i+1])))
	}
	return sum
}

func minAbsDiff(slc []byte) int {
	sorted := sortSlc(slc)
	return absSum(sorted)
}

func lowestScoreID(scores... []byte) uint8 {
	id := none
	if len(scores) == 1 {
		return id
	}
	lowest := minAbsDiff(scores[0])
	for i := 1; i < len(scores); i++ {
		if mad := minAbsDiff(scores[i]); mad < lowest {
			lowest = mad
			id = uint8(i)
		}
	}
	return id
}

func prepend(slc []byte, b byte) []byte {
	slc = append(slc, 0)
	copy(slc[1:], slc)
	slc[0] = b
	return slc
}

// TODO: parellelize with goroutines
func Filter(bits *([][]byte), w, h, bpp int) (filtered [][]byte) { // returns filtered row with prepended filter index
	filtered = make([][]byte, h)
	if bpp == 8 {
		for r := 0; r < h; r++ {
			filtered[r] = (*bits)[r]
			filtered[r] = prepend(filtered[r], none)
		}
		return filtered
	}
	for r := 0; r < h; r++ {
		subF := subFltr(bits, r, w)
		upF := upFltr(bits, r, w)
		averageF := averageFltr(bits, r, w)
		paethF := paethFltr(bits, r, w)
		switch lowestScoreID((*bits)[r], subF, upF, averageF, paethF) {
		case none:
			filtered[r] = (*bits)[r]
			filtered[r] = prepend(filtered[r], none)
			fmt.Println(none)
		case sub:
			filtered[r] = subF
			filtered[r] = prepend(filtered[r], sub)
			fmt.Println(sub)
		case up:
			filtered[r] = upF
			filtered[r] = prepend(filtered[r], up)
			fmt.Println(up)
		case average:
			filtered[r] = averageF
			filtered[r] = prepend(filtered[r], average)
			fmt.Println(average)
		case paeth:
			filtered[r] = paethF
			filtered[r] = prepend(filtered[r], paeth)
			fmt.Println(paeth)
		}
	}
	return filtered
}

func subFltr(orig *([][]byte), r, w int) []byte {
	filt := make([]byte, w*4)
	filt[0], filt[1] = (*orig)[r][0], (*orig)[r][1]
	for i := 4; i < w*4; i++ {
		filt[i] = (*orig)[r][i] - (*orig)[r][i-4]
	}
	return filt
}

func upFltr(orig *([][]byte), r, w int) []byte {
	if r == 0 {
		return (*orig)[r]
	}
	filt := make([]byte, w*4)
	for i := 0; i < w*4; i++ {
		filt[i] = (*orig)[r][i] - (*orig)[r-1][i]
	}
	return filt
}

func averageFltr(orig *([][]byte), r, w int) []byte {
	var prev *[]byte
	if r == 0 {
		zero := make([]byte, w*4)
		prev = &zero
	} else {
		prev = &(*orig)[r-1]
	}
	filt := make([]byte, w*4)
	for i := 0; i < 4; i++ {
		filt[i] = (*orig)[r][i] - ((*prev)[i] / 2)
	}
	for i := 4; i < w*4; i++ {
		filt[i] = (*orig)[r][i] - uint8((uint16((*orig)[r][i-4]) + uint16((*prev)[i])) / 2)
	}
	return filt
}

func paethFltr(orig *([][]byte), r, w int) []byte {
	var prev *[]byte
	if r == 0 {
		zero := make([]byte, w*4)
		prev = &zero
	} else {
		prev = &(*orig)[r-1]
	}
	filt := make([]byte, w*4)
	for i := 0; i < 4; i++ {
		filt[i] = (*orig)[r][i] - paethAlg.PaethPred(0, (*prev)[i], 0)
	}
	for i := 4; i < w*4; i++ {
		filt[i] = (*orig)[r][i] - paethAlg.PaethPred((*orig)[r][i-1], (*prev)[i], (*prev)[i-1])
	}
	return filt
}