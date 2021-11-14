package diskread

import (
	"os"
)

func ReadFile(file_path string) (f *os.File) {
	f, err := os.OpenFile(file_path, os.O_CREATE|os.O_RDWR, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	return f
}
