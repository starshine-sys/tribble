package moderation

import (
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/tribble/bot"
)

// Bot ...
type Bot struct {
	*bot.Bot
}

// Init ...
func Init(bot *bot.Bot) (s string, list []*bcr.Command) {
	s = "Moderation commands"

	b := &Bot{Bot: bot}

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "members",
		Summary: "Show a filtered list of members.",
		Usage:   "--help",

		Permissions: discord.PermissionManageMessages,
		Command:     b.members,
	}))

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "roledump",
		Aliases: []string{"role-dump"},
		Summary: "Show a list of *all* roles with permissions and basic information.",

		Permissions: discord.PermissionManageRoles,
		Command:     b.roleDump,
	}))

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "massban",
		Summary: "Ban all the given members with an optional reason.",
		Usage:   "<users...> [reason]",
		Args:    bcr.MinArgs(1),

		Permissions: discord.PermissionBanMembers,
		Command:     b.massban,
	}))

	echo := b.Router.AddCommand(&bcr.Command{
		Name:        "echo",
		Aliases:     []string{"say"},
		Summary:     "Make the bot say something.",
		Description: "To echo something into a different channel, use the `echo to` subcommand.",

		Permissions: discord.PermissionManageMessages,

		Command: b.echo,
	})

	echo.AddSubcommand(&bcr.Command{
		Name:    "to",
		Summary: "Echo something to the specified channel.",
		Usage:   "<channel>",

		Permissions: discord.PermissionManageMessages,

		Command: b.echoTo,
	})

	embed := b.Router.AddCommand(&bcr.Command{
		Name:    "embed",
		Summary: "Send an embed to this channel.",
		Usage:   "<JSON>",
		Args:    bcr.MinArgs(1),

		Permissions: discord.PermissionManageMessages,

		Command: b.embed,
	})

	embed.AddSubcommand(&bcr.Command{
		Name:    "to",
		Summary: "Send an embed to the given channel.",
		Usage:   "<channel> <JSON>",
		Args:    bcr.MinArgs(2),

		Permissions: discord.PermissionManageMessages,

		Command: b.embedTo,
	})

	embed.AddSubcommand(&bcr.Command{
		Name:    "edit",
		Summary: "Edit the given message.",
		Usage:   "<message> <JSON>",
		Args:    bcr.MinArgs(2),

		Permissions: discord.PermissionManageMessages,

		Command: b.editEmbed,
	})

	list = append(list, echo, embed)

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "makeinvite",
		Aliases: []string{"createinvite"},
		Summary: "Make an invite for the current channel, or the given channel.",
		Usage:   "[channel]",
		Args:    bcr.MinArgs(1),

		Permissions: discord.PermissionCreateInstantInvite,
		Command:     b.makeInvite,
	}))

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "lockdown",
		Aliases: []string{"lock"},
		Summary: "Toggles a channel being locked, hiding it from the `@everyone` role.",
		Usage:   "[channel]",

		Permissions: discord.PermissionManageRoles,
		Command:     b.lockdown,
	}))

	return
}
