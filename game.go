package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"onigoko/model"
	"sync"
)

// Game 实现相关接口，负责游戏的绘制以及输入输出
//游戏流程：
//game接收输入，处理后调用client方法向后端发起通信，后端更新游戏状态后向客户端发送信息
//client接收到游戏消息后，转发至render模块，该模块负责通过消息更新游戏内部状态，
//game模块接受输入继续处理
type Game struct {
	id       int            //房间id号
	state    int            //房间状态
	playerId int            //当前游戏的玩家id
	players  []model.Player //当前游戏的玩家状态列表
	graph    [][]*Piece     //游戏地图
}

type Piece struct {
	mu    sync.RWMutex  //读写锁，避免竞态
	block *model.Block  //内部维护的状态
	image *ebiten.Image //复用的image
}

func (g Game) Update() error {
	//TODO implement me
	panic("implement me")
}

func (g Game) Draw(screen *ebiten.Image) {
	//TODO implement me
	panic("implement me")
}

func (g Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	//TODO implement me
	panic("implement me")
}
