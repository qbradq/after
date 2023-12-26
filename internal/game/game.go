package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

var save *badger.DB

func init() {
	os.MkdirAll("saves", 0664)
}

// ChunkGen getter
var GetChunkGen func(string) ChunkGen

// ListSaves returns a list of all save file names.
func ListSaves() []string {
	return nil
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

// openSave blindly opens the named save.
func openSave(si *SaveInfo) error {
	p := path.Join("saves", si.ID)
	db, err := badger.Open(badger.DefaultOptions(p))
	if err != nil {
		return err
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
