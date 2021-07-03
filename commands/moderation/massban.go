package moderation

import (
	"fmt"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/starshine-sys/bcr"
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
		toBan += fmt.Sprintf("%v#%v (%v)\n", u.Username, u.Discriminator, u.ID)
	}

	msg, err := ctx.Send("", discord.Embed{
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
		Timestamp: discord.NewTimestamp(time.Now().Add(time.Minute)),
	})

	yes, timeout := ctx.YesNoHandler(*msg, ctx.Author.ID)
	if timeout {
		_, err = ctx.Send("Timed out.")
		return
	}
	if !yes {
		_, err = ctx.Send("Massban cancelled.")
	}

	ctx.State.DeleteMessage(msg.ChannelID, msg.ID)

	for _, u := range users {
		err = ctx.State.Ban(ctx.Message.GuildID, u.ID, api.BanData{
			DeleteDays: option.NewUint(0),
			Reason: option.NewString(
				fmt.Sprintf("%v#%v: %v", ctx.Author.Username, ctx.Author.Discriminator, reason)),
		})
		if err != nil {
			_, err = ctx.Sendf("I could not ban **%v#%v**, aborting.", u.Username, u.Discriminator)
			return
		}

		bot.ModLog.Ban(ctx, ctx.Message.GuildID, u.ID, ctx.Author.ID, reason)
	}

	_, err = ctx.Sendf("Banned %v members.", len(users))
	return
}
