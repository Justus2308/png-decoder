package decode

import (
	"fmt"
	"png-decoder/src/global"
	"png-decoder/src/utils"
)


func reconstruct(filt, prev []byte, w, s int) ([]byte, error) {
	if len(filt) == 0 {
		return []byte{}, nil
	}
	fmt.Println(filt[0])
	switch filt[0] {
	case 0:
		return filt[1:], nil
	case 1:
		return subRecon(filt[1:], w, s), nil
	case 2:
		return upRecon(filt[1:], prev, w, s), nil
	case 3:
		return averageRecon(filt[1:], prev, w, s), nil
	case 4:
		return paethRecon(filt[1:], prev, w, s), nil
	}
	return nil, global.ErrSyntax
}

func subRecon(filt []byte, w, s int) []byte {
	recon := make([]byte, w*s, w*s)
	for x := 0; x < s; x++ {
		recon[x] = filt[x]
	}
	for x := s; x < w*s; x++ {
		recon[x] = filt[x] + recon[x-s]
	}
	return recon
}

func upRecon(filt, prev []byte, w, s int) []byte {
	recon := make([]byte, w*s, w*s)
	for x := 0; x < w*s; x++ {
		recon[x] = filt[x] + prev[x]
	}
	return recon
}

func averageRecon(filt, prev []byte, w, s int) []byte {
	recon := make([]byte, w*s, w*s)
	for x := 0; x < s; x++ {
		recon[x] = filt[x] + (prev[x] >> 1) // right shift by 1 == division by 2
	}
	for x := s; x < w*s; x++ {
		recon[x] = filt[x] + uint8((uint16(recon[x-s]) + uint16(prev[x])) >> 1)
	}
	return recon
}

func paethRecon(filt, prev []byte, w, s int) []byte {
	recon := make([]byte, w*s, w*s)
	for x := 0; x < s; x++ {
		recon[x] = filt[x] + utils.PaethPred(0, prev[x], 0)
	}
	for x := s; x < w*s; x++ {
		recon[x] = filt[x] + utils.PaethPred(recon[x-s], prev[x], prev[x-s])
	}
	return recon
}