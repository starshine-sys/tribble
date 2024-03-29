// SPDX-License-Identifier: AGPL-3.0-only
package notes

import (
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/tribble/bot"
)

// Bot ...
type Bot struct {
	*bot.Bot
}

// Init ...
func Init(bot *bot.Bot) (s string, list []*bcr.Command) {
	s = "Notes"

	b := &Bot{bot}

	notes := b.Router.AddCommand(&bcr.Command{
		Name:    "notes",
		Summary: "List a user's notes.",
		Usage:   "<user>",
		Args:    bcr.MinArgs(1),

		Command: b.list,
	})

	notes.AddSubcommand(&bcr.Command{
		Name:    "import",
		Summary: "Import notes from JSON.",
		Command: b.importNotes,
	})

	notes.AddSubcommand(&bcr.Command{
		Name:    "export",
		Summary: "Export notes to JSON.",
		Command: b.exportNotes,
	})

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "setnote",
		Aliases: []string{"addnote"},
		Summary: "Add a note.",
		Usage:   "<user> <note>",
		Args:    bcr.MinArgs(2),

		Command: b.addNote,
	}))

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "delnote",
		Aliases: []string{"rmnote"},
		Summary: "Remove a note.",
		Usage:   "<note ID>",
		Args:    bcr.MinArgs(1),

		Command: b.delNote,
	}))

	b.Router.AddCommand(&bcr.Command{
		Name:    "bgc",
		Aliases: []string{"backgroundcheck"},
		Summary: "Show a background check for the given user.",
		Usage:   "[user]",

		Command: func(ctx *bcr.Context) (err error) {
			if len(ctx.Args) == 0 {
				ctx.Args = []string{ctx.Author.ID.String()}
				ctx.RawArgs = ctx.Author.ID.String()
			}

			err = bot.Router.GetCommand("i").Command(ctx)
			if err != nil {
				return
			}
			err = bot.Router.GetCommand("notes").Command(ctx)
			if err != nil {
				return
			}
			return bot.Router.GetCommand("modlogs").Command(ctx)
		},
	})

	return "Notes", append(list, notes)
}
