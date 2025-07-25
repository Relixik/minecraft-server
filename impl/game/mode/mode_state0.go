package mode

import (
	"github.com/Relixik/minecraft-server/apis/util"
	"github.com/Relixik/minecraft-server/impl/base"
	"github.com/Relixik/minecraft-server/impl/prot/server"
)

/**
 * handshake
 */

func HandleState0(watcher util.Watcher) {

	watcher.SubAs(func(packet *server.PacketIHandshake, conn base.Connection) {
		conn.SetState(packet.State)
	})

}
