package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-telegram/bot"
)

type Challenge struct {
	Question   string    `json:"question"`
	Key        string    `json:"key"`
	Answer     string    `json:"answer"`
	MessageIDs []int     `json:"message_ids"`
	ChatID     int64     `json:"chat_id"`
	UserID     int64     `json:"user_id"`
	UserName   string    `json:"user_name"`
	CreatedAt  time.Time `json:"created_at"`
}

var (
	db *badger.DB
)

func main() {
	var err error
	db, err = openBadgerDB()
	if err != nil {
		panic(fmt.Errorf("error abriendo Badger: %w", err))
	}
	defer db.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(messages),
		bot.WithAllowedUpdates(bot.AllowedUpdates{
			"message", "chat_member",
		}),
	}

	key, exists := os.LookupEnv("TELEGRAM_BOT_KEY")
	if !exists {
		panic("`TELEGRAM_BOT_KEY` no ha sido definido en el ambiente")
	}

	b, err := bot.New(key, opts...)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(2 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := cleanupOldChallenges(ctx, db, b)
				if err != nil {
					log.Printf("no se pudo limpiar los retos, error: %v", err)
				}
			}
		}
	}()

	b.Start(ctx)
}

func openBadgerDB() (*badger.DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener el directorio home: %w", err)
	}

	dbPath := filepath.Join(home, "emacs_es-antispam")

	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	return badger.Open(opts)
}
