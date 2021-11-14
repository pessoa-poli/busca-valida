package main

import (
	//modelo de importação de packages
	"bufio"
	dr "buscavalida/packages/diskread"
	"fmt"
)

var (
	FILE_PATH = "path/to/my/file.txt"
)

func main() {
	//Aqui montamos o workflow principal
	f := dr.ReadFile(FILE_PATH)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	var text []string

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}

	for _, each_ln := range text {
		fmt.Println(each_ln)
	}
}
