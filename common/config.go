// SPDX-License-Identifier: AGPL-3.0-only
package common

import "github.com/diamondburned/arikawa/v3/discord"

// BotConfig ...
type BotConfig struct {
	Token       string
	DatabaseURL string `yaml:"database_url"`
	SentryURL   string `yaml:"sentry_url"`

	Prefixes []string
	Owners   []discord.UserID

	GlobalQuotes bool `yaml:"global_quotes"`

	NoSyncCommands bool              `yaml:"no_sync_commands"`
	SyncCommandsIn []discord.GuildID `yaml:"sync_commands_in"`

	GuildLogWebhook Webhook `yaml:"guild_log"`

	Branding struct {
		Name string

		Private  bool
		PublicID discord.UserID `yaml:"public_id"`

		SupportServer string `yaml:"support_server"`
	}

	DMs struct {
		Open    bool
		Webhook Webhook

		BlockedUsers []discord.UserID `yaml:"blocked_users"`
	} `yaml:"dms"`
}

// Webhook is a single webhook config
type Webhook struct {
	ID    discord.WebhookID
	Token string
}
