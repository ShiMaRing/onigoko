package game

import (
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"onigoko/data"
	"onigoko/mynet"
)

// Game 实现相关接口，负责游戏的绘制以及输入输出
//游戏流程：
//game接收输入，处理后调用client方法向后端发起通信，后端更新游戏状态后向客户端发送信息
//client接收到游戏消息后，转发至render模块，该模块负责通过消息更新游戏内部状态，
//game模块接受输入继续处理
type Game struct {
	state        State              //当前的游戏状态，可能为游戏状态，主菜单状态，进入房间状态
	Options      data.Options       //游戏设置
	communicator mynet.Communicator //通信客户端
	PlayerId     uint32
}

var ScreenWidth = data.GraphWith*int(data.PIXEL) + 40
var ScreenHeight = data.GraphHeight*int(data.PIXEL) + 40

func (g *Game) Init() error {
	data.LoadData()
	ebiten.SetScreenFilterEnabled(false)
	ebiten.SetWindowSize(data.GraphWith*int(data.PIXEL)+40, data.GraphHeight*int(data.PIXEL)+40)
	ebiten.SetWindowTitle("Onigoko  ——created by Gaosong Xu")
	//构建通信器，使用接口，协助mock测试
	g.communicator = mynet.NewFakeClient()

	id := uuid.New().ID() //启动通信
	g.PlayerId = id
	go func() {
		if err := g.communicator.Run(); err != nil {
			log.Fatalln(err)
		}
	}()

	if err := g.SetState(&MenuState{
		game: g,
	}); err != nil {
		return err
	}
	return nil
}

func (g *Game) SetState(s State) error {
	if g.state != nil {
		//清空游戏状态
		if err := s.Dispose(); err != nil {
			panic(err)
		}
	}
	//初始化
	if err := s.Init(); err != nil {
		panic(err)
	}
	g.state = s
	return nil
}

func (g *Game) Dispose() error {
	return nil
}

func (g *Game) Update() error {
	return g.state.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.state.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return data.GraphWith*int(data.PIXEL) + 40, data.GraphHeight*int(data.PIXEL) + 40
}

type NoError struct {
}

func (e NoError) Error() string {
	return "no error here captain"
}
