// SPDX-License-Identifier: AGPL-3.0-only
package quotes

import (
	"context"

	"github.com/starshine-sys/bcr"
)

func (bot *Bot) disable(ctx *bcr.Context) (err error) {
	blocked := bot.isUserBlocked(ctx.User().ID)

	if blocked {
		yes, timeout := ctx.ConfirmButton(ctx.User().ID, bcr.ConfirmData{
			Message:   "Quoting is currently **disabled** for your messages. Do you want to opt in again?",
			YesPrompt: "Unblock quotes",
		})
		if !yes || timeout {
			return ctx.SendX("Cancelled.")
		}

		_, err = bot.DB.Pool.Exec(context.Background(), "delete from quote_block where user_id = $1", ctx.User().ID)
		if err != nil {
			return bot.Report(ctx, err)
		}

		return ctx.SendX("Your messages can now be quoted again!")
	}

	yes, timeout := ctx.ConfirmButton(ctx.User().ID, bcr.ConfirmData{
		Message:   "Quoting is currently **enabled** for your messages. Do you want to opt out?",
		YesPrompt: "Block quotes",
	})
	if !yes || timeout {
		return ctx.SendX("Cancelled.")
	}

	_, err = bot.DB.Pool.Exec(context.Background(), "insert into quote_block (user_id) values ($1)", ctx.User().ID)
	if err != nil {
		return bot.Report(ctx, err)
	}

	return ctx.SendX("Your messages can no longer be quoted!")
}
