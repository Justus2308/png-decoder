package encode

import "os"

var (
	magicNumbers = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
)

func makePng(stream []byte, trgt string, name string) error {
	f, err := os.Create(trgt+"/"+name+".png")
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(magicNumbers)
	if err != nil {
		return err
	}
	_, err = f.Write(stream)
	if err != nil {
		return err
	}
	return nil
}