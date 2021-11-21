package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"log"
	"os"
)

func splitGif(reader io.Reader) (names []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error while decoding: %s", r)
		}
	}()

	gif, err := gif.DecodeAll(reader)
	if err != nil {
		return []string{""}, err
	}

	imgWidth, imgHeight := getGifDimensions(gif)

	overpaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(overpaintImage, overpaintImage.Bounds(), gif.Image[0], image.ZP, draw.Src)

	var ns []string

	for i, srcImg := range gif.Image {
		draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.ZP, draw.Over)

		file, err := os.Create(fmt.Sprintf("%s%d%s", "temp", i, ".png"))

		if err != nil {
			return []string{""}, err
		}

		err = png.Encode(file, overpaintImage)
		if err != nil {
			return []string{""}, err
		}

		ns = append(ns, file.Name())
		file.Close()
	}

	return ns, nil
}

func getGifDimensions(gif *gif.GIF) (x, y int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int

	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}

		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}

		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}

		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}
	return highestX - lowestX, highestY - lowestY
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, os.ErrInvalid)
		return
	}

	// open image file
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	defer file.Close()

	fmt.Println(flag.Arg(0))

	names, err := splitGif(file)
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println(names)
}
