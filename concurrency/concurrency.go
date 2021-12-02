package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io/ioutil"
	"sync"

	"io"
	"log"
	"os"

	"github.com/soniakeys/quant/median"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

type ProcessedImage struct {
	palatted *image.Paletted
	delay    int
	index    int
}

type ByIndex []ProcessedImage

func (a ByIndex) Len() int {
	return len(a)
}

func (a ByIndex) Less(i, j int) bool {
	return a[i].index < a[j].index
}

func (a ByIndex) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// TODO: 順序をなんとかして保証する
func cropGifConcurrent(reader io.Reader, cropStartX, cropStartY, cropSize int) (files []*os.File, err error) {
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
	// eg, ctx := errgroup.WithContext(context.Background())
	wg := &sync.WaitGroup{}

	outGif := &gif.GIF{}
	outGif2 := &gif.GIF{}
	// processedImages := []ProcessedImage{}
	processedImageList := make([]ProcessedImage, len(inputGif.Image))

	for i, srcImg := range inputGif.Image {
		iterator := i
		img := srcImg
		processedImageList[i].index = i
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// 画像に書き込む
			draw.Draw(overpaintImage, overpaintImage.Bounds(), img, image.ZP, draw.Over)

			tempFile, err := ioutil.TempFile(os.TempDir(), "temp")
			defer tempFile.Close()

			if err != nil {
				panic(err)
				// return err
			}

			// ここはencodeする必要ないかも
			err = png.Encode(tempFile, overpaintImage)
			if err != nil {
				fmt.Println("encode: ", err)
				// return err
			}

			_, err = tempFile.Seek(0, 0)
			if err != nil {
				fmt.Println("seek:", err)
				// return err
			}

			pngImg, _, err := image.Decode(tempFile)

			if err != nil {
				fmt.Println("decode: ", err)
				// return err
			}

			_, err = tempFile.Seek(0, 0)
			if err != nil {
				fmt.Println("seek:", err)
				// return err
			}

			cimg := pngImg.(SubImager).SubImage(image.Rect(cropStartX, cropStartY, cropStartX+cropSize, cropStartY+cropSize))

			err = png.Encode(tempFile, cimg)
			if err != nil {
				fmt.Println("encode: ", err)
				// return err
			}

			_, err = tempFile.Seek(0, 0)
			if err != nil {
				fmt.Println("seek:", err)
				// return err
			}

			inGif, _, err := image.Decode(tempFile)

			// 256色を決定
			q := median.Quantizer(256)
			p := q.Quantize(make(color.Palette, 0, 256), inGif)
			paletted := image.NewPaletted(inGif.Bounds(), p)

			// ディザリング
			draw.FloydSteinberg.Draw(paletted, inGif.Bounds(), inGif, image.ZP)

			for y := inGif.Bounds().Min.Y; y < inGif.Bounds().Max.Y; y++ {
				for x := inGif.Bounds().Min.X; x < inGif.Bounds().Max.X; x++ {
					paletted.Set(x, y, inGif.At(x, y))
				}
			}

			outGif.Image = append(outGif.Image, paletted)
			outGif.Delay = append(outGif.Delay, 0)

			processedImageList[i].palatted = paletted
			processedImageList[i].delay = 0

			ns = append(ns, tempFile.Name())
			splitedFiles = append(splitedFiles, tempFile)
		}(iterator)
	}

	// 各フレームごとで並行処理する
	// 	eg.Go(func() error {
	// 		// 画像に書き込む
	// 		draw.Draw(overpaintImage, overpaintImage.Bounds(), img, image.ZP, draw.Over)

	// 		tempFile, err := ioutil.TempFile(os.TempDir(), "temp")
	// 		defer tempFile.Close()

	// 		if err != nil {
	// 			return err
	// 		}

	// 		// ここはencodeする必要ないかも
	// 		err = png.Encode(tempFile, overpaintImage)
	// 		if err != nil {
	// 			fmt.Println("encode: ", err)
	// 			return err
	// 		}

	// 		_, err = tempFile.Seek(0, 0)
	// 		if err != nil {
	// 			fmt.Println("seek:", err)
	// 			return err
	// 		}

	// 		pngImg, _, err := image.Decode(tempFile)

	// 		if err != nil {
	// 			fmt.Println("decode: ", err)
	// 			return err
	// 		}

	// 		_, err = tempFile.Seek(0, 0)
	// 		if err != nil {
	// 			fmt.Println("seek:", err)
	// 			return err
	// 		}

	// 		cimg := pngImg.(SubImager).SubImage(image.Rect(cropStartX, cropStartY, cropStartX+cropSize, cropStartY+cropSize))

	// 		err = png.Encode(tempFile, cimg)
	// 		if err != nil {
	// 			fmt.Println("encode: ", err)
	// 			return err
	// 		}

	// 		_, err = tempFile.Seek(0, 0)
	// 		if err != nil {
	// 			fmt.Println("seek:", err)
	// 			return err
	// 		}

	// 		inGif, _, err := image.Decode(tempFile)

	// 		// 256色を決定
	// 		q := median.Quantizer(256)
	// 		p := q.Quantize(make(color.Palette, 0, 256), inGif)
	// 		paletted := image.NewPaletted(inGif.Bounds(), p)

	// 		// ディザリング
	// 		draw.FloydSteinberg.Draw(paletted, inGif.Bounds(), inGif, image.ZP)

	// 		for y := inGif.Bounds().Min.Y; y < inGif.Bounds().Max.Y; y++ {
	// 			for x := inGif.Bounds().Min.X; x < inGif.Bounds().Max.X; x++ {
	// 				paletted.Set(x, y, inGif.At(x, y))
	// 			}
	// 		}

	// 		outGif.Image = append(outGif.Image, paletted)
	// 		outGif.Delay = append(outGif.Delay, 0)
	// 		processedImages = append(processedImages, ProcessedImage{
	// 			palatted: paletted,
	// 			delay:    0,
	// 			index:    iterator,
	// 		})

	// 		ns = append(ns, tempFile.Name())
	// 		splitedFiles = append(splitedFiles, tempFile)

	// 		select {
	// 		case <-ctx.Done():
	// 			// fmt.Println("Canceled: ", i)
	// 			return nil
	// 		default:
	// 			// fmt.Println("End: ", i)
	// 			return nil
	// 		}
	// 	})

	// }

	// if err := eg.Wait(); err != nil {
	// 	log.Fatal(err)
	// }

	wg.Wait()

	// sort.Sort(ByIndex(processedImages))

	for _, img := range processedImageList {
		outGif2.Image = append(outGif2.Image, img.palatted)
		outGif2.Delay = append(outGif2.Delay, img.delay)
	}

	f, _ := os.OpenFile("out.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, outGif2)

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

	_, err = cropGifConcurrent(file, 0, 0, 100)
	if err != nil {
		log.Fatalf(err.Error())
	}

}
