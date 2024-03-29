// SPDX-License-Identifier: AGPL-3.0-only
package levels

import (
	"context"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) setXP(ctx *bcr.Context) (err error) {
	u, err := ctx.ParseMember(ctx.Args[0])
	if err != nil {
		_, err = ctx.Send("Couldn't find a member with that name.")
		return
	}

	xp, err := strconv.ParseInt(ctx.Args[1], 0, 0)
	if err != nil || xp < 0 || xp > 100000000 {
		_, err = ctx.Send("Couldn't parse your input as a number.")
		return
	}

	_, err = bot.DB.Pool.Exec(context.Background(), `insert into levels
	(server_id, user_id, xp) values ($1, $2, $3)
	on conflict (server_id, user_id) do update
	set xp = $3`, ctx.Message.GuildID, u.User.ID, xp)
	if err != nil {
		return bot.Report(ctx, err)
	}

	_, err = ctx.Reply("Updated %v's XP to `%v`.", u.Mention(), humanize.Comma(xp))
	return
}
