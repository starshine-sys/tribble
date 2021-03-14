package info

import (
	"fmt"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) invite(ctx *bcr.Context) (err error) {
	perms := discord.PermissionViewChannel | discord.PermissionReadMessageHistory | discord.PermissionAddReactions | discord.PermissionAttachFiles | discord.PermissionBanMembers | discord.PermissionCreateInstantInvite | discord.PermissionUseExternalEmojis | discord.PermissionEmbedLinks | discord.PermissionKickMembers | discord.PermissionManageEmojis | discord.PermissionManageMessages | discord.PermissionManageRoles | discord.PermissionSendMessages

	if bot.Config.Branding.Private {
		if bot.Config.Branding.PublicID.IsValid() {
			u, err := bot.State.User(bot.Config.Branding.PublicID)
			if err == nil {
				_, err = ctx.Sendf("This instance of the bot is private, but you can invite %v, the public version of this bot, using the following link: <https://discord.com/api/oauth2/authorize?client_id=%v&permissions=%v&scope=applications.commands%%20bot>", u.Username, u.ID, perms)
				return err
			}
		}

		_, err = ctx.Send("This instance of the bot is private, please DM the bot's owner for details.", nil)
		return
	}

	invite := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%v&permissions=%v&scope=applications.commands%%20bot", ctx.Bot.ID, perms)

	_, err = ctx.Sendf("Use this link to invite me to your server: <%v>", invite)
	return
}
