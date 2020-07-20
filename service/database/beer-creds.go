package database

import (
	"bytes"
	"github.com/boltdb/bolt"
	"strconv"
)

func (c *appdbimpl) BeerCreds(userid int64) (map[int64]int, error) {
	var ret = map[int64]int{}
	var meAsByte = []byte(strconv.FormatInt(userid, 10))
	return ret, c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("beers"))

		c := b.Cursor()

		// Cycle all sub buckets and find my user - not very efficient
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v != nil {
				// Key - this should not happen in the root bucket
				continue
			}
			userb := b.Bucket(k)

			c2 := userb.Cursor()
			found := false
			for k2, v2 := c2.First(); k2 != nil && !found; k2, v2 = c2.Next() {
				if bytes.Compare(meAsByte, k2) == 0 {
					toid, err := strconv.ParseInt(string(k2), 10, 64)
					if err != nil {
						return err
					}
					nbeers, err := strconv.ParseInt(string(v2), 10, 64)
					if err != nil {
						return err
					}
					ret[toid] = int(nbeers)
					found = true
				}
			}
		}

		return nil
	})
}
