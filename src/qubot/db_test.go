package qubot

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

// tempfile returns the path to a non-existent file in the temp directory.
func tempfile() string {
	f, _ := ioutil.TempFile("", "raft-")
	path := f.Name()
	f.Close()
	os.Remove(path)
	return path
}

// Ensure that a database can be opened and closed.
func TestDB_Open(t *testing.T) {
	db := NewTestDB()
	ok(t, db.Close())
}

//  Ensue that a meta string can be persisted to the database.
func TestTx_SetMeta(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	key := "foo"
	val := "bar"

	ok(t, db.Update(func(tx *Tx) error {
		ok(t, tx.SetMeta(key, val))
		return nil
	}))

	ok(t, db.View(func(tx *Tx) error {
		meta := tx.Meta(key)
		equals(t, meta, val)
		return nil
	}))
}

// Ensure that a user can be persisted to the database.
func TestTx_SaveUser(t *testing.T) {
	db := NewTestDB()
	defer db.Close()

	ok(t, db.Update(func(tx *Tx) error {
		ok(t, tx.SaveUser(&User{ID: 100, Username: "foo", Email: "foo@fighers.com"}))
		return nil
	}))

	ok(t, db.View(func(tx *Tx) error {
		u, _ := tx.User(100)
		equals(t, &User{ID: 100, Username: "foo", Email: "foo@fighers.com"}, u)
		return nil
	}))
}

// TestDB wraps the DB to provide helper functions and clean up.
type TestDB struct {
	*DB
}

func NewTestDB() *TestDB {
	db := &TestDB{DB: &DB{}}
	if err := db.Open(tempfile(), 0600); err != nil {
		log.Fatal("open: ", err)
	}
	return db
}

func (db *TestDB) Close() error {
	defer os.RemoveAll(db.Path())
	return db.DB.Close()
}
