package static

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/starshine-sys/bcr"
)

func (bot *Bot) bubble(ctx *bcr.Context) (err error) {
	var (
		prepop bool
		size   int
	)

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolVar(&prepop, "prepop", false, "Whether to pre-pop some bubbles.")
	fs.IntVar(&size, "size", 10, "Size of the generated bubble wrap.")
	err = fs.Parse(ctx.Args)
	if err != nil {
		_, err = ctx.Send("Invalid arguments provided, valid arguments are `-prepop` and `-size int`.")
		return
	}

	if size > 13 {
		_, err = ctx.Sendf("A size of %v is too large, maximum 13.", size)
		return
	} else if size < 1 {
		_, err = ctx.Sendf("A size of %v is too small, minimum 1.", size)
		return
	}

	var out string
	for i := size; i > 0; i-- {
		for j := size; j > 0; j-- {
			if prepop {
				if j != 1 && j != size && i != 1 && i != size {
					if rand.Intn(6) == 5 {
						out += "pop"
					} else {
						out += "||pop||"
					}
				} else {
					out += "||pop||"
				}
			} else {
				out += "||pop||"
			}
		}
		out += "\n"
	}
	_, err = ctx.Send(out)
	return err
}

func (bot *Bot) bubbleSlash(v bcr.Contexter) (err error) {
	ctx := v.(*bcr.SlashContext)

	size := 10
	prepop := false
	ephemeral := false

	for _, option := range ctx.Event.Data.Options {
		switch option.Name {
		case "size":
			size, err = strconv.Atoi(option.Value)
			if err != nil {
				err = ctx.SendEphemeral("I couldn't parse your input as a valid size.")
				if err != nil {
					bot.Sugar.Errorf("Error sending message: %v", err)
				}
			}
		case "prepop":
			fmt.Printf("prepop value: `%v`\n", option.Value)
		case "ephemeral":
			fmt.Printf("ephemeral value: `%v`\n", option.Value)
		}
	}

	if size > 13 {
		return ctx.SendEphemeral(
			fmt.Sprintf("A size of %v is too large, maximum 13.", size),
		)
	} else if size < 1 {
		return ctx.SendEphemeral(
			fmt.Sprintf("A size of %v is too small, minimum 1.", size),
		)
	}

	var out string
	for i := size; i > 0; i-- {
		for j := size; j > 0; j-- {
			if prepop {
				if j != 1 && j != size && i != 1 && i != size {
					if rand.Intn(6) == 5 {
						out += "pop"
					} else {
						out += "||pop||"
					}
				} else {
					out += "||pop||"
				}
			} else {
				out += "||pop||"
			}
		}
		out += "\n"
	}
	if ephemeral {
		return ctx.SendEphemeral(out)
	}
	return ctx.SendX(out)
}
