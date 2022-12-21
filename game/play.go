package game

import (
	"context"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"math"
	"onigoko/data"
	"onigoko/mynet"
	"time"
)

type Piece struct {
	block data.Block        //内部维护的状态
	image *data.CustomImage //内部维护的图
}

type PlayState struct {
	game       *Game
	roomId     int                    //房间id号
	playerId   uint32                 //当前游戏的玩家id
	players    map[uint32]data.Player //当前游戏的玩家状态列表
	graph      [][]*Piece             //游戏地图
	client     mynet.Communicator     //通信器
	ctx        context.Context
	cancelFunc context.CancelFunc
	initWorld  []data.Block //初始设置的block
	keys       []ebiten.Key
	received   chan data.Operation
	count      int
	threshold  int
}

func (p *PlayState) Init() error {
	//初始化,心跳发送
	tick := time.NewTicker(2 * time.Second)
	p.ctx, p.cancelFunc = context.WithCancel(context.Background())
	p.keys = make([]ebiten.Key, 10)
	go func() {
		for true {
			select {
			case <-tick.C:
				op := data.Operation{}
				op.OperationType = data.HEART_BEAT
				op.PlayerId = p.game.PlayerId
			case <-p.ctx.Done():
				tick.Stop()
				return
			}
		}
	}()
	p.graph = make([][]*Piece, data.GraphHeight)
	for i := range p.graph {
		p.graph[i] = make([]*Piece, data.GraphWith)
		for j := range p.graph[i] {
			p.graph[i][j] = &Piece{
				block: data.Block{},
				image: nil,
			}
		}
	}
	p.UpdateWorldBlocks(p.initWorld) //初始化地图
	p.received = p.game.communicator.ReceivedMessage()
	player := p.players[p.playerId]
	if player.Identity == data.GHOST {
		p.threshold = 15
	} else {
		p.threshold = 20
	}
	return nil
}

func (p *PlayState) UpdateWorldBlocks(blocks []data.Block) {
	for i := range blocks {
		block := blocks[i]
		p.graph[block.X][block.Y].block = block //指向block
		//更新视图
		switch block.BlockType {
		case data.ROAD:
			p.graph[block.X][block.Y].image = data.GetImageByName("road")
		case data.BARRIER:
			p.graph[block.X][block.Y].image = data.GetImageByName("barrier")
		case data.GATE:
			p.graph[block.X][block.Y].image = data.GetImageByName("gate")
		case data.MINE:
			p.graph[block.X][block.Y].image = data.GetImageByName("mine")
		case data.KEY:
			p.graph[block.X][block.Y].image = data.GetImageByName("key")
		}
	}
}

// UpdatePlayer 更新用户视图
func (p *PlayState) UpdatePlayer(players []data.Player) {
	for i := range players {
		player := players[i]
		p.players[player.Id] = player
	}
}

func (p *PlayState) Dispose() error {
	//发送退出消息
	op := data.Operation{}
	op.OperationType = data.LEAVE_ROOM //离开房间
	op.RoomId = p.roomId               //发送房间号
	p.game.communicator.SendMessage(op)
	p.cancelFunc()
	return nil
}

//当前操作是否合法
func (p *PlayState) isOperationAvailable(key ebiten.Key) bool {
	//当前玩家一旦死亡或者逃跑则操作非法
	player := p.players[p.playerId]
	if player.IsEscaped || player.IsDead || player.IsDizziness {
		return false
	}
	switch key {
	case ebiten.KeyW, ebiten.KeyA, ebiten.KeyS, ebiten.KeyD:
		return true
	case ebiten.KeyQ, ebiten.KeyE:
		if player.Identity == data.GHOST {
			return false
		} else {
			return true
		}
	}
	//其他非法操作
	return false
}

