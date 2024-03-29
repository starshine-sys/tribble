// SPDX-License-Identifier: AGPL-3.0-only
package cc

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

// This file contains functions used for custom commands.

func (s *State) setArgumentFuncs() {
	t := s.ls.NewTable()
	t.RawSetString("pop", s.ls.NewFunction(s.pop))
	t.RawSetString("peek", s.ls.NewFunction(s.peek))
	t.RawSetString("has_next", s.ls.NewFunction(s.hasNext))
	t.RawSetString("remainder", s.ls.NewFunction(s.remainder))
	t.RawSetString("next_channel", s.ls.NewFunction(s.nextChannel))
	t.RawSetString("next_channel_check", s.ls.NewFunction(s.nextChannelCheck))

	s.ls.SetGlobal("current_timestamp", s.ls.NewFunction(s.currentTimestamp))
	s.ls.SetGlobal("args", t)
}

// current_timestamp
// - 0 arguments
// - returns:
// 1. string of current timestamp in RFC3339 format
func (s *State) currentTimestamp(ls *lua.LState) int {
	str := time.Now().UTC().Format(time.RFC3339)

	ls.Push(lua.LString(str))
	return 1
}

// pop
// - 0 arguments
// - returns:
// 1. string
func (s *State) pop(ls *lua.LState) int {
	ls.Push(lua.LString(s.params.Pop()))
	return 1
}

// peek
// - 0 arguments
// - returns:
// 1. string
func (s *State) peek(ls *lua.LState) int {
	ls.Push(lua.LString(s.params.Peek()))
	return 1
}

// has_next
// - 0 arguments
// - returns:
// 1. bool
func (s *State) hasNext(ls *lua.LState) int {
	ls.Push(lua.LBool(s.params.HasNext()))
	return 1
}

// remainder
// - 0 arguments
// - returns:
// 1. string
func (s *State) remainder(ls *lua.LState) int {
	ls.Push(lua.LString(s.params.Remainder(false)))
	return 1
}

// next_channel
// - 0 arguments
// - returns:
// 1. channel
func (s *State) nextChannel(ls *lua.LState) int {
	arg := s.params.Pop()
	ch, err := s.ctx.ParseChannel(arg)
	if err != nil {
		ls.RaiseError("channel %q not found", arg)
		return 1
	}
	if ch.GuildID != s.ctx.Message.GuildID {
		ls.RaiseError("channel must be in this guild")
		return 1
	}

	ls.Push(s.channelToLua(*ch))
	return 1
}

// next_channel_check
// - 0 arguments
// - returns:
// 1. channel
// 2. ok
func (s *State) nextChannelCheck(ls *lua.LState) int {
	arg := s.params.Pop()
	ch, err := s.ctx.ParseChannel(arg)
	if err != nil {
		ls.Push(lua.LNil)
		ls.Push(lua.LBool(false))
		return 2
	}
	if ch.GuildID != s.ctx.Message.GuildID {
		ls.Push(lua.LNil)
		ls.Push(lua.LBool(false))
		return 2
	}

	ls.Push(s.channelToLua(*ch))
	ls.Push(lua.LBool(true))
	return 2
}
