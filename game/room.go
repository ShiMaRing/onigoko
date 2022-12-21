package game

import (
	"context"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"onigoko/data"
	"time"
)

// RoomState 进入房间状态
type RoomState struct {
	game                *Game
	currentPlayerNumber int            //当前的玩家数量
	buttons             []*data.Button //按钮，返回按钮
	roomId              int
	receivedMessage     chan data.Operation
	ctx                 context.Context
	cancelFunc          context.CancelFunc
}

// Init 等待房间，只需要显示当前的人数，以及退出房间按钮
func (r *RoomState) Init() error {
	//绘制界面
	//初始化，申请加入游戏

	operation := data.Operation{}
	operation.OperationType = data.JOIN_ROOM
	if r.roomId != 0 {
		operation.RoomId = r.roomId
	}
	operation.PlayerId = r.game.PlayerId
	communicator := r.game.communicator
	communicator.SendMessage(operation)
	r.receivedMessage = communicator.ReceivedMessage()

	tick := time.NewTicker(2 * time.Second)
	r.ctx, r.cancelFunc = context.WithCancel(context.Background())
	go func() {
		for true {
			select {
			case <-tick.C:
				op := data.Operation{}
				op.OperationType = data.HEART_BEAT
				op.PlayerId = r.game.PlayerId
			case <-r.ctx.Done():
				tick.Stop()
				return
			}
		}
	}()
	x := ScreenWidth / 2
	y := int(float64(ScreenHeight)/3) + 40
	goBackButton := data.NewButton(
		x,
		y,
		"leave room",
		func() {
			r.game.SetState(&MenuState{
				game: r.game,
			})
		},
	)
	goBackButton.Hover = true
	r.buttons = []*data.Button{
		goBackButton,
	}
	return nil
}

func (r *RoomState) Dispose() error {
	//发送退出消息
	op := data.Operation{}
	op.OperationType = data.LEAVE_ROOM //离开房间
	op.RoomId = r.roomId               //发送房间号
	r.game.communicator.SendMessage(op)
	r.cancelFunc()
	return nil
}

func (r *RoomState) Update() error {
	//不断读取消息
	for _, button := range r.buttons {
		button.Update()
	}
	select {
	case operation := <-r.receivedMessage:
		switch operation.OperationType {
		//检查消息类型
		case data.GAME_START:
			//游戏开始前准备
			playState := PlayState{}
			playState.players = make(map[uint32]data.Player)
			for i := range operation.Player {
				playState.players[operation.Player[i].Id] = operation.Player[i]
			}
			playState.playerId = r.game.PlayerId
			playState.roomId = r.roomId
			//初始化
			playState.initWorld = operation.Blocks
			r.game.SetState(&PlayState{
				game: r.game,
			})
		case data.JOIN_ROOM:
			//玩家加入
			r.currentPlayerNumber++
		case data.LEAVE_ROOM:
			//玩家离开
			r.currentPlayerNumber--
		case data.JOIN_SUCCESS:
			r.roomId = operation.RoomId
			r.currentPlayerNumber = len(operation.Player)
		case data.JOIN_FAIL:
			//退出失败，暂时不管
			r.game.SetState(&MenuState{
				game: r.game,
			})
		}
	default:
		return nil
	}
	return nil
}

func (r *RoomState) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	x := ScreenWidth / 2
	y := int(float64(ScreenHeight) / 3)
	text := fmt.Sprintf("current user count: %d", r.currentPlayerNumber)
	data.DrawStaticText(text, data.NormalFace, x, y, color.White, screen, true)
	// Draw game buttons
	for _, button := range r.buttons {
		button.Draw(screen, &ebiten.DrawImageOptions{})
	}
	return
}
