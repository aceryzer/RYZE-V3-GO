package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"time"

	"github.com/bwmarrin/discordgo"
)

func ConfigureRun(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	p := cfg.Prefix
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}

	// Uptime helper
	uptime := func() string {
		duration := time.Since(startTime)
		h := int(duration.Hours())
		min := int(duration.Minutes()) % 60
		sec := int(duration.Seconds()) % 60
		return fmt.Sprintf("%dh %dm %ds", h, min, sec)
	}

	// info (or just $configure / $configure info / $configure i)
	if sub == "" || sub == "info" || sub == "i" {
		serverCount := len(s.State.Guilds)
		ping := math.Round(float64(s.HeartbeatLatency().Milliseconds()))

		infoMsg := fmt.Sprintf(`
╔══════════════════════════════════════════════╗
║               RYZE V3 • INFO                 ║
╚══════════════════════════════════════════════╝

Uptime       : %s
Ping         : %.0fms
Prefix       : %s
Servers      : %d
Friends      : %d
Blacklisted  : %d
Version      : 3.0 Final
`, uptime(), ping, p, serverCount, len(storage.FriendsList), len(storage.Blacklist))

		s.ChannelMessageSend(m.ChannelID, infoMsg)
		return
	}

	// restart
	if sub == "restart" {
		s.ChannelMessageSend(m.ChannelID, "Reconnecting...")
		s.Close()
		go func() {
			time.Sleep(3 * time.Second)
			newSession, err := discordgo.New(cfg.Token)
			if err == nil {
				newSession.Open()
				*session = *newSession
			}
		}()
		return
	}

	// prefix
	if sub == "prefix" {
		if len(args) < 2 {
			usage := fmt.Sprintf(`
╔══════════════════════════════════════════════╗
║               RYZE V3 • PREFIX               ║
╚══════════════════════════════════════════════╝

Usage: %sprefix <new_prefix>
Example: %sprefix !
`, p, p)
			s.ChannelMessageSend(m.ChannelID, usage)
			return
		}

		newPrefix := args[1]
		cfg.Prefix = newPrefix

		// Save new prefix to config.json
		configBytes, _ := json.MarshalIndent(cfg, "", "  ")
		ioutil.WriteFile("config.json", configBytes, 0644)

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Prefix changed to `%s`", newPrefix))
		return
	}

	// alias
	if sub == "alias" {
		action := ""
		if len(args) > 1 {
			action = args[1]
		}

		// list aliases (default behavior)
		if action == "" || action == "list" {
			var list string
			if len(storage.Aliases) == 0 {
				list = "None"
			} else {
				for alias, cmd := range storage.Aliases {
					list += fmt.Sprintf("%s → %s\n", alias, cmd)
				}
			}

			aliasMsg := fmt.Sprintf(`
╔══════════════════════════════════════════════╗
║               RYZE V3 • ALIAS                ║
╚══════════════════════════════════════════════╝

%salias add <command> <alias>
%salias remove <alias>
%salias list

Current aliases:
%s
`, p, p, p, list)

			s.ChannelMessageSend(m.ChannelID, aliasMsg)
			return
		}

		// add alias
		if action == "add" {
			if len(args) < 4 {
				s.ChannelMessageSend(m.ChannelID, "Usage: alias add <command> <alias>")
				return
			}
			cmd := args[2]
			alias := args[3]

			if _, exists := storage.Aliases[alias]; exists {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Alias `%s` already exists.", alias))
				return
			}
			storage.Aliases[alias] = cmd
			saveFunc()
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Created alias: %s → %s", alias, cmd))
			return
		}

		// remove alias
		if action == "remove" {
			if len(args) < 3 {
				s.ChannelMessageSend(m.ChannelID, "Usage: alias remove <alias>")
				return
			}
			alias := args[2]
			if _, exists := storage.Aliases[alias]; !exists {
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Alias `%s` not found.", alias))
				return
			}
			delete(storage.Aliases, alias)
			saveFunc()
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Removed alias `%s`", alias))
			return
		}
	}
}
