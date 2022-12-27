package global

import "flag"

var path = flag.String("src", "", "path to the source image")

func Path() string {
	return *path
}

func SetPath(p string) { // for testing
	*path = p
}