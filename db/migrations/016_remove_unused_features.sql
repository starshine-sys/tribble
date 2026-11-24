-- +migrate Up

-- 2026-06-11
-- Remove a bunch of unused features to make the bot *slightly* more maintainable

-- lua custom commands
drop table custom_commands;
-- tickets
drop table tickets;
drop table ticket_categories;
-- role/derole commands
drop table role_categories;
-- gatekeeper
drop table gatekeeper;
alter table servers drop column welcome_channel;
alter table servers drop column welcome_message;
alter table servers drop column member_role;
-- approval
alter table servers drop column approve_remove_roles;
alter table servers drop column approve_add_roles;
alter table servers drop column approve_welcome_channel;
alter table servers drop column approve_welcome_message;