func (p *PlayState) Update() error {
	//接收用户输入，更新视图
	p.count++
	communicator := p.game.communicator
	//处理移动
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyD) {
		key := inpututil.AppendPressedKeys(p.keys[:0])[0]
		//检查操作合法性
		available := p.isOperationAvailable(key)
		if !available {
			return nil
		}
		switch key {
		case ebiten.KeyA:
			if p.count >= p.threshold {
				//允许移动
				p.count = 0
				communicator.SendMessage(p.CreateOperation(data.LEFT))
			}
		case ebiten.KeyW:
			if p.count >= p.threshold {
				//允许移动
				p.count = 0
				communicator.SendMessage(p.CreateOperation(data.UP))
			}
		case ebiten.KeyS:
			if p.count >= p.threshold {
				//允许移动
				p.count = 0
				communicator.SendMessage(p.CreateOperation(data.DOWN))
			}
		case ebiten.KeyD:
			if p.count >= p.threshold {
				//允许移动
				p.count = 0
				communicator.SendMessage(p.CreateOperation(data.RIGHT))
			}
		}
	}
	//其他操作
	if ebiten.IsKeyPressed(ebiten.KeyQ) || ebiten.IsKeyPressed(ebiten.KeyE) {
		key := inpututil.AppendPressedKeys(p.keys[:0])[0]
		//检查操作合法性
		available := p.isOperationAvailable(key)
		if !available {
			return nil
		}
		switch key {
		case ebiten.KeyQ:
			communicator.SendMessage(p.CreateOperation(data.PLACE_MINE))
		case ebiten.KeyE:
			communicator.SendMessage(p.CreateOperation(data.OPEN_LIGHT))
			go func() {
				//三秒后关灯
				time.Sleep(3 * time.Second)
				communicator.SendMessage(p.CreateOperation(data.CLOSE_LIGHT))
			}()
		}
	}
	var message data.Operation
	select {
	case message = <-p.received:
		//处理后端消息
		switch message.OperationType {
		case data.UPDATE:
			//后端要求更新游戏状态
			p.UpdatePlayer(message.Player)
			p.UpdateWorldBlocks(message.Blocks)
		case data.GAME_END:
			//游戏结束，弹窗，退出游戏，暂时不需要实现
		}
	default:
		return nil
	}
	return nil
}

func (p *PlayState) CreateOperation(opType int) data.Operation {
	return data.Operation{
		RoomId:        p.roomId,
		PlayerId:      p.playerId,
		OperationType: opType,
	}
}

// Draw 渲染
func (p *PlayState) Draw(screen *ebiten.Image) {
	//渲染三部分，当前玩家状态栏，剩余照明次数，存活状态，游戏地图，各个玩家位置状态
	//绘制地图,视野外的地方需要填充为黑暗，暂时不需要实现
	playerLocal := p.players[p.playerId]
	for i := range p.graph {
		for j := range p.graph[i] {
			//绘制图块
			customImage := p.graph[i][j].image
			x := float64(i) * data.PIXEL
			y := float64(j) * data.PIXEL
			option := customImage.Option
			option.GeoM.Translate(x, y)
			//如果当前玩家为鬼的话，不绘制地雷
			if playerLocal.Identity == data.GHOST && p.graph[i][j].block.BlockType == data.MINE {
				continue
			}
			screen.DrawImage(customImage.Image, option)
		}
	}
	//绘制玩家,根据玩家的nickName 选择玩家image
	for _, player := range p.players {
		if player.IsEscaped || player.IsDead {
			continue //不需要绘制
		}
		//检查玩家状态
		customImage := data.GetImageByName(player.NickName)
		dir := player.Direct
		option := customImage.Option
		switch dir {
		case data.DOWN:
		case data.UP:
			option.GeoM.Translate(-float64(customImage.Image.Bounds().Dx())/2, -float64(customImage.Image.Bounds().Dy())/2)
			option.GeoM.Rotate(math.Pi)
			option.GeoM.Translate(data.PIXEL*float64(player.X), data.PIXEL*float64(player.Y))
		case data.LEFT:
			option.GeoM.Translate(-float64(customImage.Image.Bounds().Dx())/2, -float64(customImage.Image.Bounds().Dy())/2)
			option.GeoM.Rotate(math.Pi / 2)
			option.GeoM.Translate(data.PIXEL*float64(player.X), data.PIXEL*float64(player.Y))
		case data.RIGHT:
			option.GeoM.Translate(-float64(customImage.Image.Bounds().Dx())/2, -float64(customImage.Image.Bounds().Dy())/2)
			option.GeoM.Rotate(-math.Pi / 2)
			option.GeoM.Translate(data.PIXEL*float64(player.X), data.PIXEL*float64(player.Y))
		}
		screen.DrawImage(customImage.Image, option)
		//检查是否需要套笼子
		if player.Identity == data.GHOST && player.IsDizziness {
			cageImage := data.GetImageByName("cage")
			screen.DrawImage(cageImage.Image, cageImage.Option)
		}
	}

}
