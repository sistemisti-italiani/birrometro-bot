package database

import (
	"fmt"
	"github.com/boltdb/bolt"
)

type AppDatabase interface {
	// Prepare DB
	Init() error

	// This function is used to update the chat info
	AddOrUpdateChatInfo(chatid int64, name string) error

	// This function is used to map an user ID to a human-friendly name
	AddOrUpdateUserInfo(userid int64, username string, firstname string, lastname string) error

	// This function is used to retrieve the user human-friendly name
	GetUserName(userid int64) (string, error)

	// Add beer to count
	AddBeer(from int64, to int64) error

	// Remove beer to count (eg. if you made a mistake) - TODO
	//RemoveBeer(from int64, to int64) error

	// Get beer "debs"
	BeerDebts(userid int64) (map[int64]int, error)

	// Get beer "credits" - TODO
	BeerCreds(userid int64) (map[int64]int, error)
}

type appdbimpl struct {
	db *bolt.DB
}

func NewAppDatabase(db *bolt.DB) AppDatabase {
	return &appdbimpl{
		db: db,
	}
}

func (c *appdbimpl) Init() error {
	return c.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("beers"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte("chats"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}
