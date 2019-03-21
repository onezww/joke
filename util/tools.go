package util

import (
	"bufio"
	"fmt"
	"os"

	iconv "github.com/djimenez/iconv-go"
)

func IsPathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	return true
}

func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		println("it is no file")
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func WriteLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func Gb2312ToUtf8(obj string) (string, error) {
	converter, err := iconv.NewConverter("gb2312", "utf-8")
	if err != nil {
		return "", err
	}

	data, err := converter.ConvertString(obj)
	return data, err
}
