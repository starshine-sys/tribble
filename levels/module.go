// SPDX-License-Identifier: AGPL-3.0-only
package levels

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/spf13/pflag"
	"github.com/starshine-sys/bcr"
	bcr2 "github.com/starshine-sys/bcr/v2"
	"github.com/starshine-sys/tribble/bot"
)

// Bot ...
type Bot struct {
	*bot.Bot
}

// Init ...
func Init(bot *bot.Bot) (s string, list []*bcr.Command) {
	s = "Levels"

	b := &Bot{bot}

	bot.Router.AddHandler(b.messageCreate)

	bot.Interactions.Command("level/show").Check(func(ctx *bcr2.CommandContext) (err error) {
		if ctx.Guild == nil {
			return bcr2.NewCheckError[*bcr2.CommandContext]("This command cannot be run in DMs.")
		}
		return nil
	}).Exec(b.showLevel)

	bot.Interactions.Command("level/leaderboard").Check(func(ctx *bcr2.CommandContext) (err error) {
		if ctx.Guild == nil {
			return bcr2.NewCheckError[*bcr2.CommandContext]("This command cannot be run in DMs.")
		}
		return nil
	}).Exec(b.leaderboardSlash)

	lvl := bot.Router.AddCommand(&bcr.Command{
		Name:    "level",
		Aliases: []string{"lvl", "rank"},
		Summary: "Show your or another user's level.",
		Usage:   "[user]",

		GuildOnly: true,
		Command:   b.level,
	})

	lvl.AddSubcommand(&bcr.Command{
		Name:    "setxp",
		Aliases: []string{"setexp", "set-xp", "set-exp"},
		Summary: "Set the given user's XP.",
		Usage:   "<user> <new XP>",
		Args:    bcr.MinArgs(2),

		Command: b.setXP,
	})

	lvl.AddSubcommand(&bcr.Command{
		Name:    "setlvl",
		Aliases: []string{"setlevel", "set-lvl", "set-level"},
		Summary: "Set the given user's XP to the minimum needed for the given level.",
		Usage:   "<user> <new level>",
		Args:    bcr.MinArgs(2),

		Command: b.setlvl,
	})

	lvl.AddSubcommand(&bcr.Command{
		Name:    "colour",
		Aliases: []string{"color"},
		Summary: "Set the colour used in your level card.",
		Usage:   "[new colour|clear]",

		GuildOnly: true,
		Command:   b.colour,
	})

	bg := lvl.AddSubcommand(&bcr.Command{
		Name:        "background",
		Aliases:     []string{"bg"},
		Summary:     "Set the background used in your level card to the attached image.",
		Description: "The image will automatically be resized to be 1200 pixels wide, with the original aspect ratio preserved.\nOnly the top 400 pixels (when resized) will be shown as part of the background.",
		Usage:       "[clear]",

		GuildOnly: true,
		Command:   b.background,
	})

	bg.AddSubcommand(&bcr.Command{
		Name:    "server",
		Summary: "Set this server's default level background.",
		Usage:   "[clear]",

		Command: b.serverBackground,
	})

	cfg := lvl.AddSubcommand(&bcr.Command{
		Name:    "config",
		Aliases: []string{"cfg"},
		Summary: "Configure levels.",
		Usage:   "[key <new value>]",

		Command: b.config,
	})

	cfg.AddSubcommand(&bcr.Command{
		Name:    "import",
		Summary: "Import levels",
		Usage:   "[JSON url]",
		Command: b.importLevels,
	})

	cfg.AddSubcommand(&bcr.Command{
		Name:    "export",
		Summary: "Export levels",
		Command: b.exportLevels,
	})

	cfg.AddSubcommand(&bcr.Command{
		Name:    "add-reward",
		Aliases: []string{"addreward"},
		Summary: "Add a level reward.",
		Usage:   "<lvl> <role>",
		Args:    bcr.MinArgs(2),

		Permissions: discord.PermissionManageRoles,
		Command:     b.cmdAddReward,
	})

	cfg.AddSubcommand(&bcr.Command{
		Name:    "del-reward",
		Aliases: []string{"delreward"},
		Summary: "Remove a level reward.",
		Usage:   "<lvl>",
		Args:    bcr.MinArgs(1),

		Permissions: discord.PermissionManageRoles,
		Command:     b.cmdDelReward,
	})

	bl := cfg.AddSubcommand(&bcr.Command{
		Name:    "blacklist",
		Aliases: []string{"bl"},
		Summary: "Configure this server's blacklist.",

		Command: func(ctx *bcr.Context) error { return ctx.Help([]string{"lvl", "config", "blacklist"}) },
	})

	bl.AddSubcommand(&bcr.Command{
		Name:    "add",
		Aliases: []string{"+", "block"},
		Summary: "Add a category, channel, or role to the blacklist.",
		Usage:   "<channel|role>",
		Args:    bcr.MinArgs(1),

		Command: b.blacklistAdd,
	})

	bl.AddSubcommand(&bcr.Command{
		Name:    "remove",
		Aliases: []string{"rm", "-", "unblock"},
		Summary: "Remove a category, channel, or role from the blacklist.",
		Usage:   "<channel|role>",
		Args:    bcr.MinArgs(1),

		Command: b.blacklistRemove,
	})

	list = append(list, bot.Router.AddCommand(&bcr.Command{
		Name:    "leaderboard",
		Aliases: []string{"lb"},
		Summary: "Show this server's leaderboard.",

		Flags: func(fs *pflag.FlagSet) *pflag.FlagSet {
			fs.BoolP("full", "f", false, "Show the full leaderboard, including people who left the server.")

			return fs
		},

		GuildOnly: true,
		Command:   b.leaderboard,
	}))

	nolevels := bot.Router.AddCommand(&bcr.Command{
		Name:    "nolevels",
		Aliases: []string{"nolevel"},
		Summary: "Manage the user blacklist for levels.",
		Usage:   "[user [time]]",

		Command: b.nolevelsList,
	})

	nolevels.AddSubcommand(&bcr.Command{
		Name:    "add",
		Summary: "Blacklist a user.",
		Usage:   "<user> [time]",
		Args:    bcr.MinArgs(1),

		Command: b.nolevelsAdd,
	})

	nolevels.AddSubcommand(&bcr.Command{
		Name:    "remove",
		Summary: "Unblacklist a user.",
		Usage:   "<user>",
		Args:    bcr.MinArgs(1),

		Command: b.nolevelsRemove,
	})

	state, _ := bot.Router.StateFromGuildID(0)

	var o sync.Once
	state.AddHandler(func(_ *gateway.ReadyEvent) {
		o.Do(func() {
			go b.nolevelLoop()
			go b.voiceLevelsLoop()
		})
	})

	return s, append(list, lvl, bl, nolevels)
}
