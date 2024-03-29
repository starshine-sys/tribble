// SPDX-License-Identifier: AGPL-3.0-only
package static

import (
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/pkgo/v2"

	flag "github.com/spf13/pflag"
)

var keycaps = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}

func (bot *Bot) poll(ctx *bcr.Context) (err error) {
	question := ctx.Args[0]
	options := ctx.Args[1:]
	if len(options) > 10 {
		_, err = ctx.Send("Too many options, maximum 10.")
		return err
	}

	var desc string
	for i, o := range options {
		desc += fmt.Sprintf("%v %v\n", keycaps[i], o)
	}

	if len(desc) > 4000 {
		_, err = ctx.Send("Embed description too long.")
		return err
	}
	if len(question) > 256 {
		_, err = ctx.Send("Question too long (maximum 256 characters)")
		return err
	}

	msg, err := ctx.Send("", discord.Embed{
		Title:       question,
		Description: desc,
		Footer: &discord.EmbedFooter{
			Text: fmt.Sprintf("%v (%v)", ctx.Author.Tag(), ctx.Author.ID),
			Icon: ctx.Author.AvatarURL(),
		},
	})
	if err != nil {
		return err
	}

	for i := 0; i < len(options); i++ {
		err = ctx.State.React(ctx.Channel.ID, msg.ID, discord.APIEmoji(keycaps[i]))
		if err != nil {
			return err
		}
	}
	return
}

func (bot *Bot) quickpoll(ctx *bcr.Context) (err error) {
	var reacts int
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.IntVarP(&reacts, "options", "o", -1, "How many options to have")
	fs.Parse(ctx.Args)
	ctx.Args = fs.Args()

	// i cant be bothered to write actual error messages so well just do this
	if reacts != -1 && reacts < 2 {
		_, err = ctx.Send("less than 2 options? do you really think i can do something with that?")
		return err
	} else if reacts != -1 && reacts > 10 {
		_, err = ctx.Send("look buddy i can't help you with that, that's way too many choices, i can only count to 10")
		return err
	}

	// indicate that were processing
	ctx.State.React(ctx.Channel.ID, ctx.Message.ID, "🔄")
	id := ctx.Message.ID

	// wait a second for pk
	time.Sleep(time.Second)

	m, err := bot.PK.Message(pkgo.Snowflake(ctx.Message.ID))
	if err == nil {
		id = discord.MessageID(m.ID)
	} else {
		ctx.State.Unreact(ctx.Channel.ID, ctx.Message.ID, "🔄")
	}

	if reacts < 2 || reacts > 10 {
		err = ctx.State.React(ctx.Channel.ID, id, ":greentick:754647778390442025")
		if err != nil {
			return err
		}
		err = ctx.State.React(ctx.Channel.ID, id, ":redtick:754647803837415444")
		if err != nil {
			return err
		}
	} else {
		for i := 0; i < reacts; i++ {
			err = ctx.State.React(ctx.Channel.ID, id, discord.APIEmoji(keycaps[i]))
			if err != nil {
				return err
			}
		}
	}
	return
}
