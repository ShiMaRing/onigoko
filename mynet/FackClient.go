package mynet

import (
	"onigoko/data"
)

type FakeClient struct {
	receivedChan chan data.Operation
}

func NewFakeClient() *FakeClient {
	return &FakeClient{}
}

func (f *FakeClient) SendMessage(operation data.Operation) {
	//接收到消息，直接向chan中塞

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
