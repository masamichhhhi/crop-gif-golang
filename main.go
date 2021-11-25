package main

import (
	"flag"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"

	"io"
	"log"
	"os"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

// 各フレームを切り出し→各フレームの画像を切り抜く
func splitGif(reader io.Reader) (files []*os.File, err error) {
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

	for i, srcImg := range inputGif.Image {
		// 画像に書き込む
		draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.ZP, draw.Over)

		// file, err := os.Create(fmt.Sprintf("%s%d%s", "temp", i, ".jpg"))
		file, err := os.Create(fmt.Sprintf("%s%d%s", "temp", i, ".png"))
		defer file.Close()

		if err != nil {
			return nil, err
		}

		// ここはencodeする必要ないかも
		// err = jpeg.Encode(file, overpaintImage, &jpeg.Options{Quality: 100})
		err = png.Encode(file, overpaintImage)

		if err != nil {
			fmt.Println("encode: ", err)
			return nil, err
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			fmt.Println("seek:", err)
			return nil, err
		}

		// inGif, err := gif.Decode(file)
		inGif, _, err := image.Decode(file)

		cimg := inGif.(SubImager).SubImage(image.Rect(0, 0, 200, 200))

		paletted := image.NewPaletted(cimg.Bounds(), palette.WebSafe)
		for y := cimg.Bounds().Min.Y; y < cimg.Bounds().Max.Y; y++ {
			for x := cimg.Bounds().Min.X; x < cimg.Bounds().Max.X; x++ {
				paletted.Set(x, y, cimg.At(x, y))
			}
		}

		outGif.Image = append(outGif.Image, paletted)
		outGif.Delay = append(outGif.Delay, 0)

		_, err = file.Seek(0, 0)
		if err != nil {
			fmt.Println("seek:", err)
			return nil, err
		}

		// img, _, err := image.Decode(file)

		// if err != nil {
		// 	fmt.Println("decode:", err)
		// 	return nil, err
		// }

		// fso, err := os.Create(fmt.Sprintf("%s%d%s", "out", i, ".jpg"))
		// if err != nil {
		// 	fmt.Println("create:", err)
		// 	return nil, err
		// }

		// cimg := img.(SubImager).SubImage(image.Rect(0, 0, 200, 200))
		// jpeg.Encode(fso, cimg, &jpeg.Options{Quality: 100})

		// if err != nil {
		// 	return nil, err
		// }

		ns = append(ns, file.Name())
		splitedFiles = append(splitedFiles, file)
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

	_, err = splitGif(file)
	if err != nil {
		log.Fatalf(err.Error())
	}
	// fmt.Println(files)

	// for _, f := range files {
	// 	_, err := png.Decode(f)

	// 	if err != nil {
	// 		fmt.Println("decode: ", err)
	// 		return
	// 	}
	// }

	// 	defer f.Close()

	// 	fso, err := os.Create(fmt.Sprintf("out_%d.png", i))

	// 	defer fso.Close()
	// 	cimg := img.(SubImager).SubImage(image.Rect(50, 0, 150, 100))
	// 	png.Encode(fso, cimg)
	// }

}
