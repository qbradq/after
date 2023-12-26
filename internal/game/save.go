package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

// Global save database handle
var save *bbolt.DB

// NewSave creates a new save with the given name.
func NewSave(name string, mods []string) error {
	s := uuid.NewString()
	CloseSave()
	p := path.Join("saves", s)
	_, err := os.Stat(p)
	if !os.IsNotExist(err) {
		if err == nil {
			err = fmt.Errorf("save \"%s\" already exists", s)
		}
		return err
	}
	si := &SaveInfo{
		ID:   s,
		Name: name,
		Mods: mods,
	}
	d, err := json.Marshal(si)
	if err != nil {
		return err
	}
	os.WriteFile(p+".json", d, 0664)
	Saves[si.ID] = si
	return openSave(si)
}

// LoadSave loads the named save file.
func LoadSave(id string) error {
	CloseSave()
	save, found := Saves[id]
	if !found {
		return fmt.Errorf("save file %s not found", id)
	}
	p := path.Join("saves", id)
	if _, err := os.Stat(p); err != nil {
		return err
	}
	return openSave(save)
}

// openSave blindly opens the named save.
func openSave(si *SaveInfo) error {
	p := path.Join("saves", si.ID)
	db, err := bbolt.Open(p, 0664, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return err
	}
	if err := db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("After")); err != nil {
			return err
		}
		return nil
	}); err != nil {
		panic(err)
	}
	save = db
	LoadTileRefs()
	return nil
}

// CloseSave closes the save file and should be called!
func CloseSave() {
	if save != nil {
		save.Close()
		save = nil
	}
}

// SaveValue saves the given data to the save database or panics.
func SaveValue(key string, value []byte) {
	if err := save.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("After"))
		err := b.Put([]byte(key), value)
		return err
	}); err != nil {
		panic(err)
	}
}

// LoadValue returns a reader with the requested data or panics.
func LoadValue(key string) io.Reader {
	var buf []byte
	save.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("After"))
		d := b.Get([]byte(key))
		buf = make([]byte, len(d))
		copy(buf, d)
		return nil
	})
	if len(buf) == 0 {
		return nil
	}
	return bytes.NewReader(buf)
}
