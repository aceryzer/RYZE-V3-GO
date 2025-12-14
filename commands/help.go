package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func HelpRun(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	prefix := cfg.Prefix

	helpMenu := fmt.Sprintf(```
╔══════════════════════════════════╗
║      RYZE V3 - S3LFB0T           ║
╚══════════════════════════════════╝

┌─ CONFIGURE ──────────────────────┐
│ %sinfo      - Bot information     │
│ %sprefix    - Change prefix       │
│ %srestart   - Reconnect bot       │
│ %salias     - Command aliases     │
└──────────────────────────────────┘

┌─ COMMANDS ───────────────────────┐
│ SOON.                            │
└──────────────────────────────────┘
```, prefix, prefix, prefix, prefix)

	s.ChannelMessageSend(m.ChannelID, helpMenu)

	// Delete reply after 2 minutes, then delete original message
	go func() {
		time.Sleep(120 * time.Second)
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	}()
	s.ChannelMessageDelete(m.ChannelID, m.ID)
}
