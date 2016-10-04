package qubot

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

// DB represents the application-level database.
type DB struct {
	*bolt.DB
}

// Open opens and initializes the database.
func (db *DB) Open(path string, mode os.FileMode) error {
	d, err := bolt.Open(path, mode, nil)
	if err != nil {
		return err
	}
	db.DB = d

	return db.init()
}

// init initializes the top-level buckets.
func (db *DB) init() error {
	return db.Update(func(tx *Tx) error {
		_, _ = tx.CreateBucketIfNotExists([]byte("meta"))
		_, _ = tx.CreateBucketIfNotExists([]byte("users"))

		return nil
	})
}

// View executes a function in the context of a read-only transaction.
func (db *DB) View(fn func(*Tx) error) error {
	return db.DB.View(func(tx *bolt.Tx) error {
		return fn(&Tx{tx})
	})
}

// Update executes a function in the context of a writable transaction.
func (db *DB) Update(fn func(*Tx) error) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		return fn(&Tx{tx})
	})
}

// Tx represents an application-level transaction.
type Tx struct {
	*bolt.Tx
}

func (tx *Tx) meta() *bolt.Bucket  { return tx.Bucket([]byte("meta")) }
func (tx *Tx) users() *bolt.Bucket { return tx.Bucket([]byte("users")) }

// Meta retrieves a meta field by name.
func (tx *Tx) Meta(key string) string {
	return string(tx.meta().Get([]byte(key)))
}

// SetMeta sets the value of a meta field by name.
func (tx *Tx) SetMeta(key, value string) error {
	return tx.meta().Put([]byte(key), []byte(value))
}

// User retrieves an user from the database by ID.
func (tx *Tx) User(id int) (u *User, err error) {
	if v := tx.users().Get(i64tob(int64(id))); v != nil {
		err = json.Unmarshal(v, &u)
	}
	return
}

// SaveUser stores an user in the database.
func (tx *Tx) SaveUser(u *User) error {
	if u == nil {
		panic("nil user")
	}
	if u.ID == 0 {
		panic("user id required")
	}
	b, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("marshal user: %s", err)
	}
	return tx.users().Put(i64tob(int64(u.ID)), b)
}

// Converts an integer to a big-endian encoded byte slice.
func i64tob(v int64) []byte {
	var b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
