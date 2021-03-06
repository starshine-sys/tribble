package modlog

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/starshine-sys/bcr"
)

func (bot *ModLog) reason(ctx *bcr.Context) (err error) {
	id := 0

	if equalFoldAny(ctx.Args[0], "latest", "last", "l") {
		err = bot.DB.Pool.QueryRow(context.Background(), "select id from mod_log where server_id = $1 order by id desc limit 1", ctx.Guild.ID).Scan(&id)
		if err != nil {
			if errors.Cause(err) == pgx.ErrNoRows {
				ctx.Replyc(bcr.ColourRed, "This server has no mod log entries.", nil)
				return
			}
			return bot.Report(ctx, err)
		}
	} else {
		id, err = strconv.Atoi(ctx.Args[0])
		if err != nil {
			_, err = ctx.Replyc(bcr.ColourRed, "Couldn't parse ``%v`` as a number.", bcr.EscapeBackticks(ctx.Args[0]))
			return
		}
	}

	exists := false
	err = bot.DB.Pool.QueryRow(context.Background(), "select exists(select * from mod_log where id = $1 and server_id = $2)", id, ctx.Guild.ID).Scan(&exists)
	if err != nil {
		return bot.Report(ctx, err)
	}

	if !exists {
		_, err = ctx.Replyc(bcr.ColourRed, "There's no mod log with the ID %v.", id)
		return
	}

	reason := strings.TrimSpace(strings.TrimPrefix(ctx.RawArgs, ctx.Args[0]))
	if reason == ctx.RawArgs {
		reason = strings.Join(ctx.Args[1:], " ")
	}

	var oldEntry, entry Entry
	// get original
	err = pgxscan.Get(context.Background(), bot.DB.Pool, &oldEntry, "select * from mod_log where id = $1 and server_id = $2", id, ctx.Guild.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	oldReason := oldEntry.Reason
	entryReason := reason
	if oldEntry.ActionType == "channelban" || oldEntry.ActionType == "unchannelban" {
		oldReason = strings.Join(strings.Split(oldEntry.Reason, ":")[1:], ":")

		ch := strings.Split(oldEntry.Reason, ":")
		entryReason = ch[0] + ": " + reason
	}

	// update and get the new one
	err = pgxscan.Get(context.Background(), bot.DB.Pool, &entry, "update mod_log set reason = $1, mod_id = $2 where id = $3 and server_id = $4 returning *", entryReason, ctx.Author.ID, id, ctx.Guild.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	if !entry.ChannelID.IsValid() || !entry.MessageID.IsValid() {
		_, err = ctx.Replyc(bcr.ColourOrange, "I updated the reason, but couldn't update the log message.")
		return
	}

	msg, err := ctx.State.Message(entry.ChannelID, entry.MessageID)
	if err != nil {
		_, err = ctx.Replyc(bcr.ColourOrange, "I updated the reason, but couldn't update the log message.")
		return
	}

	content := strings.NewReplacer(oldReason, reason).Replace(msg.Content)
	// remove the last line so we can replace it with the new responsible moderator
	s := strings.Split(content, "\n")
	s = s[:len(s)-1]
	s = append(s, fmt.Sprintf("**Responsible moderator:** %v (%v)", ctx.Author.Tag(), ctx.Author.ID))
	content = strings.Join(s, "\n")

	_, err = ctx.Edit(msg, content, false)
	if err != nil {
		_, err = ctx.Replyc(bcr.ColourOrange, "I updated the reason, but couldn't update the log message.")
		return
	}

	// ignore errors on reacting, the message might be deleted already
	ctx.State.React(ctx.Message.ChannelID, ctx.Message.ID, "???")

	return
}

func equalFoldAny(s string, options ...string) bool {
	for _, o := range options {
		if strings.EqualFold(s, o) {
			return true
		}
	}
	return false
}
