package cc

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/starshine-sys/bcr"
	lua "github.com/yuin/gopher-lua"
)

// This file contains functions used for custom commands.

func (s *State) setMessageFuncs() {
	s.ls.SetGlobal("send_message", s.ls.NewFunction(s.sendMessage))
	s.ls.SetGlobal("react", s.ls.NewFunction(s.react))
}

func (s *State) sendMessage(ls *lua.LState) int {
	chID := s.ctx.Message.ChannelID

	// first argument is channel ID
	id := s._getChannelID(ls, 1)
	if id != nil {
		chID = *id
	}

	ch, err := s.ctx.State.Channel(chID)
	if err != nil {
		ls.RaiseError("channel not found in state")
		return 1
	}

	if ch.GuildID != s.ctx.Message.GuildID {
		ls.RaiseError("channel is not in this guild")
		return 1
	}

	data := api.SendMessageData{
		AllowedMentions: allowedMentions(s.ctx),
	}

	v := ls.Get(2)
	switch v.Type() {
	case lua.LTString:
		data.Content = string(v.(lua.LString))
	default:
		ls.RaiseError("send_message does not yet support non-string messages")
		return 1
	}

	msg, err := s.ctx.State.SendMessageComplex(chID, data)
	if err != nil {
		ls.RaiseError("error sending message: %s", err.Error())
		return 1
	}

	ls.Push(lua.LString(msg.ID.String()))
	return 1
}

func (s *State) react(ls *lua.LState) int {
	chID := s.ctx.Message.ChannelID
	mID := s.ctx.Message.ID

	// first argument is channel ID
	chIDp := s._getChannelID(ls, 1)
	if chIDp != nil {
		chID = *chIDp
	}

	mIDp := s._getMessageID(ls, 2)
	if mIDp != nil {
		mID = *mIDp
	}

	msg, err := s.ctx.State.Message(chID, mID)
	if err != nil {
		ls.RaiseError("message %d/%d not found", chID, mID)
		return 0
	}
	if msg.GuildID != s.ctx.Message.GuildID {
		ls.RaiseError("message %d/%d not in this guild", chID, mID)
		return 0
	}

	react := s._getString(ls, 3)

	err = s.ctx.State.React(msg.ChannelID, msg.ID, discord.APIEmoji(react))
	if err != nil {
		ls.RaiseError("error reacting to message: %s", err.Error())
	}
	return 0
}

func allowedMentions(ctx *bcr.Context) *api.AllowedMentions {
	mentions := &api.AllowedMentions{
		Parse: []api.AllowedMentionType{api.AllowUserMention},
	}

	if ctx.Guild == nil {
		return mentions
	}

	for _, r := range ctx.Guild.Roles {
		if r.Mentionable {
			mentions.Roles = append(mentions.Roles, r.ID)
		}
	}
	return mentions
}
