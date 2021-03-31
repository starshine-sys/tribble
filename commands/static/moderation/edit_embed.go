package moderation

import (
	"encoding/json"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) editEmbed(ctx *bcr.Context) (err error) {
	msg, err := ctx.ParseMessage(ctx.Args[0])
	if err != nil {
		_, err = ctx.Send("I could not find that message.", nil)
		return
	}

	args := []byte(
		strings.TrimSpace(
			strings.TrimPrefix(ctx.RawArgs, ctx.Args[0]),
		),
	)

	var e discord.Embed
	err = json.Unmarshal(args, &e)
	if err != nil {
		ctx.Sendf("Error: %v", err)
		return
	}

	_, err = ctx.Edit(msg, "", &e)
	if err != nil {
		_, err = ctx.Sendf("Error: %v", err)
		return
	}

	_, err = ctx.Send("Message edited!", nil)
	return
}
