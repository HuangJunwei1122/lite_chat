package util

import (
	"fmt"
	"io"
)

func MustCopy(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	return err
}

func MustWrite(dst io.Writer, src string) {
	if _, err := io.WriteString(dst, src); err != nil {
		fmt.Println("write string err, ", err)
	}
}
