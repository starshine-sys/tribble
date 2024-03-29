// SPDX-License-Identifier: AGPL-3.0-only
package info

import (
	"fmt"
	"time"

	"github.com/starshine-sys/bcr"
)

func (bot *Bot) ping(ctx *bcr.Context) (err error) {
	t := time.Now()

	m, err := ctx.Send("Pong!")
	if err != nil {
		return err
	}

	latency := time.Since(t).Round(time.Millisecond)

	// this will return 0ms in the first minute after the bot is restarted
	// can't do much about that though
	heartbeat := ctx.Session().Gateway().EchoBeat().Sub(ctx.Session().Gateway().SentBeat()).Round(time.Millisecond)

	_, err = ctx.Edit(m, fmt.Sprintf("Heartbeat: %v | Message: %v", heartbeat, latency), false)
	return err
}
