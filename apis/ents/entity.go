package ents

import "github.com/Relixik/minecraft-server/apis/base"

type Entity interface {
	Sender
	base.Unique

	EntityUUID() int64
}
