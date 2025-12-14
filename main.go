package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"discord-selfbot/commands"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Token  string `json:"token"`
	Prefix string `json:"prefix"`
}

type Storage struct {
	Counter     int                    `json:"counter"`
	Data        map[string]interface{} `json:"example_data"`
	Aliases     map[string]string      `json:"aliases"`
	Blacklist   []string               `json:"blacklist"`
	FriendsList []string               `json:"friends_list"`
}

var (
	cfg         Config
	storage     Storage
	session     *discordgo.Session
	startTime   = time.Now()
)

func main() {
	loadConfig()
	loadStorage()

	// Allow token/prefix override via env (best for Leapcell)
	if os.Getenv("USER_TOKEN") != "" {
		cfg.Token = os.Getenv("USER_TOKEN")
	}
	if os.Getenv("PREFIX") != "" {
		cfg.Prefix = os.Getenv("PREFIX")
	}

	if cfg.Token == "" {
		log.Fatal("No token found! Set USER_TOKEN env var or put it in config.json")
	}
	if cfg.Prefix == "" {
		cfg.Prefix = "!" // default fallback
	}

	// Create session with plain user token
	var err error
	session, err = discordgo.New(cfg.Token)
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}

	// CRITICAL: DO NOT set any Intents → selfbots get full access automatically
	// DO NOT set any Presence/Status → keeps your real client status untouched (stealthy)

	// Register all commands
	commands.RegisterAll(session, &cfg, &storage, saveStorage)

	// Message handler with command routing + alias support
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore own messages
		if m.Author.ID == s.State.User.ID {
			return
		}

		// Check prefix
		if !strings.HasPrefix(m.Content, cfg.Prefix) {
			return
		}

		content := strings.TrimPrefix(m.Content, cfg.Prefix)
		args := strings.Fields(content)
		if len(args) == 0 {
			return
		}

		cmdName := strings.ToLower(args[0])

		// Check aliases first
		if realCmd, isAlias := storage.Aliases[cmdName]; isAlias {
			cmdName = realCmd
			// Shift args: remove the alias, keep the rest
			args = args[1:]
		} else {
			args = args[1:] // normal command: remove command name
		}

		// Execute command
		commands.Execute(cmdName, s, m, args)
	})

	// Open connection
	err = session.Open()
	if err != nil {
		log.Fatal("Failed to connect: ", err)
	}
	log.Printf("Selfbot running stealthily | Prefix: %s | User: %s", cfg.Prefix, session.State.User.String())

	// Health check for Leapcell (keeps instance alive)
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Wait for shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Save storage on exit
	saveStorage()
	session.Close()
	log.Println("Selfbot stopped safely.")
}

func loadConfig() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Cannot read config.json: ", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatal("Invalid config.json: ", err)
	}
}

func loadStorage() {
	data, err := ioutil.ReadFile("storage.json")
	if err != nil {
		// Create default
		storage = Storage{
			Data:        make(map[string]interface{}),
			Aliases:     make(map[string]string),
			Blacklist:   make([]string, 0),
			FriendsList: make([]string, 0),
		}
		saveStorage()
		return
	}
	json.Unmarshal(data, &storage)

	// Ensure nil safety
	if storage.Data == nil {
		storage.Data = make(map[string]interface{})
	}
	if storage.Aliases == nil {
		storage.Aliases = make(map[string]string)
	}
	if storage.Blacklist == nil {
		storage.Blacklist = make([]string, 0)
	}
	if storage.FriendsList == nil {
		storage.FriendsList = make([]string, 0)
	}
}

func saveStorage() {
	data, _ := json.MarshalIndent(storage, "", "  ")
	ioutil.WriteFile("storage.json", data, 0644)
}
