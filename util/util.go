package util

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const (
	EndDelim = '\t'
)

func MustCopy(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	return err
}

func PrintStdout(src io.Reader) error {
	input := bufio.NewReader(src)
	var lastLineLen int
	for {
		text, err := input.ReadString(EndDelim)
		if len(text) <= 0 {
			 return nil
		}
		text = text[:len(text) - 1]
		if err != nil {
			return err
		}
		fmt.Printf("\r\b\r%s\r%s\n", strings.Join(make([]string, lastLineLen), " "), text)
		lines := strings.Split(text, "\n")
		lastLineLen = len(lines[len(lines)-1])
	}
}
