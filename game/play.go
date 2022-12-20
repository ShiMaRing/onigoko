package game

import (
	"onigoko/data"
	"onigoko/mynet"
	"sync"
)

type Piece struct {
	mu    sync.RWMutex      //读写锁，避免竞态
	block *data.Block       //内部维护的状态
	image *data.CustomImage //内部维护的图
}

type PlayState struct {
	id       int                //房间id号
	playerId int                //当前游戏的玩家id
	players  []data.Player      //当前游戏的玩家状态列表
	graph    [][]*Piece         //游戏地图
	client   mynet.Communicator //通信
}
