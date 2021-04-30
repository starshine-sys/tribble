package moderation

import (
	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/starshine-sys/bcr"
)

func (bot *Bot) lockdown(ctx *bcr.Context) (err error) {
	ch := ctx.Channel
	if len(ctx.RawArgs) > 0 {
		ch, err = ctx.ParseChannel(ctx.RawArgs)
		if err != nil {
			_, err = ctx.Send("Could not find that channel!", nil)
			return
		}
	}
	if ch.GuildID != ctx.Message.GuildID {
		_, err = ctx.Send("That channel is not in this server.", nil)
		return
	}

	perms, err := bot.State.Permissions(ch.ID, ctx.Bot.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	// if we don't have perms, return
	if !perms.Has(discord.PermissionManageRoles) {
		_, err = ctx.Sendf("I do not have the **Manage Roles** permission in %v.", ch.Mention())
		return
	}

	err = bot.State.EditChannelPermission(ch.ID, discord.Snowflake(ctx.Bot.ID), api.EditChannelPermissionData{
		Type:  discord.OverwriteMember,
		Allow: discord.PermissionViewChannel,
	})
	if err != nil {
		_, err = ctx.Send("Could not change permissions for the given channel.", nil)
		return
	}

	var allow, deny discord.Permissions
	for _, p := range ch.Permissions {
		if p.ID == discord.Snowflake(ch.GuildID) {
			allow = p.Allow
			deny = p.Deny
		}
	}

	err = bot.State.EditChannelPermission(ch.ID, discord.Snowflake(ch.GuildID), api.EditChannelPermissionData{
		Type:  discord.OverwriteRole,
		Allow: allow,
		Deny:  deny ^ discord.PermissionViewChannel,
	})
	if err != nil {
		_, err = ctx.Send("Could not change permissions for the given channel.", nil)
		return
	}

	if (deny ^ discord.PermissionViewChannel).Has(discord.PermissionViewChannel) {
		_, err = ctx.Sendf("Locked down %v.", ch.Mention())
		return
	}
	_, err = ctx.Sendf("Unlocked %v.", ch.Mention())
	return
}
