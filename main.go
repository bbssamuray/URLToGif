package main

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fogleman/gg"
)

const frames = 100
const picX, picY = 800, 96
const shadow bool = true

func draw(textToWrite string, writer io.Writer) {

	start := time.Now()

	var fileName string = "./cache/" + textToWrite

	if textToWrite[len(textToWrite)-4:] == ".gif" {
		textToWrite = textToWrite[:len(textToWrite)-4]
	}

	var wg sync.WaitGroup
	var palettedImages [frames]*image.Paletted
	var GifPalette color.Palette = palette.Plan9

	delays := make([]int, 0)

	dc := gg.NewContext(picX, picY)
	if err := dc.LoadFontFace("ubuntumono.ttf", picY); err != nil {
		panic(err)
	}

	imageRectangle := dc.Image().Bounds()

	var s string = textToWrite + "       "
	var w float64
	for int(w) < picX {
		s = s + " "
		w, _ = dc.MeasureString(s)
	}
	s2 := s + s

	var movementPerFrame float64 = w / float64(frames)
	for i := 0; i < frames; i++ {

		delays = append(delays, 5)

		palettedImages[i] = image.NewPaletted(imageRectangle, GifPalette)
		wg.Add(1)
		go copyToPalleted(palettedImages[i], i, movementPerFrame, s2, &wg)

	}

	wg.Wait()

	var g gif.GIF

	g.Image = palettedImages[:]
	g.Delay = delays
	/*g.LoopCount = -1 // It turns out we don't need any of this! Yay!
	g.Disposal = nil
	g.Config = image.Config{
		GifPalette,
		picX,
		picY,
	}*/

	out, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer out.Close()

	err = gif.EncodeAll(out, &g)
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}

	elapsed := time.Since(start)
	log.Printf("\""+textToWrite+"\" took %s", elapsed)

}

func copyToPalleted(palettedImage *image.Paletted, frameNumber int, movementPerFrame float64, s2 string, wg *sync.WaitGroup) {

	defer wg.Done()

	dc := gg.NewContext(picX, picY)
	if err := dc.LoadFontFace("ubuntumono.ttf", picY); err != nil {
		panic(err)
	}

	dc.SetRGB(1, 1, 1)
	dc.Clear()

	sourceImage := dc.Image()

	if shadow {
		for x := float64(1.0); x >= 0; x -= 0.2 {
			dc.SetRGB(x, x, x)
			dc.DrawString(s2, -float64(frameNumber)*movementPerFrame+x*10, picY/4*3+x*5)
		}
	} else {
		dc.SetRGB(0, 0, 0)
		dc.DrawString(s2, -float64(frameNumber)*movementPerFrame, picY/4*3)
	}

	for y := 0; y < picY; y++ {
		for x := 0; x < picX; x++ {
			palettedImage.Set(x, y, sourceImage.At(x, y))
		}
	}
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path[1:] == "favicon.ico" {

		} else {

			var urla string = r.URL.Path[1:] //This has problems with ? and #
			if len(urla) > 4 {
				if urla[len(urla)-4:] != ".gif" {

					var newURL string = "/" + urla + ".gif"
					http.Redirect(w, r, newURL, http.StatusSeeOther)
					return
				}
				urla = strings.Replace(urla, "-", " ", -1)
				urla = strings.Replace(urla, "/", "?", -1)
				var fileName string = "./cache/" + urla

				_, err := os.Stat(fileName)
				if err == nil {
					fmt.Println("cached -- ", urla)
					http.ServeFile(w, r, fileName)
				} else if os.IsNotExist(err) {
					fmt.Println("request -- ", urla)
					draw(urla, w)
					http.ServeFile(w, r, fileName)
				} else if err != nil {
					fmt.Println(err)
					http.Redirect(w, r, "/help.gif", http.StatusSeeOther)
				}
			} else {
				var newURL string = "/" + urla + ".gif"
				http.Redirect(w, r, newURL, http.StatusSeeOther)
				return
			}

		}
	})

	http.ListenAndServe(":3005", nil)

}
