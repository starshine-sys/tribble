package gatekeeper

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/kataras/hcaptcha"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/tribble/bot"
)

// Bot ...
type Bot struct {
	*bot.Bot

	httpRouter *httprouter.Router

	pending map[string]PendingUser

	HCaptcha *hcaptcha.Client
}

// Init ...
func Init(bot *bot.Bot) (s string, list []*bcr.Command) {
	s = "Verification"

	b := &Bot{
		Bot:        bot,
		httpRouter: httprouter.New(),
		HCaptcha:   hcaptcha.New(bot.Config.HCaptchaSecret),
	}

	b.httpRouter.ServeFiles("/style/*filepath", http.Dir("../../gatekeeper/style"))
	b.httpRouter.GET("/gatekeeper/:uuid", b.GatekeeperGET)
	b.httpRouter.POST("/verify", b.VerifyPOST)

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

		CustomPermissions: bot.ModRole,
		Command:           b.settings,
	})

	conf.AddSubcommand(&bcr.Command{
		Name:    "channel",
		Summary: "Set the welcome channel",
		Usage:   "<new channel>",
		Args:    bcr.MinArgs(1),

		CustomPermissions: bot.ModRole,
		Command:           b.setChannel,
	})

	conf.AddSubcommand(&bcr.Command{
		Name:    "message",
		Summary: "Set the welcome message",
		Usage:   "<new message>",
		Args:    bcr.MinArgs(1),

		CustomPermissions: bot.ModRole,
		Command:           b.setMessage,
	})

	conf.AddSubcommand(&bcr.Command{
		Name:    "role",
		Summary: "Set the member role",
		Usage:   "<new role>",
		Args:    bcr.MinArgs(1),

		CustomPermissions: bot.ModRole,
		Command:           b.setRole,
	})

	go func() {
		for {
			err := http.ListenAndServe(bot.Config.VerifyListen, b.httpRouter)
			if err != nil {
				bot.Sugar.Errorf("Error running HTTP server, restarting: %v", err)
			}
			time.Sleep(30 * time.Second)
		}
	}()

	b.Router.AddHandler(b.memberAdd)
	b.Router.AddHandler(b.memberLeave)

	return s, append(list, conf)
}
