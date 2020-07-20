package database

import (
	"github.com/boltdb/bolt"
	"strconv"
)

func (c *appdbimpl) GetUserName(userid int64) (string, error) {
	var username string
	return username, c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))

		ret := b.Get([]byte(strconv.FormatInt(userid, 10)))
		if ret == nil {
			// Può accadere? return errors.New("username doesn't exists")
			username = "?"
		} else {
			username = string(ret)
		}
		return nil
	})
}
