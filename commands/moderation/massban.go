// SPDX-License-Identifier: AGPL-3.0-only
package moderation

import (
	"fmt"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/tribble/commands/moderation/modlog"
)

func (bot *Bot) massban(ctx *bcr.Context) (err error) {
	// check bot perms
	if p, _ := ctx.State.Permissions(ctx.Channel.ID, ctx.Bot.ID); !p.Has(discord.PermissionBanMembers) {
		_, err = ctx.Send("I do not have the **Ban Members** permission.")
		return
	}

	reason := "N/A"
	users, n := ctx.GreedyUserParser(ctx.Args)
	if n == 0 {
		_, err = ctx.Sendf("Couldn't parse any users.")
		return
	}
	if n != -1 {
		reason = strings.Join(ctx.Args[n:], " ")
	}

	var toBan string
	for _, u := range users {
		toBan += fmt.Sprintf("%v (%v)\n", u.Tag(), u.ID)
	}

	yes, timeout := ctx.ConfirmButton(ctx.Author.ID, bcr.ConfirmData{
		Embeds: []discord.Embed{{
			Title:       "Confirmation",
			Description: "Are you sure you want to ban the following users?",
			Color:       ctx.Router.EmbedColor,
			Fields: []discord.EmbedField{
				{
					Name:  "Users",
					Value: toBan,
				},
				{
					Name:  "Reason",
					Value: reason,
				},
			},
			Timestamp: discord.NewTimestamp(time.Now().Add(5 * time.Minute)),
		}},
		YesPrompt: "Ban",
		YesStyle:  discord.DangerButtonStyle(),

		Timeout: 5 * time.Minute,
	})
	if timeout {
		_, err = ctx.Send("Timed out.")
		return
	}
	if !yes {
		_, err = ctx.Send("Massban cancelled.")
		return
	}

	for _, u := range users {
		err = ctx.State.Ban(ctx.Message.GuildID, u.ID, api.BanData{
			DeleteDays: option.NewUint(0),
			AuditLogReason: api.AuditLogReason(
				fmt.Sprintf("%v: %v", ctx.Author.Tag(), reason)),
		})
		if err != nil {
			_, err = ctx.Sendf("I could not ban **%v**, aborting.", u.Tag())
			return
		}

		bot.ModLog.Log(ctx.State, modlog.ActionBan, ctx.Message.GuildID, u.ID, ctx.Author.ID, reason)
	}

	_, err = ctx.Sendf("Banned %v members.", len(users))
	return
}
