package mode

import (
	"bytes"
	"fmt"

	"github.com/Relixik/minecraft-server/apis/data/chat"
	"github.com/Relixik/minecraft-server/apis/data/msgs"
	"github.com/Relixik/minecraft-server/apis/game"
	"github.com/Relixik/minecraft-server/apis/util"
	"github.com/Relixik/minecraft-server/apis/uuid"
	"github.com/Relixik/minecraft-server/impl/base"
	"github.com/Relixik/minecraft-server/impl/game/auth"
	"github.com/Relixik/minecraft-server/impl/game/ents"
	"github.com/Relixik/minecraft-server/impl/prot/client"
	"github.com/Relixik/minecraft-server/impl/prot/server"
)

/**
 * login
 */

func HandleState2(watcher util.Watcher, join chan base.PlayerAndConnection) {

	watcher.SubAs(func(packet *server.PacketILoginStart, conn base.Connection) {
		conn.CertifyValues(packet.PlayerName)

		_, public := auth.NewCrypt()

		// Check if encryption keys were generated successfully
		if public == nil {
			conn.SendPacket(&client.PacketODisconnect{
				Reason: *msgs.New("Server encryption error - please try again").SetColor(chat.Red),
			})
			return
		}

		response := client.PacketOEncryptionRequest{
			Server: "",
			Public: public,
			Verify: conn.CertifyData(),
		}

		conn.SendPacket(&response)
	})

	watcher.SubAs(func(packet *server.PacketIEncryptionResponse, conn base.Connection) {
		defer func() {
			if err := recover(); err != nil {
				conn.SendPacket(&client.PacketODisconnect{
					Reason: *msgs.New(fmt.Sprintf("Authentication failed: %v", err)).SetColor(chat.Red),
				})
			}
		}()

		ver, err := auth.Decrypt(packet.Verify)
		if err != nil {
			conn.SendPacket(&client.PacketODisconnect{
				Reason: *msgs.New(fmt.Sprintf("Failed to decrypt token: %v", err)).SetColor(chat.Red),
			})
			return
		}

		if !bytes.Equal(ver, conn.CertifyData()) {
			conn.SendPacket(&client.PacketODisconnect{
				Reason: *msgs.New("Encryption verification failed").SetColor(chat.Red),
			})
			return
		}

		sec, err := auth.Decrypt(packet.Secret)
		if err != nil {
			conn.SendPacket(&client.PacketODisconnect{
				Reason: *msgs.New(fmt.Sprintf("Failed to decrypt secret: %v", err)).SetColor(chat.Red),
			})
			return
		}

		conn.CertifyUpdate(sec) // enable encryption on the connection

		auth.RunAuthGet(sec, conn.CertifyName(), func(auth *auth.Auth, err error) {
			defer func() {
				if err := recover(); err != nil {
					conn.SendPacket(&client.PacketODisconnect{
						Reason: *msgs.New(fmt.Sprintf("Authentication failed: %v", err)).SetColor(chat.Red),
					})
				}
			}()

			if err != nil {
				conn.SendPacket(&client.PacketODisconnect{
					Reason: *msgs.New(fmt.Sprintf("Authentication failed: %v", err)).SetColor(chat.Red),
				})
				return
			}

			uuid, err := uuid.TextToUUID(auth.UUID)
			if err != nil {
				conn.SendPacket(&client.PacketODisconnect{
					Reason: *msgs.New(fmt.Sprintf("Invalid UUID format: %s", auth.UUID)).SetColor(chat.Red),
				})
				return
			}

			prof := game.Profile{
				UUID: uuid,
				Name: auth.Name,
			}

			for _, prop := range auth.Prop {
				prof.Properties = append(prof.Properties, &game.ProfileProperty{
					Name:      prop.Name,
					Value:     prop.Data,
					Signature: prop.Sign,
				})
			}

			player := ents.NewPlayer(&prof, conn)

			conn.SendPacket(&client.PacketOLoginSuccess{
				PlayerName: player.Name(),
				PlayerUUID: player.UUID().String(),
			})

			conn.SetState(base.PLAY)

			join <- base.PlayerAndConnection{
				Player:     player,
				Connection: conn,
			}
		})

	})

}
