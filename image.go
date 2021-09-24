package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"

	ico "github.com/Kodeworks/golang-image-ico"
	"github.com/apex/log"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gomonobold"
)

func getTextIcon(text string) ([]byte, error) {
	img, err := makeImage(text)
	if err != nil {
		return nil, err
	}
	w := &bytes.Buffer{}
	err = ico.Encode(w, img)
	if err != nil {
		log.Errorf("cant encode ico", err)
		return nil, err
	}
	return w.Bytes(), nil
}

var (
	red   = color.RGBA{255, 0, 0, 255}
	blue  = color.RGBA{0, 0, 255, 255}
	white = color.RGBA{255, 255, 255, 255}
	black = color.RGBA{0, 0, 0, 255}
)

func makeImage(label string) (image.Image, error) {
	var (
		img      = image.NewRGBA(image.Rect(0, 0, 32, 32))
		fontSize = 24.0
	)

	draw.Draw(img, img.Bounds(), image.NewUniform(white), image.ZP, draw.Src)
	ctx := freetype.NewContext()
	f, err := truetype.Parse(gomonobold.TTF)
	if err != nil {
		log.Errorf("cant parse font", err)
		return nil, err
	}
	ctx.SetFont(f)
	ctx.SetDPI(72) // ?
	ctx.SetFontSize(fontSize)
	ctx.SetDst(img)
	ctx.SetSrc(image.NewUniform(blue))
	ctx.SetClip(img.Bounds())
	/*
		pt := freetype.Pt(x, y+int(ctx.PointToFixed(fontSize)>>6))
		if _, err := ctx.DrawString(label, pt); err != nil {
			return nil, err
		}
	*/

	// Draw the text to the background
	pt := freetype.Pt(6, 1+int(ctx.PointToFixed(fontSize)>>6))

	// not all utf8 fonts are supported by wqy-zenhei.ttf
	// use your own language true type font file if your language cannot be printed

	if _, err := ctx.DrawString(label, pt); err != nil {
		log.Errorf("cant draw label", err)
		return nil, err
	}
	pt.Y += ctx.PointToFixed(fontSize * 1.5) // spacing

	return img, nil

}
