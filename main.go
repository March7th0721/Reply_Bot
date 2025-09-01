package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"
)

var keys = make(map[string]string)

const filepath = "keys.json"

var adminID int64

func saveKeys() {
	file, err := os.Create(filepath)
	if err != nil {
		log.Println("Failed to save file:", err)
		return
	}
	defer file.Close()

	encodingJson := json.NewEncoder(file)
	encodingJson.SetIndent("", "  ")
	err = encodingJson.Encode(keys)
	if err != nil {
		log.Println("Failed to save keys:", err)
	}
}

func loadKeys() {
	file, err := os.Open(filepath)
	if err != nil {
		log.Println("Failed to load file:", err)
		keys = make(map[string]string)
		return
	}
	defer file.Close()

	decodingJson := json.NewDecoder(file)
	err = decodingJson.Decode(&keys)
	if err != nil {
		log.Println("Failed to load keys:", err)
		keys = make(map[string]string)
	}
}

func loadAdminID() {
	idStr := os.Getenv("ADMIN_ID")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	adminID = id
}

func main() {
	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	loadKeys()
	loadAdminID()

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/addkey", func(c tele.Context) error {
		if c.Sender().ID != adminID {
			return c.Reply("You don't have permission to use this command.")
		}
		args := c.Args()
		if len(args) < 2 {
			return c.Reply("Usage: /addkey <keyword> <reply>")
		}
		key := args[0]
		reply := strings.Join(args[1:], " ")
		keys[key] = reply
		saveKeys()
		return c.Reply("Keyword added.")
	})

	b.Handle("/delkey", func(c tele.Context) error {
		if c.Sender().ID != adminID {
			return c.Reply("You don't have permission to use this command.")
		}
		args := c.Args()
		if len(args) < 1 {
			return c.Reply("Usage: /delkey <keyword>")
		}
		key := args[0]
		if _, exists := keys[key]; exists {
			delete(keys, key)
			saveKeys()
			return c.Reply("Keyword deleted.")
		}
		return c.Reply("Keyword not found.")
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		text := strings.ToLower(c.Text())
		for key, reply := range keys {
			if strings.Contains(text, strings.ToLower(key)) {
				return c.Reply(reply)
			}
		}
		return nil
	})

	b.Handle("/ping", func(c tele.Context) error {
		return c.Send("pong!")
	})
	b.Start()
}
