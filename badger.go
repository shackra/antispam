package main

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
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
