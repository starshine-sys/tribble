package mirror

import (
	"sync"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/tribble/bot"
	"github.com/starshine-sys/tribble/commands/moderation/modlog"
)

// Bot ...
type Bot struct {
	*bot.Bot
	ModLog *modlog.ModLog

	modIDMap   map[string]discord.UserID
	modIDMapMu sync.Mutex
}

// Init ...
func Init(bot *bot.Bot) (s string, list []*bcr.Command) {
	b := &Bot{
		Bot:      bot,
		ModLog:   modlog.New(bot),
		modIDMap: map[string]discord.UserID{},
	}

	b.Router.AddCommand(&bcr.Command{
		Name:    "modlog-import",
		Summary: "Bulk import mod logs from YAGPDB.xyz and Carl-bot.",
		Usage:   "<channel>",
		Args:    bcr.MinArgs(1),
		Hidden:  true,

		OwnerOnly:   true,
		Permissions: discord.PermissionManageGuild,
		Command:     b.cmdImport,
	})

	b.State.AddHandler(b.messageCreate)

	return
}