package client

import (
	"encoding/json"
	"log"

	"github.com/Relixik/minecraft-server/apis/buff"
	"github.com/Relixik/minecraft-server/impl/base"
	"github.com/Relixik/minecraft-server/impl/data/status"
)

// done

type PacketOResponse struct {
	Status status.Response
}

func (p *PacketOResponse) UUID() int32 {
	return 0x00
}

func (p *PacketOResponse) Push(writer buff.Buffer, conn base.Connection) {
	if text, err := json.Marshal(p.Status); err != nil {
		log.Printf("WARNING: Failed to marshal server status: %v - sending empty response", err)
		writer.PushTxt("{}")
	} else {
		writer.PushTxt(string(text))
	}
}

type PacketOPong struct {
	Ping int64
}

func (p *PacketOPong) UUID() int32 {
	return 0x01
}

func (p *PacketOPong) Push(writer buff.Buffer, conn base.Connection) {
	writer.PushI64(p.Ping)
}
