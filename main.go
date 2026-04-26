package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-telegram/bot"
)

type Challenge struct {
	Question   string `json:"question"`
	Key        string `json:"key"`
	Answer     string `json:"answer"`
	MessageIDs []int  `json:"message_ids"`
}

var (
	db *badger.DB
)

func main() {
	db, err := openBadgerDB()
	if err != nil {
		panic(fmt.Errorf("error abriendo Badger: %w", err))
	}
	defer db.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(messages),
	}

	key, exists := os.LookupEnv("TELEGRAM_BOT_KEY")
	if !exists {
		panic("`TELEGRAM_BOT_KEY` no ha sido definido en el ambiente")
	}

	b, err := bot.New(key, opts...)
	if err != nil {
		panic(err)
	}

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
