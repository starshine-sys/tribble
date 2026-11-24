// SPDX-License-Identifier: AGPL-3.0-only
package bot

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/handler"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/getsentry/sentry-go"
	"github.com/starshine-sys/bcr"
	bcrbot "github.com/starshine-sys/bcr/bot"
	bcr2 "github.com/starshine-sys/bcr/v2"
	"github.com/starshine-sys/pkgo/v2"
	"github.com/starshine-sys/tribble/common"
	"github.com/starshine-sys/tribble/db"
	"go.uber.org/zap"
)

// Bot is the main bot struct
type Bot struct {
	*bcrbot.Bot

	Interactions *bcr2.Router

	Config      *common.BotConfig
	Sugar       *zap.SugaredLogger
	DB          *db.DB
	Scheduler   *Scheduler
	EmbedColour discord.Color

	GuildLogWebhook *webhook.Client

	Counters struct {
		Mu sync.Mutex

		Mentions int
		Messages int
	}

	Hub *sentry.Hub

	PK *pkgo.Session

	members   map[memberKey]member
	membersMu sync.RWMutex

	ValidNodes []string
}

type member struct {
	discord.Member
	GuildID discord.GuildID
}

type memberKey struct {
	GuildID discord.GuildID
	UserID  discord.UserID
}

// Module is a single module/category of commands
type Module interface {
	String() string
	Commands() []*bcr.Command
}

// New creates a new instance of Bot
func New(
	bot *bcrbot.Bot,
	sugar *zap.SugaredLogger,
	db *db.DB,
	config *common.BotConfig) *Bot {
	b := &Bot{
		Bot:         bot,
		Sugar:       sugar,
		DB:          db,
		Config:      config,
		PK:          pkgo.New(""),
		EmbedColour: bcr.ColourBlurple,

		members: map[memberKey]member{},
	}

	b.Scheduler = NewScheduler(b, db)

	b.Interactions = bcr2.NewFromShardManager("Bot "+config.Token, bot.Router.ShardManager)

	// create a Sentry config
	if config.SentryURL != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: config.SentryURL,
		})
		if err != nil {
			sugar.Fatalf("sentry.Init: %s", err)
		}
		sugar.Infof("Initialised Sentry")
		// defer this to flush buffered events
		defer sentry.Flush(2 * time.Second)
	}
	b.Hub = sentry.CurrentHub()
	if config.SentryURL == "" {
		b.Hub = nil
	}

	if config.GuildLogWebhook.ID.IsValid() {
		b.GuildLogWebhook = webhook.New(config.GuildLogWebhook.ID, config.GuildLogWebhook.Token)
	}

	// set the prefix checker
	b.Router.Prefixer = b.CheckPrefix
	// set permission checker
	b.Router.PermissionCheck = b.CheckPermissions

	b.Router.ShardManager.ForEach(func(s shard.Shard) {
		state := s.(*state.State)

		// add guild create handler
		state.AddHandler(b.GuildCreate)

		// add guild remove handler
		state.PreHandler = handler.New()
		state.PreHandler.AddSyncHandler(b.guildDelete)

		// add message create handler
		state.AddHandler(b.MessageCreate)
		// add interaction handler
		state.AddHandler(b.interactionCreate)

		// add member update handler (this isn't handled by default apparently?)
		// state.AddHandler(b.guildMemberUpdate)

		// add cache handlers
		// state.AddHandler(b.requestGuildMembers)
		// state.AddHandler(b.guildMemberChunk)
		// state.AddHandler(b.memberUpdateEvent)
		// state.AddHandler(b.memberAddEvent)
		// state.AddHandler(b.memberRemoveEvent)

		state.AddHandler(func(ev *ws.BackgroundErrorEvent) {
			b.Sugar.Errorf("background gateway error: %v", ev.Err)
		})
	})

	return b
}

// Add adds a module to the bot
func (bot *Bot) Add(fns ...func(*Bot) (string, []*bcr.Command)) {
	for _, fn := range fns {
		m, c := fn(bot)

		// sort the list of commands
		sort.Sort(bcr.Commands(c))

		// add the module
		bot.Modules = append(bot.Modules, &botModule{
			name:     m,
			commands: c,
		})
	}
}

type botModule struct {
	name     string
	commands bcr.Commands
}

// String returns the module's name
func (b botModule) String() string {
	return b.name
}

// Commands returns a list of commands
func (b *botModule) Commands() []*bcr.Command {
	return b.commands
}

// PagedEmbed ...
func (bot *Bot) PagedEmbed(ctx *bcr.Context, embeds []discord.Embed, timeout time.Duration) (msg *discord.Message, err error) {
	reactions, _ := bot.DB.UserBoolGet(ctx.Author.ID, "reaction_pages")

	if reactions {
		msg, _, err = ctx.PagedEmbedTimeout(embeds, false, timeout)
		return
	}

	msg, _, err = ctx.ButtonPages(embeds, timeout)
	return
}

// Start wraps around Router.ShardManager.Open()
func (bot *Bot) Start(ctx context.Context) error {
	return bot.Router.ShardManager.Open(ctx)
}
