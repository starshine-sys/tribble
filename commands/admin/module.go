// SPDX-License-Identifier: AGPL-3.0-only
package admin

import (
	"context"
	"net/http"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/voice"
	"github.com/spf13/pflag"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/tribble/bot"
	"gitlab.com/1f320/x/concurrent"
)

// Bot ...
type Bot struct {
	*bot.Bot

	loadedAllowedGuilds bool

	Client        *http.Client
	VoiceSessions *concurrent.Map[discord.GuildID, *voice.Session]
}

// Init ...
func Init(bot *bot.Bot) (s string, list []*bcr.Command) {
	s = "Bot owner commands"

	b := &Bot{
		Bot:           bot,
		Client:        &http.Client{},
		VoiceSessions: concurrent.NewMap[discord.GuildID, *voice.Session](),
	}

	if b.Config.Branding.Private {
		var guildCount int
		err := b.DB.Pool.QueryRow(context.Background(), "select count(*) from allowed_guilds").Scan(&guildCount)
		if err != nil {
			b.Sugar.Fatalf("Error getting allowed guild count: %v", err)
		}

		b.loadedAllowedGuilds = guildCount != 0

		if !b.loadedAllowedGuilds {
			b.Sugar.Info("No allowed guilds found, will *not* leave any guilds this session.")
		}
	}

	allowList := b.Router.AddCommand(&bcr.Command{
		Name:      "allowlist",
		Summary:   "Show a list of guilds the bot is allowed to join.",
		Hidden:    true,
		OwnerOnly: true,
		Command:   b.listAllowedGuilds,
	})

	allowList.AddSubcommand(&bcr.Command{
		Name:      "add",
		Summary:   "Add a guild to the allowlist",
		Hidden:    true,
		OwnerOnly: true,
		Command:   b.addAllowedGuild,
	})

	list = append(list, allowList)

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "guild",
		Summary: "Show info for the given server ID.",
		Usage:   "<ID>",
		Args:    bcr.MinArgs(1),

		Hidden:    true,
		OwnerOnly: true,
		Command:   b.serverInfo,
	}))

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "status",
		Summary: "Set the bot's status.",
		Usage:   "[new status]",

		Hidden:    true,
		OwnerOnly: true,
		Command:   b.status,
	}))

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "activity",
		Summary: "Set the bot's activity.",
		Usage:   "[type] [new activity]",

		Hidden:    true,
		OwnerOnly: true,
		Command:   b.activity,
	}))

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "dm",
		Aliases: []string{"dmuser", "admindm"},
		Summary: "DM the given user a text-only message.",
		Usage:   "<user> <message>",

		Hidden:    true,
		OwnerOnly: true,
		Command:   b.dm,
	}))

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "rest",
		Summary: "Call a REST API and print the returned JSON.",
		Usage:   "<url>",

		Hidden:    true,
		OwnerOnly: true,
		Command:   b.rest,
	}))

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "lurk",
		Summary: "Lurk in a voice channel.",
		Usage:   "<channel>",
		Args:    bcr.MinArgs(1),
		Flags: func(fs *pflag.FlagSet) *pflag.FlagSet {
			fs.BoolP("unmute", "m", false, "Join unmuted")
			fs.BoolP("undeafen", "d", false, "Join undeafened")
			return fs
		},

		Hidden:    true,
		OwnerOnly: true,
		Command:   b.lurk,
	}))

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "unlurk",
		Summary: "Stop lurking in a voice channel.",
		Usage:   "[guild ID]",

		Hidden:    true,
		OwnerOnly: true,
		Command:   b.unlurk,
	}))

	b.Router.ShardManager.ForEach(func(s shard.Shard) {
		state := s.(*state.State)

		state.AddHandler(func(_ *gateway.ReadyEvent) {
			b.updateStatus(state)
		})

		if b.Config.Branding.Private {
			state.AddHandler(b.guildCreate)
		}
	})

	return
}
