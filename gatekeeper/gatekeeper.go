// SPDX-License-Identifier: AGPL-3.0-only
package gatekeeper

import (
	"net/http"

	"github.com/kataras/hcaptcha"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/tribble/bot"

	_ "embed"
)

// Bot ...
type Bot struct {
	*bot.Bot
	HCaptcha *hcaptcha.Client
}

//go:embed gatekeeper.html
var gatekeeperHtml string

// Init ...
func Init(bot *bot.Bot) (s string, list []*bcr.Command) {
	s = "Verification"

	b := &Bot{
		Bot:      bot,
		HCaptcha: hcaptcha.New(bot.Config.HCaptchaSecret),
	}

	b.Chi.Get("/gatekeeper/{uuid}", b.GatekeeperGET)
	b.Chi.Get("/", func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, "https://github.com/starshine-sys/tribble", http.StatusTemporaryRedirect)
	})
	b.Chi.Post("/verify", b.VerifyPOST)

	list = append(list, b.Router.AddCommand(&bcr.Command{
		Name:    "agree",
		Summary: "Agree to the server's rules and get sent a captcha.",

		GuildOnly: true,
		Command:   b.agree,
	}))

	conf := b.Router.AddCommand(&bcr.Command{
		Name:    "captcha",
		Aliases: []string{"verify", "verification"},
		Summary: "View or change this server's captcha verification settings.",
		Command: b.settings,
	})

	conf.AddSubcommand(&bcr.Command{
		Name:    "channel",
		Summary: "Set the welcome channel",
		Usage:   "<new channel>",
		Args:    bcr.MinArgs(1),
		Command: b.setChannel,
	})

	conf.AddSubcommand(&bcr.Command{
		Name:    "log",
		Summary: "Set the gatekeeper log channel",
		Usage:   "<new channel>",
		Args:    bcr.MinArgs(1),
		Command: b.setLog,
	})

	conf.AddSubcommand(&bcr.Command{
		Name:    "message",
		Summary: "Set the welcome message",
		Usage:   "<new message>",
		Args:    bcr.MinArgs(1),
		Command: b.setMessage,
	})

	conf.AddSubcommand(&bcr.Command{
		Name:    "role",
		Summary: "Set the member role",
		Usage:   "<new role>",
		Args:    bcr.MinArgs(1),
		Command: b.setRole,
	})

	conf.AddSubcommand(&bcr.Command{
		Name:    "response",
		Aliases: []string{"resp"},
		Summary: "Set the response to the `agree` command. {mention} is replaced with the user's mention.",
		Usage:   "<new response>",
		Command: b.setResponse,
	})

	b.Router.AddHandler(b.memberLeave)

	return s, append(list, conf)
}
