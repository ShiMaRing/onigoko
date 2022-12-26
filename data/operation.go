package data

// Operation 前后端传输的消息体
type Operation struct {
	RoomId        int      `json:"roomId"`                  //房间号
	PlayerId      uint32   `json:"playerId,omitempty"`      //用户号
	OperationType int      `json:"operationType,omitempty"` //执行的操作类型
	Players       []Player `json:"players,omitempty"`       //包含每一个玩家的状态
	Blocks        []Block  `json:"blocks,omitempty"`        //更新指定的地块的状态
	Message       string   `json:"message,omitempty"`       //携带的消息
}
