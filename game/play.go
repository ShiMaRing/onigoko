package game

import (
	"context"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image/color"
	"onigoko/data"
	"onigoko/mynet"
	"strings"
	"time"
)

type Piece struct {
	block data.Block        //内部维护的状态
	image *data.CustomImage //内部维护的图
}

//还差点灯和游戏结束

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
	notify     []string //通知，显示系统的消息
	stateInfo  []string //状态信息，显示所有的玩家状态,只有玩家有 格式
	isGameEnd  bool     //游戏是否结束,需要退出游戏
	endMessage string   //游戏结束的消息
}

func (p *PlayState) Init() error {
	//初始化,心跳发送
	data.GetImageFlyweightFactory()
	p.notify = make([]string, 0)
	p.ctx, p.cancelFunc = context.WithCancel(context.Background())
	p.keys = make([]ebiten.Key, 10)
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
		p.threshold = 10
	} else {
		p.threshold = 15
	}
	//初始化信息列表
	p.stateInfo = make([]string, 0)
	//根据玩家身份确定是否需要显示状态信息
	p.stateInfo = append(p.stateInfo, fmt.Sprintf("You Are %s", strings.ToTitle(player.NickName)))
	if player.Identity == data.HUMAN {
		p.stateInfo = append(p.stateInfo, "Light: 2  (press E)")
		p.stateInfo = append(p.stateInfo, "Trap:  1  (press Q)")
		p.stateInfo = append(p.stateInfo, "")
	}
	p.stateInfo = append(p.stateInfo, "P1: Alive")
	p.stateInfo = append(p.stateInfo, "P2: Alive")
	p.stateInfo = append(p.stateInfo, "P3: Alive")
	p.stateInfo = append(p.stateInfo, "")
	p.stateInfo = append(p.stateInfo, "Notices:")

	return nil
}

