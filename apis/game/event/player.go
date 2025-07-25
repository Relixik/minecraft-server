package event

import "github.com/Relixik/minecraft-server/apis/ents"

type PlayerEvent struct {
	ents.Player
}

type PlayerJoinEvent struct {
	PlayerEvent
}

type PlayerQuitEvent struct {
	PlayerEvent
}
