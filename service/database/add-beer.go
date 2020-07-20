package database

import (
	"github.com/boltdb/bolt"
	"strconv"
)

func (c *appdbimpl) AddBeer(from int64, to int64) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("beers"))

		userb, err := b.CreateBucketIfNotExists([]byte(strconv.FormatInt(from, 10)))
		if err != nil {
			return err
		}

		var boltKey = []byte(strconv.FormatInt(to, 10))
		var currentValue int64 = 0
		ret := userb.Get(boltKey)
		if ret != nil {
			currentValue, err = strconv.ParseInt(string(ret), 10, 64)
			if err != nil {
				return err
			}
		}

		return userb.Put(boltKey, []byte(strconv.FormatInt(currentValue+1, 10)))
	})
}
