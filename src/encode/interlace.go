package encode

import "png-decoder/src/global"


func interlace(row []byte, w, r, s, pass int) []byte {
	var inter []byte
	for i := 0; i < w*s; i += s {
		if pass == global.Adam7[r%8][(i/s)%8] {
			inter = append(inter, row[i:i+s]...)
		}
	}
	return inter
}