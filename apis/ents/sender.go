package ents

import (
	"github.com/Relixik/minecraft-server/apis/base"
)

type Sender interface {
	base.Named

	SendMessage(message ...interface{})
}
