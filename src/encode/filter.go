package encode

import (
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
	subF := subFltr(orig, w, s)
	upF := upFltr(orig, prev, w, s)
	averageF := averageFltr(orig, prev, w, s)
	paethF := paethFltr(orig, prev, w, s)
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

func subFltr(orig []byte, w, s int) []byte {
	filt := make([]byte, w*s, w*s)
	for i := 0; i < s; i++ {
		filt[i] = orig[i]
	}
	for i := s; i < w*s; i++ {
		filt[i] = orig[i] - orig[i-s]
	}
	return filt
}

func upFltr(orig, prev []byte, w, s int) []byte {
	filt := make([]byte, w*s, w*s)
	for i := 0; i < w*s; i++ {
		filt[i] = orig[i] - prev[i]
	}
	return filt
}

func averageFltr(orig, prev []byte, w, s int) []byte {
	filt := make([]byte, w*s, w*s)
	for i := 0; i < s; i++ {
		filt[i] = orig[i] - (prev[i] / 2)
	}
	for i := s; i < w*s; i++ {
		filt[i] = orig[i] - uint8((uint16(orig[i-s]) + uint16(prev[i])) / 2)
	}
	return filt
}

func paethFltr(orig, prev []byte, w, s int) []byte {
	filt := make([]byte, w*s, w*s)
	for i := 0; i < s; i++ {
		filt[i] = orig[i] - paethAlg.PaethPred(0, prev[i], 0)
	}
	for i := s; i < w*s; i++ {
		filt[i] = orig[i] - paethAlg.PaethPred(orig[i-s], prev[i], prev[i-s])
	}
	return filt
}