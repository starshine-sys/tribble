package modlog

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/georgysavva/scany/pgxscan"
)

// Entry ...
type Entry struct {
	ID       int64           `json:"id"`
	ServerID discord.GuildID `json:"-"`

	UserID discord.UserID `json:"user_id"`
	ModID  discord.UserID `json:"mod_id"`

	ActionType ActionType `json:"action_type"`
	Reason     string     `json:"reason,omitempty"`

	Time time.Time `json:"timestamp"`
}

func (bot *ModLog) logChannelFor(guildID discord.GuildID) (logChannel discord.ChannelID) {
	bot.DB.Pool.QueryRow(context.Background(), "select mod_log_channel from servers where id = $1", guildID).Scan(&logChannel)
	return
}

// ActionType ...
type ActionType string

// Constants for action types
const (
	ActionBan  ActionType = "ban"
	ActionKick ActionType = "kick"
	ActionWarn ActionType = "warn"
)

func (bot *ModLog) nextID(guildID discord.GuildID) (id int64) {
	bot.DB.Pool.QueryRow(context.Background(), "select id from mod_log where server_id = $1 order by id desc limit 1", guildID).Scan(&id)
	return id + 1
}

func (bot *ModLog) insertEntry(guildID discord.GuildID, user, mod discord.UserID, actionType ActionType, reason string) (log *Entry, err error) {
	if reason == "" {
		reason = "N/A"
	}

	log = &Entry{}

	err = pgxscan.Get(context.Background(), bot.DB.Pool, log, `insert into mod_log
	(id, server_id, user_id, mod_id, action_type, reason)
	values ($1, $2, $3, $4, $5, $6) returning *`, bot.nextID(guildID), guildID, user, mod, actionType, reason)
	if err != nil {
		return nil, err
	}
	return
}

// Channelban logs a channel ban
func (bot *ModLog) Channelban(guildID discord.GuildID, channel discord.ChannelID, userID, modID discord.UserID, reason string) (err error) {
	if reason == "" {
		reason = "N/A"
	}

	entry, err := bot.insertEntry(guildID, userID, modID, "channelban", fmt.Sprintf("%v: %v", channel.Mention(), reason))
	if err != nil {
		return err
	}

	ch := bot.logChannelFor(guildID)
	if !ch.IsValid() {
		return
	}

	if len(reason) > 1000 {
		reason = reason[:1000] + "..."
	}

	user, err := bot.State.User(userID)
	if err != nil {
		return err
	}
	mod, err := bot.State.User(modID)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(`**Channel ban for %v | Case %v**
**User:** %v#%v (%v)
**Reason:** %v
**Responsible moderator:** %v#%v (%v)`, channel.Mention(), entry.ID, user.Username, user.Discriminator, entry.UserID, reason, mod.Username, mod.Discriminator, entry.ModID)

	_, err = bot.State.SendText(ch, text)
	return
}

// Unchannelban logs a channel unban
func (bot *ModLog) Unchannelban(guildID discord.GuildID, channel discord.ChannelID, userID, modID discord.UserID, reason string) (err error) {
	if reason == "" {
		reason = "N/A"
	}

	entry, err := bot.insertEntry(guildID, userID, modID, "unchannelban", fmt.Sprintf("%v: %v", channel.Mention(), reason))
	if err != nil {
		return err
	}

	ch := bot.logChannelFor(guildID)
	if !ch.IsValid() {
		return
	}

	if len(reason) > 1000 {
		reason = reason[:1000] + "..."
	}

	user, err := bot.State.User(userID)
	if err != nil {
		return err
	}
	mod, err := bot.State.User(modID)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(`**Channel unban for %v | Case %v**
**User:** %v#%v (%v)
**Reason:** %v
**Responsible moderator:** %v#%v (%v)`, channel.Mention(), entry.ID, user.Username, user.Discriminator, entry.UserID, reason, mod.Username, mod.Discriminator, entry.ModID)

	_, err = bot.State.SendText(ch, text)
	return
}

// Warn logs a warn
func (bot *ModLog) Warn(guildID discord.GuildID, userID, modID discord.UserID, reason string) (err error) {
	entry, err := bot.insertEntry(guildID, userID, modID, "warn", reason)
	if err != nil {
		return err
	}

	ch := bot.logChannelFor(guildID)
	if !ch.IsValid() {
		return
	}

	if len(entry.Reason) > 1000 {
		entry.Reason = entry.Reason[:1000] + "..."
	}

	user, err := bot.State.User(userID)
	if err != nil {
		return err
	}
	mod, err := bot.State.User(modID)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(`**Warn | Case %v**
**User:** %v#%v (%v)
**Reason:** %v
**Responsible moderator:** %v#%v (%v)`, entry.ID, user.Username, user.Discriminator, entry.UserID, entry.Reason, mod.Username, mod.Discriminator, entry.ModID)

	_, err = bot.State.SendText(ch, text)
	return
}

// Ban logs a ban
func (bot *ModLog) Ban(guildID discord.GuildID, userID, modID discord.UserID, reason string) (err error) {
	entry, err := bot.insertEntry(guildID, userID, modID, "ban", reason)
	if err != nil {
		return err
	}

	ch := bot.logChannelFor(guildID)
	if !ch.IsValid() {
		return
	}

	if len(entry.Reason) > 1000 {
		entry.Reason = entry.Reason[:1000] + "..."
	}

	user, err := bot.State.User(userID)
	if err != nil {
		return err
	}
	mod, err := bot.State.User(modID)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(`**Ban | Case %v**
**User:** %v#%v (%v)
**Reason:** %v
**Responsible moderator:** %v#%v (%v)`, entry.ID, user.Username, user.Discriminator, entry.UserID, entry.Reason, mod.Username, mod.Discriminator, entry.ModID)

	_, err = bot.State.SendText(ch, text)
	return
}

// Unban logs a unban
func (bot *ModLog) Unban(guildID discord.GuildID, userID, modID discord.UserID, reason string) (err error) {
	entry, err := bot.insertEntry(guildID, userID, modID, "unban", reason)
	if err != nil {
		return err
	}

	ch := bot.logChannelFor(guildID)
	if !ch.IsValid() {
		return
	}

	if len(entry.Reason) > 1000 {
		entry.Reason = entry.Reason[:1000] + "..."
	}

	user, err := bot.State.User(userID)
	if err != nil {
		return err
	}
	mod, err := bot.State.User(modID)
	if err != nil {
		return err
	}

	text := fmt.Sprintf(`**Unban | Case %v**
**User:** %v#%v (%v)
**Reason:** %v
**Responsible moderator:** %v#%v (%v)`, entry.ID, user.Username, user.Discriminator, entry.UserID, entry.Reason, mod.Username, mod.Discriminator, entry.ModID)

	_, err = bot.State.SendText(ch, text)
	return
}