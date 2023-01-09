package encode

import (
	"sort"

	"png-decoder/src/utils"
)

const (
	none uint8 = iota
	sub
	up
	average
	paeth
)


func sortSlc(row []byte) []byte {
	sorted := make([]byte, len(row))
	copy(sorted, row)
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

func typeByte(slc []byte, b byte) []byte { // empty scanlines do not get a filter type byte
	if len(slc) == 0 {
		return slc
	}
	slc = append(slc, 0)
	copy(slc[1:], slc)
	slc[0] = b
	return slc
}

func filter(orig, prev []byte, w, s int) []byte {
	subF := subFilt(orig, w, s)
	upF := upFilt(orig, prev, w, s)
	averageF := averageFilt(orig, prev, w, s)
	paethF := paethFilt(orig, prev, w, s)
	switch lowestScoreID(orig, subF, upF, averageF, paethF) {
	case none:
		return typeByte(orig, none)
	case sub:
		return typeByte(subF, sub)
	case up:
		return typeByte(upF, up)
	case average:
		return typeByte(averageF, average)
	case paeth:
		return typeByte(paethF, paeth)
	}
	return nil
}

func subFilt(orig []byte, w, s int) []byte {
	filt := make([]byte, w*s, w*s)
	for x := 0; x < s; x++ {
		filt[x] = orig[x]
	}
	for x := s; x < w*s; x++ {
		filt[x] = orig[x] - orig[x-s]
	}
	return filt
}

func upFilt(orig, prev []byte, w, s int) []byte {
	filt := make([]byte, w*s, w*s)
	for x := 0; x < w*s; x++ {
		filt[x] = orig[x] - prev[x]
	}
	return filt
}

func averageFilt(orig, prev []byte, w, s int) []byte {
	filt := make([]byte, w*s, w*s)
	for x := 0; x < s; x++ {
		filt[x] = orig[x] - (prev[x] >> 1) // right shift by 1 == division by 2
	}
	for x := s; x < w*s; x++ {
		filt[x] = orig[x] - uint8((uint16(orig[x-s]) + uint16(prev[x])) >> 1)
	}
	return filt
}

func paethFilt(orig, prev []byte, w, s int) []byte {
	filt := make([]byte, w*s, w*s)
	for x := 0; x < s; x++ {
		filt[x] = orig[x] - utils.PaethPred(0, prev[x], 0)
	}
	for x := s; x < w*s; x++ {
		filt[x] = orig[x] - utils.PaethPred(orig[x-s], prev[x], prev[x-s])
	}
	return filt
}