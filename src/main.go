package main

import (
	"png-decoder/src/encode"
	"png-decoder/src/decode"
	"png-decoder/src/global"
)

func main() {
	enc, err := global.Mode()
	if err != nil {
		panic(err)
	}
	if enc {
		encode.Encode()
	} else {
		decode.Decode()
	}
}