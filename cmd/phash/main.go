package main

import (
	"context"
	"fmt"
	"image"
	"os"
	"strings"

	phash "github.com/enot-style/go-phash"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 || len(args) > 2 {
		usage()
		os.Exit(2)
	}

	img1, err := loadImage(args[0])
	if err != nil {
		fatal(err)
	}
	hash1 := phash.PHash(img1)

	if len(args) == 1 {
		fmt.Printf("%016x\n", hash1)
		return
	}

	img2, err := loadImage(args[1])
	if err != nil {
		fatal(err)
	}
	hash2 := phash.PHash(img2)
	dist := phash.HammingDistance(hash1, hash2)

	fmt.Printf("%d\n", dist)
}

func loadImage(arg string) (image.Image, error) {
	if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
		img, _, err := phash.DownloadAndDecodeAny(context.Background(), arg)
		return img, err
	}

	f, err := os.Open(arg)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := phash.DecodeAny(f)
	return img, err
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: phash <path-or-url> [path-or-url]")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
