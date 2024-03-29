// SPDX-License-Identifier: AGPL-3.0-only
package info

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/tribble/common"
)

func (bot *Bot) idtime(ctx *bcr.Context) (err error) {
	sf, err := discord.ParseSnowflake(ctx.Args[0])
	if err != nil {
		_, err = ctx.Sendf("Couldn't parse ``%v`` as a Discord ID.", bcr.EscapeBackticks(ctx.Args[0]))
		return err
	}

	_, err = ctx.Send("", discord.Embed{
		Title:       fmt.Sprintf("Timestamp for `%v`", sf),
		Description: fmt.Sprintf("<t:%v>\n%v", sf.Time().Unix(), common.FormatTime(sf.Time().UTC())),
		Color:       ctx.Router.EmbedColor,
	})
	return
}
