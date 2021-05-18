package moderation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) warn(ctx *bcr.Context) (err error) {
	go func() { ctx.State.Typing(ctx.Channel.ID) }()

	u, err := ctx.ParseMember(ctx.Args[0])
	if err != nil {
		_, err = ctx.Send("User not found.", nil)
		return
	}

	reason := strings.TrimSpace(strings.TrimPrefix(ctx.RawArgs, ctx.Args[0]))

	if u.User.ID == ctx.Bot.ID {
		_, err = ctx.Send("😭 Why would you do that?", nil)
		return
	}

	if !bot.aboveUser(ctx, u) {
		_, err = ctx.Send("You're not high enough in the hierarchy to do that.", nil)
		return
	}

	g, err := ctx.State.Guild(ctx.Message.GuildID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	err = bot.ModLog.Warn(ctx.Message.GuildID, u.User.ID, ctx.Author.ID, reason)
	if err != nil {
		return bot.Report(ctx, err)
	}

	_, err = ctx.NewDM(u.User.ID).Content(fmt.Sprintf("You were warned in %v.\nReason: %v", g.Name, reason)).Send()
	if err != nil {
		_, err = ctx.Send("The warning was logged, but I was unable to notify the user of their warning.", nil)
		return
	}

	_, err = ctx.NewMessage().Content(fmt.Sprintf("Warned **%v#%v**", u.User.Username, u.User.Discriminator)).Send()
	return
}

func (bot *Bot) aboveUser(ctx *bcr.Context, member *discord.Member) (above bool) {
	g, err := ctx.State.Guild(ctx.Message.GuildID)
	if err != nil {
		return false
	}

	var modRoles, memberRoles bcr.Roles
	for _, r := range g.Roles {
		for _, id := range ctx.Member.RoleIDs {
			if r.ID == id {
				modRoles = append(modRoles, r)
				break
			}
		}
		for _, id := range member.RoleIDs {
			if r.ID == id {
				memberRoles = append(memberRoles, r)
				break
			}
		}
	}

	if len(modRoles) == 0 {
		return false
	}
	if len(memberRoles) == 0 {
		return true
	}

	sort.Sort(modRoles)
	sort.Sort(memberRoles)

	return modRoles[0].Position > memberRoles[0].Position
}