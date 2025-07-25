package base

import "github.com/Relixik/minecraft-server/apis/uuid"

type Unique interface {
	UUID() uuid.UUID
}
