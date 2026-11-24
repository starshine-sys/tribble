// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/starshine-sys/bcr"
	bcrbot "github.com/starshine-sys/bcr/bot"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/starshine-sys/tribble/bot"
	"github.com/starshine-sys/tribble/commands/admin"
	"github.com/starshine-sys/tribble/commands/config"
	"github.com/starshine-sys/tribble/commands/moderation"
	"github.com/starshine-sys/tribble/commands/moderation/notes"
	"github.com/starshine-sys/tribble/commands/reminders"
	"github.com/starshine-sys/tribble/commands/static"
	"github.com/starshine-sys/tribble/db"
	"github.com/starshine-sys/tribble/keyroles"
	"github.com/starshine-sys/tribble/levels"
	"github.com/starshine-sys/tribble/quotes"
	"github.com/starshine-sys/tribble/reactroles"
	"github.com/starshine-sys/tribble/starboard"
)

// const intents = bcr.RequiredIntents | gateway.IntentGuildMembers | gateway.IntentGuildVoiceStates | gateway.IntentGuildPresences
const intents = bcr.RequiredIntents | gateway.IntentGuildVoiceStates

// all of the bot's modules
var modules = []func(*bot.Bot) (string, []*bcr.Command){
	static.Init,
	moderation.Init,
	levels.Init,
	reactroles.Init,
	reminders.Init,
	config.Init,
	notes.Init,
	starboard.Init,
	admin.Init,
	keyroles.Init,
	quotes.Init,
}

func main() {
	// get the config
	c := getConfig()

	// set up a logger
	zcfg := zap.NewProductionConfig()
	zcfg.Encoding = "console"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zcfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zcfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	zcfg.Level.SetLevel(zapcore.DebugLevel)

	log, err := zcfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(err)
	}
	zap.RedirectStdLog(log)
	sugar := log.Sugar()

	ws.WSDebug = sugar.Debug
	ws.WSError = func(err error) {
		log.WithOptions(zap.AddCallerSkip(1)).Error("websocket error", zap.Error(err))
	}

	// open the database
	db, err := db.New(c.DatabaseURL, sugar, c)
	if err != nil {
		sugar.Fatalf("Error connecting to database: %v", err)
	}
	sugar.Info("Connected to database.")

	// create a new bot
	nsfn := state.NewShardFunc(func(_ *shard.Manager, s *state.State) {
		s.AddIntents(intents)
	})

	mgr, err := shard.NewIdentifiedManager(gateway.IdentifyCommand{
		Token: "Bot " + c.Token,
		Properties: gateway.IdentifyProperties{
			Browser: "Discord iOS",
		},
		LargeThreshold: 250,
	}, nsfn)
	if err != nil {
		sugar.Fatal("Error creating shard manager: %v", err)
	}

	owners := make([]string, 0)
	for _, o := range c.Owners {
		owners = append(owners, o.String())
	}

	r := bcr.New(mgr, owners, c.Prefixes)
	bcrbot := bcrbot.NewWithRouter(r)

	bot := bot.New(bcrbot, sugar, db, c)
	// set default embed colour
	bot.Router.EmbedColor = bcr.ColourBlurple

	// set the bot's prefix and owners
	bot.Prefix(c.Prefixes...)
	bot.Owner(c.Owners...)

	// add modules
	bot.Add(modules...)

	// set permission nodes
	bot.InitValidPermissionNodes()

	s, _ := bot.Router.StateFromGuildID(0)

	botUser, err := s.Me()
	if err != nil {
		sugar.Fatal("error getting bot user:", err)
	}

	sugar.Infof("User: %v (%v)", botUser.Tag(), botUser.ID)
	bot.Router.Bot = botUser

	// connect to discord
	if err := bot.Start(context.Background()); err != nil {
		sugar.Fatal("Failed to connect:", err)
	}

	// Defer this to make sure that things are always cleanly shutdown even in the event of a crash
	defer func() {
		db.Pool.Close()
		sugar.Info("Closed database connection.")
		bot.Router.ShardManager.Close()
		sugar.Info("Disconnected from Discord.")
	}()

	sugar.Info("Connected to Discord. Press Ctrl-C or send an interrupt signal to stop.")

	// start event scheduler
	go bot.Scheduler.Start()

	// sync slash commands
	if !c.NoSyncCommands {
		if len(c.SyncCommandsIn) > 0 {
			sugar.Infof("Overwriting guild commands in: %v", c.SyncCommandsIn)
			for _, g := range c.SyncCommandsIn {
				_, err := bot.Interactions.Rest.BulkOverwriteGuildCommands(discord.AppID(botUser.ID), g, bot.Commands())
				if err != nil {
					sugar.Errorf("Couldn't overwrite guild slash commands in %v: %v", g, err)
				} else {
					sugar.Infof("Overwrote guild commands in %v!", g)
				}
			}
		} else {
			sugar.Infof("Overwriting global commands")
			_, err := bot.Interactions.Rest.BulkOverwriteCommands(discord.AppID(botUser.ID), bot.Commands())
			if err != nil {
				sugar.Errorf("Couldn't overwrite global slash commands: %v", err)
			} else {
				sugar.Infof("Overwrote global commands!")
			}
		}
	} else {
		sugar.Info("Not syncing commands.")
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	sugar.Infof("Interrupt signal received. Shutting down...")
}
