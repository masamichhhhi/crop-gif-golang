package main

import (
	"flag"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"io/ioutil"

	"io"
	"log"
	"os"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

// 各フレームを切り出し→各フレームの画像を切り抜く
func cropGif(reader io.Reader, cropStartX, cropStartY, cropSize int) (files []*os.File, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error while decoding: %s", r)
		}
	}()

	inputGif, err := gif.DecodeAll(reader)
	if err != nil {
		return nil, err
	}

	imgWidth, imgHeight := getGifDimensions(inputGif)

	overpaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(overpaintImage, overpaintImage.Bounds(), inputGif.Image[0], image.ZP, draw.Src)

	var ns []string
	var splitedFiles []*os.File

	outGif := &gif.GIF{}

	for _, srcImg := range inputGif.Image {
		// 画像に書き込む
		draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.ZP, draw.Over)

		tempFile, err := ioutil.TempFile(os.TempDir(), "temp")
		defer tempFile.Close()

		if err != nil {
			return nil, err
		}

		// ここはencodeする必要ないかも
		err = png.Encode(tempFile, overpaintImage)
		if err != nil {
			fmt.Println("encode: ", err)
			return nil, err
		}

		_, err = tempFile.Seek(0, 0)
		if err != nil {
			fmt.Println("seek:", err)
			return nil, err
		}

		pngImg, _, err := image.Decode(tempFile)

		if err != nil {
			fmt.Println("decode: ", err)
			return nil, err
		}

		_, err = tempFile.Seek(0, 0)
		if err != nil {
			fmt.Println("seek:", err)
			return nil, err
		}

		cimg := pngImg.(SubImager).SubImage(image.Rect(cropStartX, cropStartY, cropStartX+cropSize, cropStartY+cropSize))

		err = png.Encode(tempFile, cimg)
		if err != nil {
			fmt.Println("encode: ", err)
			return nil, err
		}

		_, err = tempFile.Seek(0, 0)
		if err != nil {
			fmt.Println("seek:", err)
			return nil, err
		}

		inGif, _, err := image.Decode(tempFile)

		paletted := image.NewPaletted(inGif.Bounds(), palette.WebSafe)
		for y := inGif.Bounds().Min.Y; y < inGif.Bounds().Max.Y; y++ {
			for x := inGif.Bounds().Min.X; x < inGif.Bounds().Max.X; x++ {
				paletted.Set(x, y, inGif.At(x, y))
			}
		}

		outGif.Image = append(outGif.Image, paletted)
		outGif.Delay = append(outGif.Delay, 0)

		ns = append(ns, tempFile.Name())
		splitedFiles = append(splitedFiles, tempFile)
	}

	f, _ := os.OpenFile("out.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, outGif)

	return splitedFiles, nil
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

	_, err = cropGif(file, 200, 200, 500)
	if err != nil {
		log.Fatalf(err.Error())
	}

}
