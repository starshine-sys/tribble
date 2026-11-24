// SPDX-License-Identifier: AGPL-3.0-only
package levels

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) leaderboard(ctx *bcr.Context) (err error) {
	sc, err := bot.getGuildConfig(ctx.Message.GuildID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	noRanks, err := bot.DB.GuildBoolGet(ctx.Message.GuildID, "levels:disable_ranks")
	if err != nil {
		return bot.Report(ctx, err)
	}
	if noRanks {
		return ctx.SendX("Ranks are disabled on this server.")
	}

	// full, _ := ctx.Flags.GetBool("full")
	lb, err := bot.getLeaderboard(ctx.Message.GuildID, true)
	if err != nil {
		return bot.Report(ctx, err)
	}

	if len(lb) == 0 {
		_, err = ctx.Sendf("There doesn't seem to be anyone on the leaderboard...")
		return
	}

	var strings []string
	for i, l := range lb {
		strings = append(strings, fmt.Sprintf(
			"%v. %v: `%v` XP, level `%v`\n",
			i+1,
			l.UserID.Mention(),
			humanize.Comma(l.XP),
			sc.CalculateLevel(l.XP),
		))
	}

	name := "Leaderboard for " + ctx.Guild.Name

	embeds := bcr.StringPaginator(name, bcr.ColourBlurple, strings, 20)

	_, err = bot.PagedEmbed(ctx,
		embeds,
		10*time.Minute,
	)
	return err
}
