// SPDX-License-Identifier: AGPL-3.0-only
package moderation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/dischtml"
)

type jsonOut struct {
	PackVersion int       `json:"packVersion"`
	Date        time.Time `json:"timestamp"`

	Messages []discord.Message `json:"messages"`
	Channel  *discord.Channel  `json:"channel"`
}

func (bot *Bot) transcript(ctx *bcr.Context) (err error) {
	out, _ := ctx.Flags.GetString("out")
	limit, _ := ctx.Flags.GetUint("limit")
	outputJSON, _ := ctx.Flags.GetBool("json")
	outputHTML, _ := ctx.Flags.GetBool("html")
	if limit > 2000 || limit == 0 {
		isOwner := false
		for _, o := range bot.Config.Owners {
			if o == ctx.Author.ID {
				isOwner = true
				break
			}
		}

		if !isOwner {
			_, err = ctx.Replyc(bcr.ColourRed, ":x: You can only make a transcript of a maximum of 2000 messages.")
			return
		}
	}

	outCh := ctx.Channel
	if out != "" {
		outCh, err = ctx.ParseChannel(out)
		if err != nil {
			_, err = ctx.Replyc(bcr.ColourRed, "Couldn't find a channel named `%v`", out)
			return
		}
	}

	ch, err := ctx.ParseChannel(ctx.Args[0])
	if err != nil || ch.GuildID != ctx.Message.GuildID || (ch.Type != discord.GuildText && ch.Type != discord.GuildNews) {
		_, err = ctx.Replyc(bcr.ColourRed, "Channel `%v` not found.", ctx.Args[0])
		return
	}

	err = ctx.SendfX("Archiving channel %v (#%v)...", ch.Mention(), ch.Name)
	if err != nil {
		return
	}

	msgs, err := ctx.State.Messages(ch.ID, limit)
	if err != nil {
		_, err = ctx.Send("I couldn't fetch all messages in this channel, aborting.")
		return
	}

	if outputJSON {
		return bot.transcriptJSON(ctx, outCh.ID, ch, msgs)
	}

	if outputHTML {
		return bot.transcriptHTML(ctx, outCh.ID, ch, msgs)
	}

	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].ID < msgs[j].ID
	})

	var buf []string
	var participants []discord.UserID
	in := func(u discord.UserID) bool {
		for _, p := range participants {
			if u == p {
				return true
			}
		}
		return false
	}

	buf = append(buf, fmt.Sprintf(`#%v (%v)
Guild: %v (%v)
Messages: %v
`, ch.Name, ch.ID, ctx.Guild.Name, ctx.Guild.ID, len(msgs)))

	for _, m := range msgs {
		b := fmt.Sprintf(`--------------------------------------------------------------------------------
[%v] %v#%v (%v)
%v`, m.Timestamp.Time().UTC().Format("2006-01-02 15:05:05"), m.Author.Tag(), m.Author.ID, m.Content)

		if len(m.Embeds) > 0 {
			bt, err := json.Marshal(&m.Embeds[0])
			if err == nil {
				b += "\n\n" + string(bt)
			}
		}

		if len(m.Attachments) > 0 {
			b += "\n\nAttachments:\n"
			for _, a := range m.Attachments {
				b += a.URL + "\n"
			}
		}

		if !in(m.Author.ID) {
			participants = append(participants, m.Author.ID)
		}

		buf = append(buf, b+"\n")
	}

	text := strings.Join(buf, "\n")

	e := discord.Embed{
		Title: "Transcript of #" + ch.Name,

		Author: &discord.EmbedAuthor{
			Name: ctx.Author.Tag(),
			Icon: ctx.Author.AvatarURL(),
		},

		Fields: []discord.EmbedField{
			{
				Name:   "Messages",
				Value:  fmt.Sprint(len(msgs)),
				Inline: true,
			},
			{
				Name:   "Name",
				Value:  "#" + ch.Name,
				Inline: true,
			},
			{
				Name:   "Transcript creator",
				Value:  ctx.Author.Mention(),
				Inline: true,
			},
		},

		Footer: &discord.EmbedFooter{
			Text: "ID: " + ch.ID.String(),
		},
		Timestamp: discord.NowTimestamp(),

		Color: bcr.ColourBlurple,
	}

	{
		var buf string

		for i, u := range participants {
			if len(buf) > 900 {
				buf += fmt.Sprintf("\n```Too many to list (showing %v/%v)```", i, len(participants))
				break
			}

			buf += u.Mention() + ", "
		}

		e.Fields = append(e.Fields, discord.EmbedField{
			Name:  "Participants",
			Value: buf,
		})
	}

	_, err = ctx.NewMessage(outCh.ID).Embeds(e).AddFile(
		fmt.Sprintf("transcript-%v.txt", ctx.Channel.Name), strings.NewReader(text),
	).Send()
	if err != nil {
		return bot.Report(ctx, err)
	}

	_, err = ctx.Reply("Transcript complete!")
	return
}

func (bot *Bot) transcriptJSON(ctx *bcr.Context, out discord.ChannelID, ch *discord.Channel, msgs []discord.Message) (err error) {
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].ID > msgs[j].ID
	})

	data := jsonOut{
		PackVersion: 3,
		Date:        time.Now().UTC(),
		Messages:    msgs,
		Channel:     ch,
	}

	b, err := json.Marshal(data)

	s := fmt.Sprintf("Archived channel %v (#%v, %v) with %v messages.", ch.Mention(), ch.Name, ch.ID, len(msgs))

	_, err = ctx.State.SendMessageComplex(out, api.SendMessageData{
		Content: s,
		Files: []sendpart.File{{
			Name:   fmt.Sprintf("darchive-%v-%v-%v.json", ch.Name, ch.ID, time.Now().UTC().Format("2006-01-02-1504")),
			Reader: bytes.NewReader(b),
		}},
	})
	return
}

func (bot *Bot) transcriptHTML(ctx *bcr.Context, outCh discord.ChannelID, ch *discord.Channel, msgs []discord.Message) (err error) {
	channels, err := ctx.State.Channels(ctx.Guild.ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	var users []discord.User

	for _, m := range msgs {
		for _, mention := range m.Mentions {
			users = append(users, mention.User)
		}
	}

	c := dischtml.Converter{
		Guild:    *ctx.Guild,
		Channels: channels,
		Roles:    ctx.Guild.Roles,
		Members:  bot.Members(ctx.Guild.ID),
		Users:    users,
	}

	html, err := c.ConvertHTML(msgs)
	if err != nil {
		return bot.Report(ctx, err)
	}

	out, err := dischtml.Wrap(*ctx.Guild, *ch, html, len(msgs))
	if err != nil {
		return bot.Report(ctx, err)
	}

	s := fmt.Sprintf("Archived channel %v (#%v, %v) with %v messages.", ch.Mention(), ch.Name, ch.ID, len(msgs))

	_, err = ctx.State.SendMessageComplex(outCh, api.SendMessageData{
		Content: s,
		Files: []sendpart.File{{
			Name:   fmt.Sprintf("%v-%v-%v.html", ch.Name, ch.ID, time.Now().UTC().Format("2006-01-02-1504")),
			Reader: strings.NewReader(out),
		}},
	})
	return
}
