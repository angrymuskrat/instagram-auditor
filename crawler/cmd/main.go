package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/angrymuskrat/instagram-auditor/crawler"
	"os"
)

func main() {
	crawlerConfig := flag.String("cc", "crawler.toml", "path to db storage configuration file")
	flag.Parse()
	//ids := readIds()
	if ids == nil {
		return
	}
	cr := crawler.New(context.Background(), *crawlerConfig)
	broken := cr.Start(context.Background(), ids)
	writeBroken(broken)
}

func readIds() []string {
	readFile, err := os.Open("ids.txt")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileTextLines []string

	for fileScanner.Scan() {
		fileTextLines = append(fileTextLines, fileScanner.Text())
	}
	return fileTextLines
}

func writeBroken(b []string) {
	f, err := os.Create("broken.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, s := range b {
		_, err := f.WriteString(s + "\n")
		if err != nil {
			return
		}
	}
}

