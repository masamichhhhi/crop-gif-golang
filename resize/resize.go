package main

import (
	"image"
	"image/draw"
	"image/gif"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

type Result struct {
	Paletted *image.Paletted
	Thumb    *Thumb
	Position int
}

type Thumb struct {
	Image  *image.NRGBA
	Width  int
	Height int
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	inputFile, err := os.Open(os.Args[2])

	dimension := os.Args[1]

	splits := strings.Split(dimension, "x")
	width, _ := strconv.Atoi(splits[0])
	height, err := strconv.Atoi(splits[1])

	if err != nil {
		panic(err)
	}

	defer inputFile.Close()

	g, err := gif.DecodeAll(inputFile)
	if err != nil {
		panic(err)
	}

	length := len(g.Image)
	done := make(chan *Result)
	images := make([]*image.Paletted, length)

	processed := 0

	for i := range g.Image {
		go func(paletted *image.Paletted, width int, height int, position int) {
			done <- &Result{
				Thumb: &Thumb{
					Image:  imaging.Thumbnail(paletted, width, height, imaging.NearestNeighbor),
					Width:  width,
					Height: height,
				},
				Position: position,
				Paletted: image.NewPaletted(image.Rect(0, 0, width, height), paletted.Palette),
			}
		}(g.Image[i], width, height, i)
	}

	for {
		result := <-done

		draw.Draw(result.Paletted, image.Rect(0, 0, result.Thumb.Width, result.Thumb.Height), result.Thumb.Image, image.Pt(0, 0), draw.Src)

		images[result.Position] = result.Paletted
		processed++

		if processed == length {
			break
		}
	}

	g.Image = images
	outputFile, err := os.Create(os.Args[3])

	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	err = gif.EncodeAll(outputFile, g)
	if err != nil {
		panic(err)
	}
}
