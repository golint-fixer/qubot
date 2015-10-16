package qubot

import (
	"log"
	"os"
	"testing"
	"time"

	"testutil"
)

var dbExampleUser = &User{
	ID:       "U100",
	Name:     "foo",
	Email:    "foo@fighters.com",
	Creation: time.Now().UTC(),
}

// Ensure that a database can be opened and closed.
func TestDB_Open(t *testing.T) {
	db := NewTestDB()
	testutil.Ok(t, db.Close())
}

//  Ensue that a meta string can be persisted to the database.
func TestTx_SetMeta(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	key := "foo"
	val := "bar"

	testutil.Ok(t, db.Update(func(tx *Tx) error {
		testutil.Ok(t, tx.SetMeta(key, val))
		return nil
	}))

	testutil.Ok(t, db.View(func(tx *Tx) error {
		meta := tx.Meta(key)
		testutil.Equals(t, meta, val)
		return nil
	}))
}

// Ensure that a user can be persisted to the database.
func TestTx_SaveUser(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	testutil.Ok(t, db.Update(func(tx *Tx) error {
		testutil.Ok(t, tx.SaveUser(dbExampleUser))
		return nil
	}))

	testutil.Ok(t, db.View(func(tx *Tx) error {
		u, _ := tx.User("U100")
		testutil.Equals(t, dbExampleUser, u)
		return nil
	}))
}

// TestDB wraps the DB to provide helper functions and clean up.
type TestDB struct {
	*DB
}

func NewTestDB() *TestDB {
	db := &TestDB{DB: &DB{}}
	if err := db.Open(testutil.Tempfile(), 0600); err != nil {
		log.Fatal("open: ", err)
	}
	return db
}

func (db *TestDB) Close() error {
	defer os.RemoveAll(db.Path())
	return db.DB.Close()
}
