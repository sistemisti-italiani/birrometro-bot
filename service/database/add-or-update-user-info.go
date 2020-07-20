package database

import (
	"github.com/boltdb/bolt"
	"strconv"
	"strings"
)

func (c *appdbimpl) AddOrUpdateUserInfo(userid int64, username string, firstname string, lastname string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))

		if username == "" {
			username = firstname + " " + lastname
		} else {
			username = "@" + username
		}

		return b.Put([]byte(strconv.FormatInt(userid, 10)), []byte(strings.TrimSpace(username)))
	})
}
