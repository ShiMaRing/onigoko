package data

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	EmptyImage           *ebiten.Image
	NormalFace, BoldFace font.Face
)

var Images = make(map[string]*ebiten.Image)

func GetImage(p string) (*ebiten.Image, error) {
	if v, ok := Images[p]; ok {
		return v, nil
	}
	if img, err := ReadImage(p); err != nil {
		return nil, err
	} else {
		eimg := ebiten.NewImageFromImage(img)
		Images[p] = eimg
		return eimg, nil
	}
}

var Sounds = make(map[string]*Sound)

func GetSound(p string) (*Sound, error) {
	if v, ok := Sounds[p]; ok {
		return v, nil
	}
	if snd, err := ReadSound(p); err != nil {
		return nil, err
	} else {
		Sounds[p] = snd
		return snd, nil
	}
}
func GetMusic(p string) (*Sound, error) {
	if v, ok := Sounds[p]; ok {
		return v, nil
	}
	if snd, err := ReadMusic(p); err != nil {
		return nil, err
	} else {
		Sounds[p] = snd
		return snd, nil
	}
}

//加载字体
func LoadData() error {
	EmptyImage = ebiten.NewImage(3, 3)
	EmptyImage.Fill(color.White)

	// Load the fonts.
	d, err := ReadFile("fonts/x12y16pxMaruMonica.ttf")
	if err != nil {
		return err
	}
	tt, err := opentype.Parse(d)
	if err != nil {
		return err
	}
	if NormalFace, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    32,
		DPI:     72,
		Hinting: font.HintingFull,
	}); err != nil {
		return err
	}
	d, err = ReadFile("fonts/x12y16pxMaruMonica.ttf")
	if err != nil {
		return err
	}
	tt, err = opentype.Parse(d)
	if err != nil {
		return err
	}
	if BoldFace, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     45,
		Hinting: font.HintingFull,
	}); err != nil {
		return err
	}

	return nil
}
