// SPDX-License-Identifier: AGPL-3.0-only
package levels

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/dustin/go-humanize"
	bcr1 "github.com/starshine-sys/bcr"
	"github.com/starshine-sys/bcr/v2"
)

func (bot *Bot) showLevel(ctx *bcr.CommandContext) (err error) {
	bot.Sugar.Infof("interaction id: %v", ctx.InteractionID)

	sc, err := bot.getGuildConfig(ctx.Guild.ID)
	if err != nil {
		return bot.ReportInteraction(ctx, err)
	}
	if !sc.LevelsEnabled {
		return ctx.ReplyEphemeral("Levels are not enabled on this server.")
	}

	u := &ctx.User
	sf, _ := ctx.Option("user").SnowflakeValue()
	if sf.IsValid() {
		u, err = ctx.State.User(discord.UserID(sf))
		if err != nil {
			return ctx.ReplyEphemeral("User not found.")
		}
	}

	uc, err := bot.getUser(ctx.Guild.ID, u.ID)
	if err != nil {
		return bot.ReportInteraction(ctx, err)
	}

	ctx.Defer()

	lvl := sc.CalculateLevel(uc.XP)
	xpForNext := sc.CalculateExp(lvl + 1)
	xpForPrev := sc.CalculateExp(lvl)

	// get leaderboard (for rank)
	// filter the leaderboard to match the `leaderboard` command
	var rank int
	noRanks, err := bot.DB.GuildBoolGet(ctx.Guild.ID, "levels:disable_ranks")
	if err != nil {
		return bot.ReportInteraction(ctx, err)
	}
	if !noRanks {
		lb, err := bot.getLeaderboard(ctx.Guild.ID, false)
		if err == nil {
			for i, uc := range lb {
				if uc.UserID == u.ID {
					rank = i + 1
					break
				}
			}
		}
	}

	// get user colour + avatar URL
	clr := uc.Colour
	avatarURL := u.AvatarURLWithType(discord.PNGImage) + "?size=256"
	username := u.Username
	if ctx.Guild != nil {
		m, err := bot.Member(ctx.Guild.ID, u.ID)
		if err == nil {
			if clr == 0 {
				clr, _ = discord.MemberColor(*ctx.Guild, m)
			}
			if m.Avatar != "" {
				avatarURL = m.AvatarURLWithType(discord.PNGImage, ctx.Guild.ID) + "?size=256"
			}
			if m.Nick != "" {
				username = m.Nick
			}
		}
	}

	// background image
	background := ""
	if sc.Background != "" {
		background = sc.Background
	}
	if uc.Background != "" {
		background = uc.Background
	}

	r, err := bot.generateImage(username, avatarURL, background, clr, rank, lvl, uc.XP, xpForNext, xpForPrev)
	if err != nil {
		bot.Sugar.Errorf("Error generating level card: %v", err)
		return ctx.Reply("", bot.generateEmbed(username, avatarURL, clr, rank, lvl, uc.XP, xpForNext, xpForPrev, sc))
	}

	err = ctx.ReplyComplex(api.InteractionResponseData{
		Content: option.NewNullableString(""),
		Files: []sendpart.File{{
			Name:   "level_card.png",
			Reader: r,
		}},
	})
	if err != nil {
		bot.Sugar.Errorf("interaction id: %v / error sending level card: %v", ctx.InteractionID, err)
	}
	return err
}

func (bot *Bot) leaderboardSlash(ctx *bcr.CommandContext) (err error) {
	sc, err := bot.getGuildConfig(ctx.Guild.ID)
	if err != nil {
		return bot.ReportInteraction(ctx, err)
	}

	noRanks, err := bot.DB.GuildBoolGet(ctx.Guild.ID, "levels:disable_ranks")
	if err != nil {
		return bot.ReportInteraction(ctx, err)
	}
	if noRanks {
		return ctx.ReplyEphemeral("Ranks are disabled on this server.")
	}

	if sc.LeaderboardModOnly || !sc.LevelsEnabled {
		if err = bot.CheckHelper(ctx); err != nil {
			if _, ok := err.(bcr.CheckError[*bcr.CommandContext]); !ok {
				bot.Sugar.Errorf("error checking for helper perms: %v", err)
			}

			return ctx.ReplyEphemeral("You don't have permission to use this command, you need the **Manage Messages** permission to use it.")
		}
	}

	full, _ := ctx.Option("full").BoolValue()
	lb, err := bot.getLeaderboard(ctx.Guild.ID, full)
	if err != nil {
		return bot.ReportInteraction(ctx, err)
	}

	if len(lb) == 0 {
		return ctx.ReplyEphemeral("There doesn't seem to be anyone on the leaderboard...")
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

	embeds := bcr1.StringPaginator(name, bcr1.ColourBlurple, strings, 20)

	_, _, err = ctx.Paginate(bcr.PaginateEmbeds(embeds...), 15*time.Minute)
	return err
}
