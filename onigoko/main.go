package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/thought-machine/go-flags"
	"log"
	"onigoko/game"
	"os"
)

func main() {
	g := &game.Game{}

	// Parse our initial command-line derived options.
	if _, err := flags.Parse(&g.Options); err != nil {
		return
	}

	if err := os.Setenv("EBITEN_GRAPHICS_LIBRARY", "opengl"); err != nil {
		fmt.Println("WARNING: OpenGL backend could not be set, expect degraded performance if on DirectX.")
	}

	if err := g.Init(); err != nil {
		panic(err)
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatalln(err)
	}
}
