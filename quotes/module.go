// SPDX-License-Identifier: AGPL-3.0-only
package quotes

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/spf13/pflag"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/tribble/bot"
)

// Bot ...
type Bot struct {
	*bot.Bot

	AESKey [32]byte

	mu   map[discord.MessageID]*sync.Mutex
	muMu sync.Mutex
}

// Init ...
func Init(bot *bot.Bot) (s string, list []*bcr.Command) {
	s = "Quotes"

	b := &Bot{
		Bot: bot,
		mu:  make(map[discord.MessageID]*sync.Mutex),
	}

	b.Router.AddHandler(b.reactionAdd)

	cmd := b.Router.AddCommand(&bcr.Command{
		Name:    "quote",
		Summary: "Show a random quote, or a quote from a given user, or a quote by ID.",
		Usage:   "[quote ID|user ID]",

		GuildOnly: true,
		Command:   b.quote,
	})

	cmd.AddSubcommand(&bcr.Command{
		Name:    "delete",
		Aliases: []string{"yeet"},
		Summary: "Delete the quote with the given ID.",
		Usage:   "<quote ID>",
		Args:    bcr.MinArgs(1),

		Command: b.cmdQuoteDelete,
	})

	cmd.AddSubcommand(&bcr.Command{
		Name:    "leaderboard",
		Aliases: []string{"lb"},
		Summary: "Show this server's quote leaderboard, for users or channels.",
		Usage:   "<\"channel\" or \"user\">",
		Args:    bcr.MinArgs(1),

		GuildOnly: true,
		Command:   b.leaderboard,
	})

	cmd.AddSubcommand(&bcr.Command{
		Name:    "block",
		Aliases: []string{"unblock", "enable", "disable", "no"},
		Summary: "Disable quotes for yourself (block your messages from being quoted).",

		Command: b.disable,
	})

	quotes := b.Router.AddCommand(&bcr.Command{
		Name:    "quotes",
		Summary: "Show a list of quotes.",

		Flags: func(fs *pflag.FlagSet) *pflag.FlagSet {
			fs.SortFlags = false

			fs.StringP("user", "u", "", "Filter by a user.")
			fs.StringP("channel", "c", "", "Filter by a channel.")

			fs.BoolP("server", "s", false, "Limit search to the current server.")

			fs.BoolP("sort-by-message", "m", false, "Sort by message ID.")
			fs.BoolP("reversed", "r", false, "Reverse sorting.")

			return fs
		},

		GuildOnly: true,
		Command:   b.list,
	})

	cmd.AddSubcommand(b.Router.AliasMust("list", nil, []string{"quotes"}, nil))
	quotes.AddSubcommand(b.Router.AliasMust("leaderboard", []string{"lb"}, []string{"quote", "leaderboard"}, nil))

	quotes.AddSubcommand(&bcr.Command{
		Name:    "toggle",
		Summary: "Enable or disable quotes for this server.",
		Usage:   "<on|off>",
		Args:    bcr.MinArgs(1),

		Command: b.toggle,
	})

	quotes.AddSubcommand(&bcr.Command{
		Name:    "messages",
		Summary: "Enable or disable quote messages for this server.",
		Usage:   "<on|off>",
		Args:    bcr.MinArgs(1),

		Command: b.toggleSuppressMessages,
	})

	return s, append(list, cmd, quotes)
}
