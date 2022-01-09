package a

import "os"

func Test() {
	f, err := os.Open("file")
	if err != nil {
		return
	}
	_, _ = f, err
	f.Close()
	f.Close()
}
