package paethAlg


func PaethPred(a, b, c byte) byte {
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