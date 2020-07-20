package database

import (
	"github.com/boltdb/bolt"
	"strconv"
)

func (c *appdbimpl) BeerDebts(userid int64) (map[int64]int, error) {
	var ret = map[int64]int{}
	return ret, c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("beers"))

		userb := b.Bucket([]byte(strconv.FormatInt(userid, 10)))
		if userb == nil {
			return nil
		}

		c := userb.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			touser, err := strconv.ParseInt(string(k), 10, 64)
			if err != nil {
				return err
			}
			nbeers, err := strconv.ParseInt(string(v), 10, 32) // limit to 2^32 beers...
			if err != nil {
				return err
			}

			ret[touser] = int(nbeers)
		}

		return nil
	})
}
