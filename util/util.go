package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func MustCopy(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	return err
}

func MustWrite(dst io.Writer, src string) {
	input := bufio.NewReader()
	for {
		text, err := input.ReadString('\n')
		if err != nil {
			printWrong()
			return
		}
}
