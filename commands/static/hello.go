package static

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/starshine-sys/bcr"
	"github.com/starshine-sys/pkgo"
)

var greetings = []string{"Hello", "Heya", "Hi", "Hiya"}

// yeah this won't work on any other instance of the bot sadly
var emotes = []string{"👋", "<:MenheraWave:807587508623507456>", "<a:ameowcomfywave:807587518857216021>"}

func (bot *Bot) hello(ctx *bcr.Context) (err error) {
	// sleep for a second to give PK time to process the message
	time.Sleep(1 * time.Second)

	var name string
	m, err := bot.PK.Message(pkgo.Snowflake(ctx.Message.ID))
	// if there's a non-nil error, chances are PK hasn't registered the message yet
	// so just fall back to a normal user
	if err != nil {
		member, err := ctx.ParseMember(ctx.Author.ID.String())
		if err != nil {
			name = ctx.Author.Username
		} else {
			if member.Nick != "" {
				name = member.Nick
			} else {
				name = ctx.Author.Username
			}
		}
	} else {
		name = m.Member.Name
	}

	// spaghetti Lite™ to get some more randomness
	greeting := fmt.Sprintf(
		"%v, %v!",
		greetings[rand.Intn(len(greetings))],
		name,
	)
	if r := rand.Intn(3); r == 1 {
		if len(emotes) != 0 {
			if r := rand.Intn(2); r == 1 {
				greeting = fmt.Sprintf("%v %v", greeting, emotes[rand.Intn(len(emotes))])
			} else {
				greeting = fmt.Sprintf("%v %v", emotes[rand.Intn(len(emotes))], greeting)
			}
		}
	}

	_, err = ctx.NewMessage().Content(greeting).BlockMentions().Send()
	return err
}

func (bot *Bot) helloSlash(v bcr.Contexter) (err error) {
	ctx := v.(*bcr.SlashContext)

	name := ""
	if ctx.Event.Member != nil {
		name = ctx.Event.Member.Nick
		if name == "" {
			name = ctx.Event.Member.User.Username
		}
	} else {
		name = ctx.Event.User.Username
	}

	greeting := fmt.Sprintf(
		"%v, %v!",
		greetings[rand.Intn(len(greetings))],
		name,
	)
	if r := rand.Intn(3); r == 1 {
		if len(emotes) != 0 {
			if r := rand.Intn(2); r == 1 {
				greeting = fmt.Sprintf("%v %v", greeting, emotes[rand.Intn(len(emotes))])
			} else {
				greeting = fmt.Sprintf("%v %v", emotes[rand.Intn(len(emotes))], greeting)
			}
		}
	}
	return ctx.SendEphemeral(greeting)
}