func (p *PlayState) UpdateWorldBlocks(blocks []data.Block) {
	for i := range blocks {
		block := blocks[i]
		X := block.X
		Y := block.Y
		p.graph[X][Y].block.BlockType = block.BlockType
		p.graph[X][Y].block.X = block.X
		p.graph[X][Y].block.Y = block.Y
		p.graph[X][Y].block.PlayerId = block.PlayerId
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
	op.PlayerId = p.playerId           //发送玩家id
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
	//不允许重复触发
	if key == ebiten.KeyE && player.IsLighting {
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
	if p.isGameEnd { //游戏结束不再接收更新
		return nil
	}
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
			p.UpdatePlayer(message.Players)
			p.UpdateWorldBlocks(message.Blocks)
			p.UpdateOther(message)
		case data.GAME_END:
			p.isGameEnd = true
			p.endMessage = message.Message
			//三秒后返回主菜单
			go func() {
				time.Sleep(3 * time.Second)
				p.game.SetState(&MenuState{game: p.game}, true)
			}()
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
	postionX := playerLocal.X
	postionY := playerLocal.Y
	roadImage := data.GetImageByName("road")
	for i := range p.graph {
		for j := range p.graph[i] {
			//绘制图块
			customImage := p.graph[i][j].image
			block := p.graph[i][j].block
			//鬼的视野是周围4格子
			//人的视野是周围2格子
			//超出视野的地方需要填充为黑暗
			x := float64(j) * data.PIXEL
			y := float64(i) * data.PIXEL
			//不应该直接修改指针指向的地址

			option := customImage.Option
			roadOption := roadImage.Option

			roadOption.GeoM.Translate(x, y)
			option.GeoM.Translate(x, y)

			//在这里判断是否在视野内,鬼是4格，人是2格，人如果在照明则无视状态
			if playerLocal.Identity == data.GHOST {
				if !(abs(i, postionX) <= 6 && abs(j, postionY) <= 6) {
					//不在视野内,绘制黑暗
					option.ColorM.Scale(0, 0, 0, 1)
					roadOption.ColorM.Scale(0, 0, 0, 1)
				}
			} else if playerLocal.Identity == data.HUMAN && !playerLocal.IsLighting {
				if !(abs(i, postionX) <= 3 && abs(j, postionY) <= 3) {
					//不在视野内,绘制黑暗
					option.ColorM.Scale(0, 0, 0, 1)
					roadOption.ColorM.Scale(0, 0, 0, 1)
				}
			}

			if block.BlockType == data.KEY || block.BlockType == data.MINE {
				screen.DrawImage(roadImage.Image, &roadOption)
			}

			screen.DrawImage(customImage.Image, &option)

			//如果当前玩家为鬼的话，不绘制地雷,只画路
			if playerLocal.Identity == data.GHOST && p.graph[i][j].block.BlockType == data.MINE {
				screen.DrawImage(roadImage.Image, &option)
			}
		}
	}
	//绘制玩家,根据玩家的nickName 选择玩家image
	for _, player := range p.players {
		if player.IsEscaped || player.IsDead {
			continue //不需要绘制
		}
		//检查玩家状态,玩家可以直接修改指针，因为要复用

		customImage := data.GetImageByName(player.NickName)
		option := customImage.Option

		if playerLocal.Identity == data.GHOST {
			if !(abs(player.X, postionX) <= 6 && abs(player.Y, postionY) <= 6) {
				//不在视野内,并且不在照明
				if !player.IsLighting {
					option.ColorM.Scale(0, 0, 0, 1)
				}
			}
		} else if playerLocal.Identity == data.HUMAN {
			if !(abs(player.X, postionX) <= 3 && abs(player.Y, postionY) <= 3) && !playerLocal.IsLighting {
				//不在视野内,绘制黑暗
				option.ColorM.Scale(0, 0, 0, 1)
			}
		}

		option.GeoM.Translate(data.PIXEL*float64(player.Y), data.PIXEL*float64(player.X))

		screen.DrawImage(customImage.Image, &option)

		//需要选择是否覆盖黑暗,如果玩家点灯并且当前玩家是鬼，则可以察觉到
		//检查是否需要套笼子
		if player.Identity == data.GHOST && player.IsDizziness {
			cageImage := data.GetImageByName("cage")
			op := cageImage.Option
			op.GeoM.Translate(data.PIXEL*float64(player.Y), data.PIXEL*float64(player.X))
			op.ColorM = option.ColorM
			screen.DrawImage(cageImage.Image, &op)
		}
	}
	var x = data.GraphWith*int(data.PIXEL) + 10
	var y = 30
	for i := range p.stateInfo {
		data.DrawStaticText(
			p.stateInfo[i],
			data.BoldFace,
			x,
			y,
			color.White,
			screen,
			false,
		)
		y += 25
	}
	for i := range p.notify {
		data.DrawStaticText(
			p.notify[i],
			data.BoldFace,
			x,
			y,
			color.White,
			screen,
			false,
		)
		y += 25
	}
	//判断游戏是否结束，选择绘制最终的结果
	if p.isGameEnd {
		//绘制最终结果
		data.DrawStaticText(
			p.endMessage,
			data.NormalFace,
			data.GraphWith*int(data.PIXEL)/2,
			data.GraphHeight*int(data.PIXEL)/2,
			color.White,
			screen,
			true,
		)
	}
}
func abs(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

//UpdateOther	 这里做一些清理或者辅助工作
func (p *PlayState) UpdateOther(message data.Operation) {
	//如果鬼被晕眩，需要延时发送晕眩解除
	for i := range message.Players {
		if message.Players[i].Identity == data.GHOST && message.Players[i].IsDizziness {
			go func() {
				time.Sleep(3 * time.Second)
				p.game.communicator.SendMessage(p.CreateOperation(data.RECOVER_FROM_DIZZINESS))
			}()
			break
		}
	}
	//更新一下Notify,Notify是一个队列，每次只显示最新的5条,
	if message.Message != "" { //如果有消息
		p.notify = append(p.notify, message.Message)
		if len(p.notify) > 5 {
			p.notify = p.notify[1:]
		}
	}
	player := p.players[p.playerId]
	//更新一下玩家的状态栏
	for i := range message.Players {
		tmp := message.Players[i]
		if tmp.Id == p.playerId && player.Identity == data.HUMAN {
			//需要更新一下Light 和 Trap
			p.stateInfo[1] = fmt.Sprintf("Light: %d  (press E)", tmp.Lights)
			p.stateInfo[2] = fmt.Sprintf("Trap:  %d  (press Q)", tmp.Mines)
			break
		}
	}
	base := len(p.stateInfo) - 5
	for i := range message.Players {
		tmp := message.Players[i]
		if tmp.IsEscaped {
			switch tmp.NickName {
			case "p1":
				p.stateInfo[base] = "P1 Escaped"
			case "p2":
				p.stateInfo[base+1] = "P2 Escaped"
			case "p3":
				p.stateInfo[base+2] = "P3 Escaped"
			}
		}
		if tmp.IsDead {
			switch tmp.NickName {
			case "p1":
				p.stateInfo[base] = "P1 Dead"
			case "p2":
				p.stateInfo[base+1] = "P2 Dead"
			case "p3":
				p.stateInfo[base+2] = "P3 Dead"
			}
		}
	}

}
