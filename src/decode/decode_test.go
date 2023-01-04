package decode

import (
	"testing"

	"png-decoder/src/global"
)


var path = "test_images/test_32bpp_transp.png"


func TestDecode(t *testing.T) {
	global.SetPath(path)
	Decode()
}