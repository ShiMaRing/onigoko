package game

import (
	"image/color"
	"onigoko/data"
	"onigoko/data/assets/lang"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// MenuState 主菜单
type MenuState struct {
	game             *Game
	backgroundImage  *ebiten.Image
	title            string
	titleImage       *ebiten.Image
	buttons          []*data.Button
	backgroundFadeIn int
	shouldQuit       bool
}

// Init 主菜单状态初始化
func (s *MenuState) Init() error {
	if img, err := data.ReadImage("/ui/logo.png"); err == nil {
		s.titleImage = ebiten.NewImageFromImage(img)
	} else {
		panic(err)
	}

	if img, err := data.ReadImage("/ui/title.png"); err == nil {
		s.backgroundImage = ebiten.NewImageFromImage(img)
	} else {
		panic(err)
	}

	x := ScreenWidth / 2
	y := int(float64(ScreenHeight) / 3)
	startGameButton := data.NewButton(
		x,
		y,
		lang.StartGame,
		func() {
			s.game.SetState(&RoomState{
				game: s.game,
			}, true)
		},
	)
	startGameButton.Hover = true
	y += startGameButton.Image().Bounds().Dy() * 3
	exitButton := data.NewButton(
		x,
		y,
		lang.LeaveGame,
		func() {
			os.Exit(0)
		},
	)
	exitButton.Hover = true
	s.buttons = []*data.Button{
		startGameButton,
		exitButton,
	}

	return nil
}

func (s *MenuState) Dispose() error {
	return nil
}

func (s *MenuState) Update() error {
	if s.shouldQuit {
		return NoError{}
	}
	for _, button := range s.buttons {
		button.Update()
	}
	return nil
}

func (s *MenuState) Draw(screen *ebiten.Image) {
	titleOp := &ebiten.DrawImageOptions{}
	titleOp.GeoM.Scale(0.7, 0.7)
	s.backgroundImage.Fill(color.Black)
	screen.DrawImage(s.backgroundImage, titleOp)
	data.DrawStaticText(
		s.title,
		data.BoldFace,
		ScreenWidth/3,
		ScreenHeight/3,
		color.White,
		screen,
		true,
	)

	// Draw our real title.
	top := &ebiten.DrawImageOptions{}
	top.GeoM.Scale(0.4, 0.4)
	top.GeoM.Translate(float64(ScreenWidth/2-s.titleImage.Bounds().Dx()/5), 16)
	screen.DrawImage(s.titleImage, top)

	// Draw game buttons
	for _, button := range s.buttons {
		button.Draw(screen, &ebiten.DrawImageOptions{})
	}
}
