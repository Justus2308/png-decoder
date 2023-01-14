package decode

import (
	"testing"

	"png-decoder/src/global"
)


var path = "img_test/test_8bpp.png"


func TestSubRecon(t *testing.T) {
	filt := []byte{1, 1, 0, 0, 255, 0, 0, 1, 0}
	// should recon to {1, 0, 0, 255, 1, 0, 1, 255}
	recon, err := reconstruct(filt, nil, 2, 4)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(recon, "should be [1 0 0 255 1 0 1 255]")
}

func TestUpRecon(t *testing.T) {
	prev := []byte{0, 0, 1, 255, 0, 1, 0, 255}
	filt := []byte{2, 1, 0, 255, 0, 1, 255, 1, 0}
	// should recon to {1, 0, 0, 255, 1, 0, 1, 255}
	recon, err := reconstruct(filt, prev, 2, 4)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(recon, "should be [1 0 0 255 1 0 1 255]")
}

func TestAverageRecon(t *testing.T) {
	prev := []byte{0, 0, 1, 255, 0, 1, 0, 255}
	filt := []byte{3, 1, 0, 0, 128, 1, 0, 1, 0}
	// should recon to {1, 0, 0, 255, 1, 0, 1, 255}
	recon, err := reconstruct(filt, prev, 2, 4)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(recon, "should be [1 0 0 255 1 0 1 255]")
}

func TestDecode(t *testing.T) {
	global.Path = path
	t.Log("decoding", global.Path)
	err := Decode()
	if err != nil {
		t.Error(err)
		return
	}
}