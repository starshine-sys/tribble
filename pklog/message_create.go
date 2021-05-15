package pklog

import (
	"context"
	"time"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/starshine-sys/pkgo"
)

var pk = pkgo.NewSession(nil)

// messageCreate is used as a backup for pkMessageCreate in case proxy logging isn't enabled.
func (bot *Bot) messageCreate(m *gateway.MessageCreateEvent) {
	var shouldLog bool
	bot.DB.Pool.QueryRow(context.Background(), "select (pk_log_channel != 0) from servers where id = $1", m.GuildID).Scan(&shouldLog)
	if !shouldLog {
		return
	}

	// only check webhook messages
	if !m.WebhookID.IsValid() {
		return
	}

	// wait 5 seconds
	time.Sleep(5 * time.Second)

	// check if the message exists in the database; if so, return
	_, err := bot.Get(m.ID)
	if err == nil {
		return
	}

	pkm, err := pk.GetMessage(m.ID.String())
	if err != nil {
		// Message is either not proxied or we got an error from the PK API. Either way, return
		return
	}

	u, _ := discord.ParseSnowflake(pkm.Sender)

	msg := Message{
		MsgID:     m.ID,
		UserID:    discord.UserID(u),
		ChannelID: m.ChannelID,
		ServerID:  m.GuildID,

		Username: m.Author.Username,
		Member:   pkm.Member.ID,
		System:   pkm.System.ID,

		Content: m.Content,
	}

	// insert the message, ignore errors as those shouldn't impact anything
	bot.Insert(msg)
}