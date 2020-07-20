package database

import (
	"github.com/boltdb/bolt"
	"strconv"
)

func (c *appdbimpl) AddOrUpdateChatInfo(chatid int64, name string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("chats"))

		return b.Put([]byte(strconv.FormatInt(chatid, 10)), []byte(name))
	})
}
