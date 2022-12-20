package main

import (
	"onigoko/data"
	game2 "onigoko/game"
)

// Render 渲染器，负责根据接收到的信息修改客户端内部维护的游戏状态
type Render struct {
	game    *game2.Game          //维护游戏状态
	message chan *data.Operation //监听channel，更新内部状态
}

//游戏入口
func main() {

}
