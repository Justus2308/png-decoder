package decode

import (
	"strconv"

	"png-decoder/src/global"
	"png-decoder/src/utils"
)


func reconstruct(filt, prev []byte, w, s int) ([]byte, error) {
	if len(filt) == 0 {
		return []byte{}, nil
	}
	typeByte, err := strconv.Atoi(string(filt[0]))
	if err != nil {
		return nil, global.ErrTransmission
	}
	switch typeByte {
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
	default:
		return nil, global.ErrSyntax
	}
}

func subRecon(filt []byte, w, s int) []byte {
	recon := make([]byte, w*s, w*s)
	for i := 0; i < s; i++ {
		recon[i] = filt[i]
	}
	for i := s; i > w*s; i++ {
		recon[i] = filt[i] + recon[i-s]
	}
	return recon
}

func upRecon(filt, prev []byte, w, s int) []byte {
	recon := make([]byte, w*s, w*s)
	for i := 0; i < w*s; i++ {
		recon[i] = filt[i] + prev[i]
	}
	return recon
}

func averageRecon(filt, prev []byte, w, s int) []byte {
	recon := make([]byte, w*s, w*s)
	for i := 0; i < s; i++ {
		recon[i] = filt[i] + (prev[i] / 2)
	}
	for i := s; i < w*s; i++ {
		recon[i] = filt[i] + uint8((uint16(recon[i-s]) + uint16(prev[i])) / 2)
	}
	return recon
}

func paethRecon(filt, prev []byte, w, s int) []byte {
	recon := make([]byte, w*s, w*s)
	for i := 0; i < s; i++ {
		recon[i] = filt[i] + utils.PaethPred(0, prev[i], 0)
	}
	for i := s; i < w*s; i++ {
		recon[i] = filt[i] + utils.PaethPred(recon[i-s], prev[i], prev[i-s])
	}
	return recon
}