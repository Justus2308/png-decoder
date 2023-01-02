package encode

import "png-decoder/src/global"


func Interlace(bits [][]byte, w, h int) [][]byte {
	var inter [][]byte
	for p := 1; p <= 7; p++ {
		for i := 0; i < h; i++ {
			var line []byte
			for j := 0; j < w*4; j += 4 {
				if (global.Adam7()[i%8][(j/4)%8]) == p {
					line = append(line, bits[i][j:j+4]...)
				}
			}
			if len(line) > 0 {
				inter = append(inter, line)
			}
		}
	}
	return inter
}