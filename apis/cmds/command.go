package cmds

import (
	"github.com/Relixik/minecraft-server/apis/base"
	"github.com/Relixik/minecraft-server/apis/ents"
)

type Command interface {
	base.Named
	base.State

	Evaluate(sender ents.Sender, params []string)

	Complete(sender ents.Sender, params []string, output *[]string)
}
