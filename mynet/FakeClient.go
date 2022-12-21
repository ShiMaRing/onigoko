package mynet

import (
	"onigoko/data"
)

type FakeClient struct {
	receivedChan chan data.Operation
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		receivedChan: make(chan data.Operation, 10),
	}
}

func (f *FakeClient) SendMessage(operation data.Operation) {
	//接收到消息，直接向chan中塞
	switch operation.OperationType {
	//申请加入房间
	case data.JOIN_ROOM:
		f.receivedChan <- data.Operation{
			OperationType: data.JOIN_SUCCESS,
			Player:        make([]data.Player, 4),
			PlayerId:      1,
			RoomId:        1,
		}
	case data.HEART_BEAT:
		//不做处理
		return

	}

}

func (f *FakeClient) receiveMessage() error {
	return nil
}

func (f *FakeClient) Run() error {
	return nil
}

func (f *FakeClient) ReceivedMessage() chan data.Operation {
	return f.receivedChan
}
