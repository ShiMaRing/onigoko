package model

// Player 玩家结构体，与后端对应
type Player struct {
	Id          int  `json:"id,omitempty"`       //玩家id，唯一标识玩家，room-id + player-id
	Identity    int  `json:"identity,omitempty"` //当前玩家是人还是鬼
	X           int  `json:"x,omitempty"`
	Y           int  `json:"y,omitempty"`           //当前人的位置或者鬼的位置，在某一具体格子上
	Mines       int  `json:"mines,omitempty"`       //人当前的地雷数量
	Lights      int  `json:"lights,omitempty"`      //当前的照明次数
	IsLighting  bool `json:"isLighting,omitempty"`  //当前是否在照明
	IsEscaped   bool `json:"isEscaped,omitempty"`   //是否成功逃跑
	IsDead      bool `json:"isDead,omitempty"`      //当前玩家是否死亡
	IsDizziness bool `json:"isDizziness,omitempty"` //是否被晕眩
}

// Block 表示每一个块数据
type Block struct {
	BlockType int `json:"blockType,omitempty"` //方块种类
	PlayerId  int `json:"playerId,omitempty"`  //若玩家存在，则需要表示玩家的id
	X         int `json:"x"`
	Y         int `json:"y"`
}
