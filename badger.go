package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-telegram/bot"
)

func saveChallenge(userID int64, ch Challenge) error {
	key := fmt.Appendf(nil, "challenge:%d", userID)
	data, err := json.Marshal(ch)
	if err != nil {
		return err
	}

	return db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

func getChallenge(userID int64) (Challenge, bool) {
	key := fmt.Appendf(nil, "challenge:%d", userID)
	var ch Challenge

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &ch)
		})
	})

	if err != nil {
		return Challenge{}, false
	}
	return ch, true
}

func deleteChallenge(userID int64) error {
	key := fmt.Appendf(nil, "challenge:%d", userID)
	return db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// CleanupChallenges elimina retos que tengan más de 60 segundos
func cleanupOldChallenges(ctx context.Context, db *badger.DB, b *bot.Bot) error {
	const timeout = 60 * time.Second

	return db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		var toDelete []int64

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil) // copia segura

			// Solo procesamos nuestras keys de retos
			if !bytes.HasPrefix(key, []byte("challenge:")) {
				continue
			}

			// Obtener el valor para leer el timestamp
			var ch Challenge
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &ch)
			})
			if err != nil {
				continue
			}

			if time.Since(ch.CreatedAt) > timeout {
				key := bytes.TrimLeft(key, "challenge:")
				id, err := strconv.ParseInt(fmt.Sprintf("%s", string(key[:])), 0, 64)
				if err != nil {
					log.Printf("no se pudo convertir %q a entero", string(key[:]))
					return err
				}
				toDelete = append(toDelete, id)
				log.Printf("🧹 Reto expirado (>%ds): %d\n", int(timeout.Seconds()), id)
			}
		}

		// Eliminar fuera del iterator (mejor práctica)
		for _, key := range toDelete {
			banUser(ctx, b, key, -1)
		}

		return nil
	})
}
