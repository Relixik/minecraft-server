package base

import "github.com/Relixik/minecraft-server/apis/ents"

type PlayerAndConnection struct {
	Connection
	ents.Player
}
