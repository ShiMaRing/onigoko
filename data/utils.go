package data

import (
	"bytes"
	"github.com/hajimehoshi/ebiten/v2"
	"image/png"
	"log"
)

type ImageFactory struct {
	maps map[string]*CustomImage
}

type CustomImage struct {
	image  *ebiten.Image
	option *ebiten.DrawImageOptions
}

var imageFactory *ImageFactory

// GetImageFlyweightFactory 单例模式，获取图像工厂
func GetImageFlyweightFactory() *ImageFactory {
	if imageFactory == nil {
		imageFactory = &ImageFactory{
			maps: make(map[string]*CustomImage),
		}
	}
	return imageFactory
}

//输入图片名字，创建对应的image以及option
func generateImage(imageName string) *CustomImage {
	var buf []byte
	switch imageName {
	case "road":
		buf = road
	case "gate":
		buf = gate
	case "mine":
		buf = mine
	case "cage":
		buf = cage
	case "key":
		buf = key
	case "barrier":
		buf = barrier
	case "ghost":
		buf = ghost
	case "p1":
		buf = p1
	case "p2":
		buf = p2
	case "p3":
		buf = p3
	}
	img, err := png.Decode(bytes.NewReader(buf))
	if err != nil {
		log.Fatalln(err)
	}
	image := ebiten.NewImageFromImage(img)
	op := &ebiten.DrawImageOptions{}
	width, height := image.Size()
	scaleX := PIXEL / float64(width)
	scaleY := PIXEL / float64(height)
	op.GeoM.Scale(scaleX, scaleY)
	return &CustomImage{
		image, op,
	}
}

func GetImageByName(imageName string) *CustomImage {
	if img, ok := imageFactory.maps[imageName]; ok {
		return img
	} else {
		//加载文件内容
		imageFactory.maps[imageName] = generateImage(imageName)
		return imageFactory.maps[imageName]
	}
}
