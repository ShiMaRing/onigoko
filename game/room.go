package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"onigoko/data"
)

// RoomState 进入房间状态
type RoomState struct {
	game                *Game
	currentPlayerNumber int            //当前的玩家数量
	buttons             []*data.Button //按钮，返回按钮
}

// Init 等待房间，只需要显示当前的人数，以及退出房间按钮
func (r *RoomState) Init() error {
	//初始化

	return nil
}

func (r *RoomState) Dispose() error {
	return nil
}

func (r *RoomState) Update() error {

	return nil
}

func (r *RoomState) Draw(screen *ebiten.Image) {

	return
}
