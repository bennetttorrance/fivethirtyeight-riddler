package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/png"
	"os"
)

type colorData struct {
	R uint32
	G uint32
	B uint32
}

func (cd colorData) String() string {
	return fmt.Sprintf(`{R:%d G:%d B:%d}`, cd.R, cd.G, cd.B)
}

func newColorData(c color.Color) colorData {
	r, g, b, _ := c.RGBA()
	return colorData{
		// take it from 16- to 4-bit
		R: r >> 8,
		G: g >> 8,
		B: b >> 8,
	}
}

func decodeImage(filename string) (*image.RGBA, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	im, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	rgba, ok := im.(*image.RGBA)
	if !ok {
		return nil, errors.New(`conversion to RGBA failed on ` + filename)
	}
	return rgba, nil
}

func getImageDimensions(im *image.RGBA) (int, int) {
	b := im.Bounds()
	return b.Dx(), b.Dy()
}

func getPixels(im *image.RGBA) (map[colorData]int, error) {
	bnds := im.Bounds().Size()

	clrMap := make(map[colorData]int)
	for i := 0; i < bnds.X; i++ {
		for j := 0; j < bnds.Y; j++ {
			c := im.At(i, j)
			cd := newColorData(c)
			_, ok := clrMap[cd]
			if !ok {
				clrMap[cd] = 1
			} else {
				clrMap[cd]++
			}
		}
	}
	return clrMap, nil
}

func getFilesInFolder(folder string) ([]os.FileInfo, error) {
	f, err := os.Open(folder)
	if err != nil {
		return nil, err
	}
	files, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func compareImages(im1, im2 map[colorData]int) (int, error) {
	im1Ct, im2Ct, inCommon := 0, 0, 0
	for clr, ct1 := range im1 {
		ct2, ok := im2[clr]
		if ok {
			if ct1 < ct2 {
				inCommon += ct1
			} else {
				inCommon += ct2
			}
		}

		im1Ct += ct1
	}
	for _, ct := range im2 {
		im2Ct += ct
	}
	if im1Ct != im2Ct {
		return 0, errors.New(`pixel count between the images are not the same`)
	}
	return inCommon, nil
}

func findClosestMatch(im *image.RGBA, flags []os.FileInfo) (string, int, error) {
	targetW, targetH := getImageDimensions(im)
	bestRank, bestImage := 0, ``
	mysteryPx, err := getPixels(im)
	if err != nil {
		return ``, 0, err
	}
	for _, f := range flags {
		thisFlag, err := decodeImage(`flags/` + f.Name())
		if err != nil {
			fmt.Printf("an error occurred decoding %s: %s skipping...\n", f.Name(), err.Error())
			continue
		}
		w, h := getImageDimensions(thisFlag)
		if w != targetW || h != targetH {
			continue
		}
		thesePx, err := getPixels(thisFlag)
		if err != nil {
			return ``, 0, err
		}
		rank, err := compareImages(mysteryPx, thesePx)
		if err != nil {
			return ``, 0, err
		}
		if rank > bestRank {
			bestRank = rank
			bestImage = f.Name()
		}
	}
	return bestImage, bestRank, nil
}

func main() {
	flags, err := getFilesInFolder(`flags/`)
	if err != nil {
		panic(err)
	}
	inputs, err := getFilesInFolder(`inputData/`)
	if err != nil {
		panic(err)
	}
	for _, in := range inputs {
		im, err := decodeImage(`inputData/` + in.Name())
		if err != nil {
			panic(err)
		}
		m, n, err := findClosestMatch(im, flags)
		if err != nil {
			panic(err)
		}
		fmt.Printf("the closest match to %s is %s with a score of %d\n", in.Name(), m, n)
	}
}
